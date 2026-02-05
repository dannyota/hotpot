package app

import (
	"os"
	"time"

	"hotpot/pkg/base/config"
)

// Options configures the App.
type Options struct {
	// ConfigSource overrides auto-detection. If nil, detects from env vars.
	ConfigSource config.ConfigSource

	// GracePeriod before closing old DB connection after reconnect.
	// Default: 5 seconds
	GracePeriod time.Duration

	// OnDBReconnect callback when database reconnects.
	OnDBReconnect func(oldDSN, newDSN string)
}

// DefaultGracePeriod is the default wait time before closing old connections.
const DefaultGracePeriod = 5 * time.Second

// detectConfigSource auto-detects config source from environment variables.
// CONFIG_SOURCE=vault uses Vault, otherwise uses file source.
func detectConfigSource() config.ConfigSource {
	switch os.Getenv("CONFIG_SOURCE") {
	case "vault":
		return config.NewVaultSource(config.VaultSourceOptions{
			Address:    os.Getenv("VAULT_ADDR"),
			Token:      os.Getenv("VAULT_TOKEN"),
			SecretPath: getEnv("VAULT_SECRET_PATH", "hotpot/config"),
			VerifySSL:  os.Getenv("VAULT_SKIP_VERIFY") != "true",
		})
	default:
		return config.NewFileSource(getEnv("CONFIG_FILE", "config.json"))
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
