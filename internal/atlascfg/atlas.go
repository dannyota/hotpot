package atlascfg

import (
	"fmt"
	"os"
	"os/exec"

	"danny.vn/hotpot/pkg/base/config"
)

// LayerOrder defines the order in which migration layers are processed.
// Bronze tables must exist before bronze_history tables can reference them.
var LayerOrder = []string{"config", "bronze", "bronzehistory", "silver", "gold"}

// EnvName returns the Atlas environment name for a layer/provider pair.
func EnvName(layer, provider string) string {
	return layer + "-" + provider
}

// RevisionsSchema returns the Postgres schema name used to store Atlas
// migration revision history for a layer/provider pair.
func RevisionsSchema(layer, provider string) string {
	return "atlas_" + layer + "_" + provider
}

// layerPGSchema maps layer names to their Postgres schema names.
var layerPGSchema = map[string]string{
	"config":        "config",
	"bronze":        "bronze",
	"bronzehistory": "bronze_history",
	"silver":        "silver",
	"gold":          "gold",
}

// PGSchema returns the Postgres schema name for a layer.
func PGSchema(layer string) string {
	if s, ok := layerPGSchema[layer]; ok {
		return s
	}
	return layer
}

// PostgresURL builds a postgres connection URL from a DatabaseConfig.
func PostgresURL(cfg config.DatabaseConfig) string {
	url := fmt.Sprintf("postgres://%s", cfg.User)
	if cfg.Password != "" {
		url += ":" + cfg.Password
	}
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "require"
	}
	url += fmt.Sprintf("@%s:%d/%s?sslmode=%s", cfg.Host, cfg.Port, cfg.DBName, sslmode)
	return url
}

// RunAtlas executes an atlas command with the given config and environment name.
// The args are passed directly to the atlas CLI (e.g. "migrate", "diff", "name").
func RunAtlas(atlasConfig, envName string, args ...string) error {
	uri, setupCmd, cleanup, err := ConfigPipe(atlasConfig)
	if err != nil {
		return fmt.Errorf("config pipe: %w", err)
	}
	defer cleanup()

	fullArgs := append(args, "--config", uri, "--env", envName)
	cmd := exec.Command("atlas", fullArgs...)
	setupCmd(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Rehash runs atlas migrate hash to update the atlas.sum file for a directory.
func Rehash(dir string) error {
	cmd := exec.Command("atlas", "migrate", "hash", "--dir", fmt.Sprintf("file://%s", dir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rehash: %w", err)
	}
	return nil
}
