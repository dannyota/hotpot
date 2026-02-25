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

	"github.com/dannyota/hotpot/internal/atlascfg"
	"github.com/dannyota/hotpot/pkg/base/app"
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

	enabledProviders := application.ConfigService().EnabledProviders()
	if len(enabledProviders) == 0 {
		log.Fatal("No providers enabled in config. Set enabled: true for at least one provider.")
	}
	fmt.Printf("Enabled providers: %s\n", strings.Join(enabledProviders, ", "))

	for _, layer := range layerOrder {
		for _, provider := range enabledProviders {
			atlasDir := filepath.Join(absSchema, layer, "atlas_schema", provider)
			if _, err := os.Stat(atlasDir); err != nil {
				continue
			}

			migDir := filepath.Join(layer, provider)
			os.MkdirAll(migDir, 0755)

			envName := layer + "-" + provider
			config := buildDiffConfig(envName, atlasDir, migDir, postgresURL)

			fmt.Printf("==> %s/%s: atlas migrate diff %s\n", layer, provider, name)

			if err := runAtlasDiff(name, envName, config); err != nil {
				log.Fatalf("%s/%s failed: %v", layer, provider, err)
			}

			if err := postProcessSQL(migDir); err != nil {
				log.Fatalf("%s/%s post-process failed: %v", layer, provider, err)
			}

			if err := renameToSequential(migDir); err != nil {
				log.Fatalf("%s/%s rename failed: %v", layer, provider, err)
			}
		}
	}

	fmt.Println("\n✅ Migration diff complete")
}

// buildDiffConfig returns an Atlas HCL config for a single layer/provider pair.
func buildDiffConfig(envName, atlasDir, migDir, dbURL string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "env %q {\n", envName)
	fmt.Fprintf(&b, "  src = \"ent://%s\"\n", atlasDir)
	fmt.Fprintf(&b, "  dev = %q\n", dbURL)
	fmt.Fprintf(&b, "  url = %q\n", dbURL)
	fmt.Fprintf(&b, "  migration {\n    dir = \"file://%s\"\n  }\n", migDir)
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

// runAtlasDiff executes atlas migrate diff for a layer/provider env.
func runAtlasDiff(name, envName, config string) error {
	uri, setupCmd, cleanup, err := atlascfg.ConfigPipe(config)
	if err != nil {
		return fmt.Errorf("config pipe: %w", err)
	}
	defer cleanup()

	cmd := exec.Command("atlas", "migrate", "diff", name, "--config", uri, "--env", envName)
	setupCmd(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// createSchemaRe matches CREATE SCHEMA statements that should use IF NOT EXISTS.
// Multiple providers share the same PG schema (e.g. "bronze"), so each provider's
// migration must be safe to run even if the schema already exists.
var createSchemaRe = regexp.MustCompile(`(?i)CREATE SCHEMA "([^"]+)"`)

// postProcessSQL rewrites CREATE SCHEMA to CREATE SCHEMA IF NOT EXISTS in all
// .sql files in the given directory, then rehashes to keep atlas.sum consistent.
func postProcessSQL(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	modified := false
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		replaced := createSchemaRe.ReplaceAllString(string(data), `CREATE SCHEMA IF NOT EXISTS "$1"`)
		if replaced != string(data) {
			if err := os.WriteFile(path, []byte(replaced), 0644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
			modified = true
		}
	}

	if modified {
		return rehash(dir)
	}
	return nil
}

// timestampRe matches Atlas's default timestamp-prefixed migration files (e.g. "20260208154545_initial.sql").
var timestampRe = regexp.MustCompile(`^\d{14}_(.+)\.sql$`)

// seqRe matches sequential migration files (e.g. "0001_initial.sql").
var seqRe = regexp.MustCompile(`^(\d{4})_.+\.sql$`)

// maxSeqInDir returns the highest sequential migration number in the given directory.
func maxSeqInDir(dir string) int {
	max := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
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
	return max
}

// renameToSequential renames any timestamp-prefixed .sql files in the directory
// to use zero-padded sequential numbers (0001_, 0002_, …), then rehashes.
func renameToSequential(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

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

	seq := maxSeqInDir(dir)
	for _, old := range toRename {
		seq++
		name := timestampRe.FindStringSubmatch(old)[1]
		newName := fmt.Sprintf("%04d_%s.sql", seq, name)

		oldPath := filepath.Join(dir, old)
		newPath := filepath.Join(dir, newName)
		fmt.Printf("    rename: %s -> %s\n", old, newName)
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("rename %s: %w", old, err)
		}
	}

	return rehash(dir)
}

func rehash(dir string) error {
	cmd := exec.Command("atlas", "migrate", "hash", "--dir", fmt.Sprintf("file://%s", dir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rehash: %w", err)
	}
	return nil
}
