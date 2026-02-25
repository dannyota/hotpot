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
	Layer      string // e.g. "bronze", "bronzehistory"
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

	discovered := discoverSchemas(schemaRoot)

	// 1. Generate per-provider atlas wrappers (with entsql.Schema annotation)
	for layer, providerSchemas := range discovered.byLayerProvider {
		pgSchema := layerToSchema[layer]
		for provider, schemas := range providerSchemas {
			dir := filepath.Join(target, layer, "atlas_schema", provider)
			os.MkdirAll(dir, 0755)

			if err := generateAtlasWrappers(dir, schemas, pgSchema); err != nil {
				log.Fatalf("generate atlas wrappers for %s/%s: %v", layer, provider, err)
			}
		}
	}

	// 2. Generate per-service ent packages (e.g., ent/gcp/compute/, ent/s1/)
	for serviceKey, schemas := range discovered.byService {
		serviceTarget := filepath.Join("ent", filepath.FromSlash(serviceKey))
		servicePkg := callerModule + "/" + filepath.ToSlash(filepath.Join(rel, "ent", serviceKey))
		packageName := filepath.Base(serviceKey)

		// Generate runtime wrappers
		serviceSchemaDir := filepath.Join(serviceTarget, "schema")
		os.MkdirAll(serviceSchemaDir, 0755)

		if err := generateWrappers(serviceSchemaDir, schemas); err != nil {
			log.Fatalf("generate per-service wrappers for %s: %v", serviceKey, err)
		}

		// Generate schema config
		if err := generateSchemaConfig(serviceTarget, packageName, schemas); err != nil {
			log.Fatalf("generate per-service schema config for %s: %v", serviceKey, err)
		}

		// Run entc
		absServiceSchemaDir, err := filepath.Abs(serviceSchemaDir)
		if err != nil {
			log.Fatalf("get abs path for %s: %v", serviceKey, err)
		}

		if err := entc.Generate(absServiceSchemaDir, &gen.Config{
			Package:  servicePkg,
			Target:   serviceTarget,
			Features: []gen.Feature{gen.FeatureSchemaConfig},
		}); err != nil {
			log.Fatalf("entc generate for %s: %v", serviceKey, err)
		}

		log.Printf("generated per-service ent package: %s (%d types)", serviceKey, len(schemas))
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

// discoverResult holds schemas grouped by layer+provider and by service.
type discoverResult struct {
	byLayerProvider map[string]map[string][]schemaInfo
	byService       map[string][]schemaInfo // key: "gcp/compute", "s1", "do", etc.
}

func discoverSchemas(root string) discoverResult {
	result := discoverResult{
		byLayerProvider: map[string]map[string][]schemaInfo{},
		byService:       map[string][]schemaInfo{},
	}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"mixin"+string(filepath.Separator)) {
			return nil
		}

		dir := filepath.Dir(path)
		relDir, _ := filepath.Rel(root, dir)
		parts := strings.Split(relDir, string(os.PathSeparator))
		layer := parts[0]

		// Provider is parts[1] if present (e.g., "gcp" in "bronze/gcp/compute").
		// Some schemas may be directly under layer (e.g., "bronze/s1/") where parts[1] is the provider.
		var provider string
		if len(parts) >= 2 {
			provider = parts[1]
		}

		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil
		}

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
					si := schemaInfo{
						TypeName:   typeSpec.Name.Name,
						ImportPath: hotpotModule + "/pkg/schema/" + strings.ReplaceAll(relDir, string(os.PathSeparator), "/"),
						Alias:      alias,
						Layer:      layer,
					}
					if provider != "" {
						if result.byLayerProvider[layer] == nil {
							result.byLayerProvider[layer] = map[string][]schemaInfo{}
						}
						result.byLayerProvider[layer][provider] = append(result.byLayerProvider[layer][provider], si)

						// Service key: everything after layer (e.g., "gcp/compute", "s1", "vault/pki")
						serviceKey := strings.Join(parts[1:], "/")
						result.byService[serviceKey] = append(result.byService[serviceKey], si)
					}
				}
			}
		}
		return nil
	})

	return result
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

func generateSchemaConfig(target string, packageName string, schemas []schemaInfo) error {
	var entries strings.Builder
	for _, s := range schemas {
		pgSchema := layerToSchema[s.Layer]
		fmt.Fprintf(&entries, "\t\t%s: %q,\n", s.TypeName, pgSchema)
	}

	src := "// Code generated by entcgen. DO NOT EDIT.\n" +
		"package " + packageName + "\n\n" +
		"// DefaultSchemaConfig returns the schema config mapping each type to its PG schema.\n" +
		"func DefaultSchemaConfig() SchemaConfig {\n" +
		"\treturn SchemaConfig{\n" +
		entries.String() +
		"\t}\n}\n"

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

func generateWrappers(dir string, schemas []schemaInfo) error {
	var imports, types strings.Builder

	seen := map[string]bool{}
	for _, s := range schemas {
		if !seen[s.ImportPath] {
			fmt.Fprintf(&imports, "\t%s %q\n", s.Alias, s.ImportPath)
			seen[s.ImportPath] = true
		}
	}

	for _, s := range schemas {
		fmt.Fprintf(&types, "type %s struct {\n\t%s.%s\n}\n\n", s.TypeName, s.Alias, s.TypeName)
	}

	src := fmt.Sprintf("// Code generated by entcgen. DO NOT EDIT.\npackage schema\n\nimport (\n%s)\n\n%s", imports.String(), types.String())

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format: %w (source: %s)", err, src)
	}
	return os.WriteFile(filepath.Join(dir, "schemas_gen.go"), formatted, 0644)
}
