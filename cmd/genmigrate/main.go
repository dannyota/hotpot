// Command genmigrate generates Atlas migration SQL files by diffing ent schemas
// against the target database. This is a dev-only tool — production deployments
// use cmd/migrate which embeds the generated SQL files.
//
// Usage:
//
//	genmigrate --schema <dir> --out <dir> [--db <name>] <name>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"hotpot/internal/atlascfg"
	"hotpot/pkg/base/app"
)

// Layer order: bronze first, then history, then silver, gold.
var layerOrder = []string{"bronze", "bronzehistory", "silver", "gold"}

func main() {
	schemaDir := flag.String("schema", "", "ent schema root directory (required)")
	outDir := flag.String("out", "", "migrations output directory (required)")
	dbFlag := flag.String("db", "", "override database name (must end with _dev)")
	flag.Parse()

	name := flag.Arg(0)
	if *schemaDir == "" || *outDir == "" || name == "" {
		log.Fatal("usage: genmigrate --schema <dir> --out <dir> [--db <name>] <name>")
	}

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
	dbName := cfg.DBName
	if *dbFlag != "" {
		dbName = *dbFlag
	}

	// Safety check: database name must end with _dev. Atlas drops and
	// recreates tables in this database during diff, so running against a
	// production database would destroy data.
	if !strings.HasSuffix(dbName, "_dev") {
		log.Fatal("SAFETY CHECK FAILED: database name must end with \"_dev\"!\n" +
			"Atlas will DROP AND RECREATE tables in this database during 'migrate diff'.\n" +
			"This would DESTROY PRODUCTION DATA.\n\n" +
			"Fix: Use --db flag or set dbname in config:\n" +
			"  genmigrate --db hotpot_dev <name>\n" +
			"  # or in config.yaml:\n" +
			"  database:\n" +
			"    dbname: hotpot_dev  # Must end with _dev")
	}

	postgresURL := fmt.Sprintf("postgres://%s", cfg.User)
	if cfg.Password != "" {
		postgresURL += ":" + cfg.Password
	}
	postgresURL += fmt.Sprintf("@%s:%d/%s", cfg.Host, cfg.Port, dbName)
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "require"
	}
	postgresURL += "?sslmode=" + sslmode

	// Change to output directory so relative paths in Atlas config resolve correctly.
	absOut, err := filepath.Abs(*outDir)
	if err != nil {
		log.Fatalf("Failed to resolve output dir: %v", err)
	}
	absSchema, err := filepath.Abs(*schemaDir)
	if err != nil {
		log.Fatalf("Failed to resolve schema dir: %v", err)
	}

	if err := os.Chdir(absOut); err != nil {
		log.Fatalf("Failed to change to output dir %s: %v", absOut, err)
	}

	config := buildDiffConfig(absSchema, postgresURL)

	for _, layer := range layerOrder {
		atlasSchemaDir := filepath.Join(absSchema, layer, "atlas_schema")
		if _, err := os.Stat(atlasSchemaDir); os.IsNotExist(err) {
			continue
		}

		fmt.Printf("==> %s: atlas migrate diff %s\n", layer, name)

		if err := runAtlasDiff(name, layer, config); err != nil {
			log.Fatalf("%s failed: %v", layer, err)
		}

		if err := renameToSequential(layer); err != nil {
			log.Fatalf("%s rename failed: %v", layer, err)
		}
	}

	fmt.Println("\n✅ Migration diff complete")
}

// buildDiffConfig returns an Atlas HCL config with src, dev, url, and migration
// dir for each layer. The same URL is used for both dev and url since genmigrate
// only runs against a _dev database.
func buildDiffConfig(schemaDir, dbURL string) string {
	var b strings.Builder
	for _, layer := range layerOrder {
		fmt.Fprintf(&b, "env %q {\n", layer)
		fmt.Fprintf(&b, "  src = \"ent://%s\"\n", filepath.Join(schemaDir, layer, "atlas_schema"))
		fmt.Fprintf(&b, "  dev = %q\n", dbURL)
		fmt.Fprintf(&b, "  url = %q\n", dbURL)
		fmt.Fprintf(&b, "  migration {\n    dir = \"file://%s\"\n  }\n", layer)
		fmt.Fprintf(&b, "}\n")
	}
	return b.String()
}

// runAtlasDiff executes atlas migrate diff for a layer.
func runAtlasDiff(name, layer, config string) error {
	uri, setupCmd, cleanup, err := atlascfg.ConfigPipe(config)
	if err != nil {
		return fmt.Errorf("config pipe: %w", err)
	}
	defer cleanup()

	cmd := exec.Command("atlas", "migrate", "diff", name, "--config", uri, "--env", layer)
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
