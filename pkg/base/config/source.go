package config

import "context"

// ConfigSource abstracts the configuration backend (Vault or file).
type ConfigSource interface {
	// Load reads all configuration from the source.
	Load(ctx context.Context) (*Config, error)

	// Watch starts watching for changes. Calls onChange when config changes.
	// Returns a stop function to cancel watching.
	Watch(ctx context.Context, onChange func()) (stop func(), err error)

	// Type returns the source type for logging ("vault" or "file").
	Type() string
}
