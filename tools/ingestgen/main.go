// ingestgen generates blank import files for the ingest binary based on
// ProviderSet() and DisableServiceSet() declarations in the calling package.
//
// Usage: go generate (from a cmd/ingest* directory containing build.go)
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const hotpotModule = "github.com/dannyota/hotpot"

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("get cwd: %v", err)
	}

	// Find the hotpot module root (walk up to find go.mod with our module).
	modRoot := findModuleRoot(cwd)
	ingestDir := filepath.Join(modRoot, "pkg", "ingest")

	// Parse all non-generated .go files in cwd for declarations.
	providers, disabledServices := parseDeclarations(cwd)
	if len(providers) == 0 {
		log.Fatal("no ingest.ProviderSet() call found — every ingest binary must declare providers")
	}

	sort.Strings(providers)
	log.Printf("ingestgen: providers: %v", providers)
	if len(disabledServices) > 0 {
		log.Printf("ingestgen: disabled services: %v", disabledServices)
	}

	// Validate providers exist and discover services.
	type providerInfo struct {
		name     string
		services []string // nil if no service subdirs
	}

	var infos []providerInfo
	for _, name := range providers {
		providerDir := filepath.Join(ingestDir, name)
		if _, err := os.Stat(filepath.Join(providerDir, "provider.go")); err != nil {
			log.Fatalf("provider %q: missing provider.go in %s", name, providerDir)
		}

		services := discoverServices(providerDir)
		if len(services) > 0 {
			// Filter out disabled services.
			disabled := disabledServices[name]
			var enabled []string
			for _, svc := range services {
				if !slices.Contains(disabled, svc) {
					enabled = append(enabled, svc)
				}
			}
			services = enabled
			sort.Strings(services)
			log.Printf("ingestgen: %s services: %v", name, services)
		}

		infos = append(infos, providerInfo{name: name, services: services})
	}

	// Generate providers_gen.go.
	var providerImports []string
	for _, info := range infos {
		providerImports = append(providerImports, hotpotModule+"/pkg/ingest/"+info.name)
	}
	writeGenFile(filepath.Join(cwd, "providers_gen.go"), providerImports)

	// Generate {name}_services_gen.go for each provider that has services.
	for _, info := range infos {
		genFile := filepath.Join(cwd, info.name+"_services_gen.go")
		if len(info.services) == 0 {
			// Clean up stale file if provider has no services.
			os.Remove(genFile)
			continue
		}
		var serviceImports []string
		for _, svc := range info.services {
			serviceImports = append(serviceImports, hotpotModule+"/pkg/ingest/"+info.name+"/"+svc)
		}
		writeGenFile(genFile, serviceImports)
	}

	// Clean up stale *_services_gen.go files for providers no longer listed.
	cleanStaleFiles(cwd, providers)
}

// parseDeclarations parses Go files in dir (excluding *_gen.go) and extracts
// provider names from ingest.ProviderSet(...) calls and disabled services from
// ingest.DisableServiceSet(provider, ...) calls.
func parseDeclarations(dir string) (providers []string, disabledServices map[string][]string) {
	disabledServices = map[string][]string{}

	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("read dir %s: %v", dir, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_gen.go") {
			continue
		}

		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			log.Printf("ingestgen: warning: cannot parse %s: %v", name, err)
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			funcName := selectorName(call.Fun)
			switch funcName {
			case "ingest.ProviderSet":
				providers = append(providers, extractStringArgs(call)...)
			case "ingest.DisableServiceSet":
				args := extractStringArgs(call)
				if len(args) >= 2 {
					provider := args[0]
					disabledServices[provider] = append(disabledServices[provider], args[1:]...)
				}
			}
			return true
		})
	}

	return providers, disabledServices
}

// discoverServices scans a provider directory for subdirectories containing
// register.go (service packages). Skips directories containing provider.go
// (the provider package itself).
func discoverServices(providerDir string) []string {
	entries, err := os.ReadDir(providerDir)
	if err != nil {
		return nil
	}

	var services []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subdir := filepath.Join(providerDir, entry.Name())
		// Must have register.go (service package) but NOT provider.go (provider package).
		if _, err := os.Stat(filepath.Join(subdir, "register.go")); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(subdir, "provider.go")); err == nil {
			continue // this is the provider package, not a service
		}
		services = append(services, entry.Name())
	}
	return services
}

// writeGenFile writes a generated file with blank imports.
func writeGenFile(path string, imports []string) {
	var buf strings.Builder
	buf.WriteString("// Code generated by ingestgen. DO NOT EDIT.\n\npackage main\n\nimport (\n")
	for _, imp := range imports {
		fmt.Fprintf(&buf, "\t_ %q\n", imp)
	}
	buf.WriteString(")\n")

	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		log.Fatalf("format %s: %v", path, err)
	}

	if err := os.WriteFile(path, formatted, 0644); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}
	log.Printf("ingestgen: wrote %s", filepath.Base(path))
}

// cleanStaleFiles removes *_services_gen.go files for providers that are no
// longer in the declared list.
func cleanStaleFiles(dir string, providers []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, "_services_gen.go") {
			continue
		}
		provider := strings.TrimSuffix(name, "_services_gen.go")
		if !slices.Contains(providers, provider) {
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				log.Printf("ingestgen: warning: cannot remove stale %s: %v", name, err)
			} else {
				log.Printf("ingestgen: removed stale %s", name)
			}
		}
	}
}

// selectorName returns "pkg.Func" for a selector expression, or "" otherwise.
func selectorName(expr ast.Expr) string {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name + "." + sel.Sel.Name
}

// extractStringArgs returns all string literal arguments from a call expression.
func extractStringArgs(call *ast.CallExpr) []string {
	var args []string
	for _, arg := range call.Args {
		lit, ok := arg.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		// Remove quotes from string literal.
		s := lit.Value
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		args = append(args, s)
	}
	return args
}

// findModuleRoot walks up from dir looking for a go.mod that declares the hotpot module.
func findModuleRoot(dir string) string {
	abs, _ := filepath.Abs(dir)
	for abs != "/" {
		modFile := filepath.Join(abs, "go.mod")
		data, err := os.ReadFile(modFile)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					mod := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					if mod == hotpotModule {
						return abs
					}
				}
			}
		}
		abs = filepath.Dir(abs)
	}
	log.Fatalf("cannot find go.mod for %s", hotpotModule)
	return ""
}
