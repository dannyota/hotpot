package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"hotpot/pkg/base/app"
)

// Layer order: bronze first, then history, then silver, gold
var layerOrder = []string{"bronze", "bronzehistory", "silver", "gold"}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [diff|apply] [name]")
	}
	command := os.Args[1]

	ctx := context.Background()

	// Create app to load config
	application, err := app.New(app.Options{})
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	if err := application.Start(ctx); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	defer application.Stop()

	// Get database URLs from config
	dbURL := application.ConfigService().DatabaseDSN()
	if dbURL == "" {
		log.Fatal("Database DSN not configured")
	}

	devDBURL := application.ConfigService().DevDatabaseDSN()
	if devDBURL == "" {
		log.Fatal("Dev database DSN not configured")
	}

	// Convert DSNs to postgres:// URL format for Atlas
	postgresURL, err := dsnToURL(dbURL)
	if err != nil {
		log.Fatalf("Failed to convert DSN to URL: %v", err)
	}

	devPostgresURL, err := dsnToURL(devDBURL)
	if err != nil {
		log.Fatalf("Failed to convert dev DSN to URL: %v", err)
	}

	// CRITICAL SAFETY CHECK: dev database MUST be different from production
	if postgresURL == devPostgresURL {
		log.Fatal("❌ SAFETY CHECK FAILED: dev database cannot be the same as target database!\n" +
			"Atlas will DROP AND RECREATE tables in the dev database during 'migrate diff'.\n" +
			"This would DESTROY PRODUCTION DATA.\n\n" +
			"Fix: Set dev_dbname to a different database in your config:\n" +
			"  database:\n" +
			"    dbname: hotpot\n" +
			"    dev_dbname: hotpot_dev  # Must be different!")
	}

	// Set database URLs as environment variables for Atlas
	// This prevents credentials from appearing in `ps aux` output
	os.Setenv("HOTPOT_DATABASE_URL", postgresURL)
	os.Setenv("HOTPOT_DEV_DATABASE_URL", devPostgresURL)

	// Change to deploy/migrations directory (where atlas.hcl is located)
	if err := os.Chdir("deploy/migrations"); err != nil {
		log.Fatalf("Failed to change to deploy/migrations directory: %v", err)
	}

	// Run atlas commands for each layer
	for _, layer := range layerOrder {
		atlasSchemaDir := filepath.Join("..", "..", "pkg", "storage", "ent", layer, "atlas_schema")
		if _, err := os.Stat(atlasSchemaDir); os.IsNotExist(err) {
			continue // skip layers with no schemas yet
		}

		fmt.Printf("==> %s: atlas migrate %s --env %s\n", layer, command, layer)

		if err := runAtlasCommand(command, layer); err != nil {
			log.Fatalf("%s failed: %v", layer, err)
		}
	}

	fmt.Println("\n✅ Migration complete")
}

// runAtlasCommand executes atlas command for a layer
func runAtlasCommand(command, layer string) error {
	var args []string

	switch command {
	case "diff":
		name := "auto"
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		args = []string{"migrate", "diff", name, "--env", layer}
	case "apply":
		// No --url flag needed - atlas reads from env var via atlas.hcl
		args = []string{"migrate", "apply", "--env", layer}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	cmd := exec.Command("atlas", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ() // Inherits HOTPOT_DATABASE_URL

	return cmd.Run()
}

// dsnToURL converts "host=x port=y user=z password=w dbname=d sslmode=s" to "postgres://z:w@x:y/d?sslmode=s"
func dsnToURL(dsn string) (string, error) {
	// Parse DSN (simple key=value parser)
	params := make(map[string]string)
	var key, value string
	inValue := false

	for i := 0; i < len(dsn); i++ {
		ch := dsn[i]
		if ch == '=' {
			inValue = true
		} else if ch == ' ' && inValue {
			params[key] = value
			key = ""
			value = ""
			inValue = false
		} else if inValue {
			value += string(ch)
		} else {
			key += string(ch)
		}
	}
	// Last pair
	if key != "" && value != "" {
		params[key] = value
	}

	host := params["host"]
	port := params["port"]
	user := params["user"]
	password := params["password"]
	dbname := params["dbname"]
	sslmode := params["sslmode"]

	if host == "" || port == "" || user == "" || dbname == "" {
		return "", fmt.Errorf("incomplete DSN: missing host, port, user, or dbname")
	}

	url := fmt.Sprintf("postgres://%s", user)
	if password != "" {
		url += ":" + password
	}
	url += fmt.Sprintf("@%s:%s/%s", host, port, dbname)
	if sslmode != "" {
		url += "?sslmode=" + sslmode
	}

	return url, nil
}
