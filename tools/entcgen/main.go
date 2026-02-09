package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

const hotpotModule = "github.com/dannyota/hotpot"

type schemaInfo struct {
	TypeName   string // e.g. "BronzeGCPComputeInstance"
	ImportPath string // e.g. "github.com/dannyota/hotpot/pkg/schema/bronze/gcp/compute"
	Alias      string // e.g. "bronze_gcp_compute"
}

// layerToSchema maps directory names to PG schema names.
var layerToSchema = map[string]string{
	"bronze":        "bronze",
	"bronzehistory": "bronze_history",
	"silver":        "silver",
	"gold":          "gold",
}

func main() {
	hotpotDir := findModuleDir(hotpotModule)
	schemaRoot := filepath.Join(hotpotDir, "pkg", "schema")

	callerModule, modDir := readModulePath(".")
	target := "ent"
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("get cwd: %v", err)
	}
	rel, err := filepath.Rel(modDir, cwd)
	if err != nil {
		log.Fatalf("compute relative path: %v", err)
	}
	pkg := callerModule + "/" + filepath.ToSlash(filepath.Join(rel, target))

	allSchemas := discoverSchemas(schemaRoot)

	// 1. Generate runtime wrappers (all layers combined)
	runtimeSchemaDir := filepath.Join(target, "schema")
	os.MkdirAll(runtimeSchemaDir, 0755)

	if err := generateWrappers(runtimeSchemaDir, allSchemas); err != nil {
		log.Fatalf("generate runtime wrappers: %v", err)
	}

	// 2. Generate per-layer wrappers for Atlas (with entsql.Schema annotation)
	for layer, schemas := range allSchemas {
		atlasSchemaDir := filepath.Join(target, layer, "atlas_schema")
		os.MkdirAll(atlasSchemaDir, 0755)

		pgSchema := layerToSchema[layer]
		if err := generateAtlasWrappers(atlasSchemaDir, schemas, pgSchema); err != nil {
			log.Fatalf("generate atlas wrappers for %s: %v", layer, err)
		}
	}

	// 3. Generate schema config helper (maps types to PG schemas)
	if err := generateSchemaConfig(target, allSchemas); err != nil {
		log.Fatalf("generate schema config: %v", err)
	}

	// 4. Run entc ONCE — generates ONE client with all types
	absRuntimeSchemaDir, err := filepath.Abs(runtimeSchemaDir)
	if err != nil {
		log.Fatalf("get abs path: %v", err)
	}

	if err := entc.Generate(absRuntimeSchemaDir, &gen.Config{
		Package:  pkg,
		Target:   target,
		Features: []gen.Feature{gen.FeatureSchemaConfig},
	}); err != nil {
		log.Fatalf("entc generate: %v", err)
	}
}

// findModuleDir returns the local directory for a Go module by running
// `go list -m -json`. Works for both the main module (local checkout)
// and dependencies (module cache).
func findModuleDir(module string) string {
	cmd := exec.Command("go", "list", "-m", "-json", module)
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("go list -m -json %s: %v", module, err)
	}

	var info struct {
		Dir string `json:"Dir"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		log.Fatalf("parse go list output: %v", err)
	}
	if info.Dir == "" {
		log.Fatalf("module %s has no local directory (not downloaded?)", module)
	}
	return info.Dir
}

// readModulePath reads the module path and directory from the go.mod
// in the given directory (or a parent).
func readModulePath(dir string) (modulePath, modDir string) {
	abs, _ := filepath.Abs(dir)
	searchDir := abs

	f, err := os.Open(filepath.Join(searchDir, "go.mod"))
	if err != nil {
		// Walk up to find go.mod
		for searchDir != "/" {
			searchDir = filepath.Dir(searchDir)
			f, err = os.Open(filepath.Join(searchDir, "go.mod"))
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Fatal("cannot find go.mod")
		}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), searchDir
		}
	}
	log.Fatal("no module directive in go.mod")
	return "", ""
}

func discoverSchemas(root string) map[string][]schemaInfo {
	layers := map[string][]schemaInfo{}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"mixin"+string(filepath.Separator)) {
			return nil
		}

		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil
		}

		dir := filepath.Dir(path)
		relDir, _ := filepath.Rel(root, dir)
		parts := strings.Split(relDir, string(os.PathSeparator))
		layer := parts[0]

		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || !typeSpec.Name.IsExported() {
					continue
				}
				if embedsEntSchema(typeSpec) {
					alias := strings.ReplaceAll(relDir, string(os.PathSeparator), "_")
					layers[layer] = append(layers[layer], schemaInfo{
						TypeName:   typeSpec.Name.Name,
						ImportPath: hotpotModule + "/pkg/schema/" + strings.ReplaceAll(relDir, string(os.PathSeparator), "/"),
						Alias:      alias,
					})
				}
			}
		}
		return nil
	})

	return layers
}

func embedsEntSchema(ts *ast.TypeSpec) bool {
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return false
	}
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 { // embedded field
			if sel, ok := f.Type.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "ent" && sel.Sel.Name == "Schema" {
						return true
					}
				}
			}
		}
	}
	return false
}

func generateSchemaConfig(target string, schemasByLayer map[string][]schemaInfo) error {
	var entries strings.Builder
	for layer, schemas := range schemasByLayer {
		pgSchema := layerToSchema[layer]
		for _, s := range schemas {
			fmt.Fprintf(&entries, "\t\t%s: %q,\n", s.TypeName, pgSchema)
		}
	}

	src := fmt.Sprintf(`// Code generated by entcgen. DO NOT EDIT.
package ent

// DefaultSchemaConfig returns the schema config mapping each type to its PG schema.
func DefaultSchemaConfig() SchemaConfig {
	return SchemaConfig{
%s	}
}
`, entries.String())

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format: %w (source: %s)", err, src)
	}
	return os.WriteFile(filepath.Join(target, "schema_config_gen.go"), formatted, 0644)
}

func generateAtlasWrappers(dir string, schemas []schemaInfo, pgSchema string) error {
	var imports, types strings.Builder

	seen := map[string]bool{}
	for _, s := range schemas {
		if !seen[s.ImportPath] {
			fmt.Fprintf(&imports, "\t%s %q\n", s.Alias, s.ImportPath)
			seen[s.ImportPath] = true
		}
	}

	for _, s := range schemas {
		fmt.Fprintf(&types, `type %s struct {
	%s.%s
}

func (%s) Annotations() []schema.Annotation {
	anns := %s.%s{}.Annotations()
	for i, a := range anns {
		if v, ok := a.(entsql.Annotation); ok {
			v.Schema = %q
			anns[i] = v
			return anns
		}
	}
	return append(anns, entsql.Annotation{Schema: %q})
}

`, s.TypeName, s.Alias, s.TypeName,
			s.TypeName,
			s.Alias, s.TypeName,
			pgSchema,
			pgSchema,
		)
	}

	src := fmt.Sprintf(`// Code generated by entcgen. DO NOT EDIT.
package schema

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
%s)

%s`, imports.String(), types.String())

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format: %w (source: %s)", err, src)
	}
	return os.WriteFile(filepath.Join(dir, "schemas_gen.go"), formatted, 0644)
}

func generateWrappers(dir string, schemasByLayer map[string][]schemaInfo) error {
	var imports, types strings.Builder

	seen := map[string]bool{}
	for _, schemas := range schemasByLayer {
		for _, s := range schemas {
			if !seen[s.ImportPath] {
				fmt.Fprintf(&imports, "\t%s %q\n", s.Alias, s.ImportPath)
				seen[s.ImportPath] = true
			}
		}
	}

	for _, schemas := range schemasByLayer {
		for _, s := range schemas {
			fmt.Fprintf(&types, "type %s struct {\n\t%s.%s\n}\n\n", s.TypeName, s.Alias, s.TypeName)
		}
	}

	src := fmt.Sprintf("// Code generated by entcgen. DO NOT EDIT.\npackage schema\n\nimport (\n%s)\n\n%s", imports.String(), types.String())

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format: %w (source: %s)", err, src)
	}
	return os.WriteFile(filepath.Join(dir, "schemas_gen.go"), formatted, 0644)
}
