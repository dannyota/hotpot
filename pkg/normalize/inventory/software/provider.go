package software

import (
	"context"
	"database/sql"
	"time"
)

// NormalizedSoftware is the common representation produced by each provider.
type NormalizedSoftware struct {
	Provider         string
	IsBase           bool
	BronzeTable      string
	BronzeResourceID string
	MachineID        string // silver.inventory_machines resource_id
	Name             string // lowercased
	Version          string
	Publisher        string
	InstalledOn      *time.Time
	CollectedAt      time.Time
	FirstCollectedAt time.Time
}

// ResourceID returns the deterministic resource ID: "{provider}:{bronze_resource_id}".
func (n *NormalizedSoftware) ResourceID() string {
	return n.Provider + ":" + n.BronzeResourceID
}

// Provider loads bronze data and normalizes it into NormalizedSoftware records.
type Provider interface {
	Key() string
	Label() string
	IsBase() bool
	Load(ctx context.Context, db *sql.DB) ([]NormalizedSoftware, error)
}
