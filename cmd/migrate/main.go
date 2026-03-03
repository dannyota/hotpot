package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dannyota/hotpot/deploy/migrations"
	"github.com/dannyota/hotpot/internal/atlascfg"
	"github.com/dannyota/hotpot/pkg/base/app"
	"github.com/dannyota/hotpot/pkg/migrate"
)

// Bronze providers.
var _ = migrate.ProviderSet("gcp", "greennode", "jenkins", "meec", "s1", "vault", "reference")

// Silver providers.
var _ = migrate.ProviderSet("machine")

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

	postgresURL := atlascfg.PostgresURL(cfg)

	// Extract embedded migrations to a temp dir so Atlas can read them.
	tmpDir, err := os.MkdirTemp("", "hotpot-migrations-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := extractMigrations(tmpDir); err != nil {
		log.Fatalf("Failed to extract migrations: %v", err)
	}

	// Discover layer/provider pairs from the extracted directory structure.
	pairs, err := discoverPairs(tmpDir)
	if err != nil {
		log.Fatalf("Failed to discover migrations: %v", err)
	}

	// Filter to build-time provider set.
	allowed := map[string]bool{}
	for _, p := range migrate.Providers() {
		allowed[p] = true
	}
	var filtered []layerProvider
	for _, lp := range pairs {
		if allowed[lp.provider] {
			filtered = append(filtered, lp)
		}
	}
	pairs = filtered

	if len(pairs) == 0 {
		log.Fatal("No migration directories found for the requested providers")
	}

	for _, lp := range pairs {
		envName := atlascfg.EnvName(lp.layer, lp.provider)
		fmt.Printf("==> %s/%s: atlas migrate apply\n", lp.layer, lp.provider)

		config := buildApplyConfig(envName, lp.layer, lp.provider, postgresURL, filepath.Join(tmpDir, lp.layer, lp.provider))

		if err := atlascfg.RunAtlas(config, envName, "migrate", "apply", "--allow-dirty"); err != nil {
			log.Fatalf("%s/%s failed: %v", lp.layer, lp.provider, err)
		}
	}

	fmt.Println("\n✅ Migration complete")
}

type layerProvider struct {
	layer    string
	provider string
}

// discoverPairs returns all layer/provider pairs found in the extracted
// migration directory, sorted in layer-first order (all bronze/* before
// all bronzehistory/*, etc.).
func discoverPairs(dir string) ([]layerProvider, error) {
	var pairs []layerProvider
	for _, layer := range atlascfg.LayerOrder {
		layerDir := filepath.Join(dir, layer)
		entries, err := os.ReadDir(layerDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		// Sort providers alphabetically for deterministic ordering.
		var providers []string
		for _, e := range entries {
			if e.IsDir() {
				providers = append(providers, e.Name())
			}
		}
		sort.Strings(providers)
		for _, p := range providers {
			pairs = append(pairs, layerProvider{layer: layer, provider: p})
		}
	}
	return pairs, nil
}

// buildApplyConfig returns a minimal Atlas HCL config for applying migrations.
// Each layer/provider pair gets its own revisions_schema so that identically
// named migration files (e.g. 0001_initial.sql) don't collide across providers.
// The caller passes --allow-dirty to atlas since multiple independent migration
// sets share the same database.
func buildApplyConfig(envName, layer, provider, dbURL, migDir string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "env %q {\n", envName)
	fmt.Fprintf(&b, "  url = %q\n", dbURL)
	fmt.Fprintf(&b, "  migration {\n    dir = \"file://%s\"\n", migDir)
	fmt.Fprintf(&b, "    revisions_schema = %q\n", atlascfg.RevisionsSchema(layer, provider))
	fmt.Fprintf(&b, "  }\n")
	fmt.Fprintf(&b, "}\n")
	return b.String()
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
