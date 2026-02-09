package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dannyota/hotpot/deploy/migrations"
	"github.com/dannyota/hotpot/internal/atlascfg"
	"github.com/dannyota/hotpot/pkg/base/app"
)

func main() {
	ctx := context.Background()

	application, err := app.New(app.Options{})
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	if err := application.Start(ctx); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	defer application.Stop()

	cfg := application.ConfigService().Config().Database
	if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" {
		log.Fatal("Database not configured")
	}

	postgresURL := fmt.Sprintf("postgres://%s", cfg.User)
	if cfg.Password != "" {
		postgresURL += ":" + cfg.Password
	}
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "require"
	}
	postgresURL += fmt.Sprintf("@%s:%d/%s?sslmode=%s", cfg.Host, cfg.Port, cfg.DBName, sslmode)

	// Extract embedded migrations to a temp dir so Atlas can read them.
	tmpDir, err := os.MkdirTemp("", "hotpot-migrations-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := extractMigrations(tmpDir); err != nil {
		log.Fatalf("Failed to extract migrations: %v", err)
	}

	// Discover layer subdirectories from the embedded FS.
	layers, err := layerDirs(tmpDir)
	if err != nil {
		log.Fatalf("Failed to list layers: %v", err)
	}

	for _, layer := range layers {
		fmt.Printf("==> %s: atlas migrate apply\n", layer)

		config := buildApplyConfig(layer, postgresURL, tmpDir)

		if err := runAtlasApply(layer, config); err != nil {
			log.Fatalf("%s failed: %v", layer, err)
		}
	}

	fmt.Println("\n✅ Migration complete")
}

// layerDirs returns the subdirectory names inside dir, sorted alphabetically.
// These correspond to the migration layers (bronze, bronzehistory, etc.).
func layerDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var layers []string
	for _, e := range entries {
		if e.IsDir() {
			layers = append(layers, e.Name())
		}
	}
	return layers, nil
}

// buildApplyConfig returns a minimal Atlas HCL config for applying migrations.
// No src or dev URL is needed — just the target URL and migration directory.
func buildApplyConfig(layer, dbURL, tmpDir string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "env %q {\n", layer)
	fmt.Fprintf(&b, "  url = %q\n", dbURL)
	fmt.Fprintf(&b, "  migration {\n    dir = \"file://%s\"\n  }\n", filepath.Join(tmpDir, layer))
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

// runAtlasApply executes atlas migrate apply for a layer.
func runAtlasApply(layer, config string) error {
	uri, setupCmd, cleanup, err := atlascfg.ConfigPipe(config)
	if err != nil {
		return fmt.Errorf("config pipe: %w", err)
	}
	defer cleanup()

	cmd := exec.Command("atlas", "migrate", "apply", "--config", uri, "--env", layer)
	setupCmd(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// extractMigrations copies the embedded migration FS to a temporary directory.
func extractMigrations(dst string) error {
	return fs.WalkDir(migrations.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dst, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(migrations.FS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
