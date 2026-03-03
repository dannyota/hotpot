package machine

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// NormalizedMachine is the common representation produced by each provider.
type NormalizedMachine struct {
	Provider         string
	IsBase           bool
	BronzeTable      string
	BronzeResourceID string
	Hostname         string
	OSType           string
	OSName           string
	Status           string // running, stopped
	InternalIP       string
	ExternalIP       string
	Environment      string
	CloudProject     string
	CloudZone        string
	CloudMachineType string
	CollectedAt      time.Time
	FirstCollectedAt time.Time
	MergeKeys        map[string][]string
}

// ResourceID returns the deterministic silver resource ID: "{provider}:{bronze_resource_id}".
func (n *NormalizedMachine) ResourceID() string {
	return n.Provider + ":" + n.BronzeResourceID
}

// Provider loads bronze data and normalizes it into NormalizedMachine records.
type Provider interface {
	Key() string
	Label() string
	IsBase() bool
	Load(ctx context.Context, db *sql.DB) ([]NormalizedMachine, error)
}

// NormalizeMAC normalizes a MAC address to uppercase colon-separated format.
// Returns empty string for invalid or all-zero MACs.
func NormalizeMAC(mac string) string {
	mac = strings.TrimSpace(mac)
	if mac == "" {
		return ""
	}

	raw := strings.ReplaceAll(mac, ":", "")
	raw = strings.ReplaceAll(raw, "-", "")
	raw = strings.ReplaceAll(raw, ".", "")
	raw = strings.ToUpper(raw)

	if len(raw) != 12 {
		return ""
	}
	if raw == "000000000000" {
		return ""
	}

	return raw[0:2] + ":" + raw[2:4] + ":" + raw[4:6] + ":" + raw[6:8] + ":" + raw[8:10] + ":" + raw[10:12]
}

// InferEnvironment guesses environment from hostname prefix or S1 site name.
func InferEnvironment(hostname, s1Site string) string {
	if s1Site != "" {
		return s1Site
	}
	lower := strings.ToLower(hostname)
	switch {
	case strings.HasPrefix(lower, "prd-") || strings.HasPrefix(lower, "prod-"):
		return "PRODUCTION"
	case strings.HasPrefix(lower, "uat-"):
		return "UAT"
	case strings.HasPrefix(lower, "drv-") || strings.HasPrefix(lower, "dev-"):
		return "UAT"
	default:
		return ""
	}
}
