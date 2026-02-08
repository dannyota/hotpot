package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

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

	// Change to deploy/migrations directory (where atlas.hcl is located)
	if err := os.Chdir("deploy/migrations"); err != nil {
		log.Fatalf("Failed to change to deploy/migrations directory: %v", err)
	}

	// Build atlas config in memory — passed via stdin to avoid
	// credentials appearing in CLI args (`ps aux`) or on disk
	atlasConfig := buildAtlasConfig(postgresURL, devPostgresURL)

	// Run atlas commands for each layer
	for _, layer := range layerOrder {
		atlasSchemaDir := filepath.Join("..", "..", "pkg", "storage", "ent", layer, "atlas_schema")
		if _, err := os.Stat(atlasSchemaDir); os.IsNotExist(err) {
			continue // skip layers with no schemas yet
		}

		fmt.Printf("==> %s: atlas migrate %s\n", layer, command)

		if err := runAtlasCommand(command, layer, atlasConfig); err != nil {
			log.Fatalf("%s failed: %v", layer, err)
		}

		// After diff: rename timestamp-based file to sequential numbering
		if command == "diff" {
			if err := renameToSequential(layer); err != nil {
				log.Fatalf("%s rename failed: %v", layer, err)
			}
		}
	}

	fmt.Println("\n✅ Migration complete")
}

// buildAtlasConfig returns an in-memory atlas HCL config with URLs embedded.
// This avoids env() (requires non-community Atlas) while keeping credentials
// out of CLI args.
func buildAtlasConfig(dbURL, devURL string) string {
	var b strings.Builder
	for _, layer := range layerOrder {
		fmt.Fprintf(&b, "env %q {\n", layer)
		fmt.Fprintf(&b, "  src = \"ent://../../pkg/storage/ent/%s/atlas_schema\"\n", layer)
		fmt.Fprintf(&b, "  dev = %q\n", devURL)
		fmt.Fprintf(&b, "  url = %q\n", dbURL)
		fmt.Fprintf(&b, "  migration {\n    dir = \"file://%s\"\n  }\n", layer)
		fmt.Fprintf(&b, "}\n")
	}
	return b.String()
}

// runAtlasCommand executes atlas command for a layer, piping the config via a
// platform-specific pipe so credentials never appear in CLI args or on disk.
func runAtlasCommand(command, layer, config string) error {
	uri, setupCmd, cleanup, err := configPipe(config)
	if err != nil {
		return fmt.Errorf("config pipe: %w", err)
	}
	defer cleanup()

	var args []string

	switch command {
	case "diff":
		name := "auto"
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		args = []string{"migrate", "diff", name, "--config", uri, "--env", layer}
	case "apply":
		args = []string{"migrate", "apply", "--config", uri, "--env", layer}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	cmd := exec.Command("atlas", args...)
	setupCmd(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// timestampRe matches Atlas's default timestamp-prefixed migration files (e.g. "20260208154545_initial.sql").
var timestampRe = regexp.MustCompile(`^\d{14}_(.+)\.sql$`)

// seqRe matches sequential migration files (e.g. "0001_initial.sql").
var seqRe = regexp.MustCompile(`^(\d{4})_.+\.sql$`)

// maxSeqAcrossLayers scans all layer directories and returns the highest
// sequential migration number found. This gives us a global counter so
// versions are unique across layers and don't collide in the shared
// atlas_schema_revisions table.
func maxSeqAcrossLayers() int {
	max := 0
	for _, layer := range layerOrder {
		entries, err := os.ReadDir(layer)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if m := seqRe.FindStringSubmatch(e.Name()); m != nil {
				n := 0
				fmt.Sscanf(m[1], "%d", &n)
				if n > max {
					max = n
				}
			}
		}
	}
	return max
}

// renameToSequential renames any timestamp-prefixed .sql files in the layer's
// migration directory to use globally-unique zero-padded sequential numbers
// (0001_, 0002_, …). After renaming it rehashes so atlas.sum stays consistent.
func renameToSequential(layer string) error {
	entries, err := os.ReadDir(layer)
	if err != nil {
		return err
	}

	// Find timestamp-prefixed files that need renaming, sorted by name
	var toRename []string
	for _, e := range entries {
		if timestampRe.MatchString(e.Name()) {
			toRename = append(toRename, e.Name())
		}
	}
	if len(toRename) == 0 {
		return nil
	}
	sort.Strings(toRename)

	seq := maxSeqAcrossLayers()
	for _, old := range toRename {
		seq++
		name := timestampRe.FindStringSubmatch(old)[1] // e.g. "initial"
		newName := fmt.Sprintf("%04d_%s.sql", seq, name)

		oldPath := filepath.Join(layer, old)
		newPath := filepath.Join(layer, newName)
		fmt.Printf("    rename: %s -> %s\n", old, newName)
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("rename %s: %w", old, err)
		}
	}

	// Rehash so atlas.sum matches the new filenames
	cmd := exec.Command("atlas", "migrate", "hash", "--dir", fmt.Sprintf("file://%s", layer))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rehash: %w", err)
	}

	return nil
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
