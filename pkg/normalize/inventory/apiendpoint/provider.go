package apiendpoint

import (
	"context"
	"database/sql"
	"time"
)

// NormalizedApiEndpoint is the common representation produced by each provider.
type NormalizedApiEndpoint struct {
	BronzeResourceID string
	Name             string
	Service          string
	URIPattern       string
	Methods          []string
	IsActive         bool
	AccessLevel      string
	Provider         string
	BronzeTable      string
	CollectedAt      time.Time
	FirstCollectedAt time.Time
}

// Provider loads bronze data and normalizes it into NormalizedApiEndpoint records.
type Provider interface {
	Key() string
	Load(ctx context.Context, db *sql.DB) ([]NormalizedApiEndpoint, error)
}
