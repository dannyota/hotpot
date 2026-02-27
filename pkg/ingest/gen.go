package ingest

// ProviderSet declares which providers to compile into the binary.
// This is a no-op at runtime — read via AST by ingestgen during code generation.
func ProviderSet(names ...string) bool { return true }

// DisableServiceSet declares which services to exclude for a provider.
// This is a no-op at runtime — read via AST by ingestgen during code generation.
func DisableServiceSet(provider string, services ...string) bool { return true }
