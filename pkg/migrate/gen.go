package migrate

var providers []string

// ProviderSet declares which providers to include in the migrate binary.
// Called at package init time from cmd/migrate/main.go.
func ProviderSet(names ...string) bool {
	providers = append(providers, names...)
	return true
}

// Providers returns the registered provider names.
func Providers() []string {
	return providers
}
