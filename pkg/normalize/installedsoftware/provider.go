package installedsoftware

import (
	"context"
	"database/sql"
	"time"
)

// NormalizedInstalledSoftware is the common representation produced by each provider.
type NormalizedInstalledSoftware struct {
	Provider         string
	IsBase           bool
	BronzeTable      string
	BronzeResourceID string
	MachineID        string // silver.machines resource_id
	Name             string // lowercased
	Version          string
	Publisher        string
	CollectedAt      time.Time
	FirstCollectedAt time.Time
}

// ResourceID returns the deterministic silver resource ID: "{provider}:{bronze_resource_id}".
func (n *NormalizedInstalledSoftware) ResourceID() string {
	return n.Provider + ":" + n.BronzeResourceID
}

// Provider loads bronze data and normalizes it into NormalizedInstalledSoftware records.
type Provider interface {
	Key() string
	Label() string
	IsBase() bool
	Load(ctx context.Context, db *sql.DB) ([]NormalizedInstalledSoftware, error)
}
