package k8snode

import (
	"context"
	"database/sql"
	"time"
)

// NormalizedK8sNode is the common representation produced by each provider.
type NormalizedK8sNode struct {
	Provider         string
	IsBase           bool
	BronzeTable      string
	BronzeResourceID string
	NodeName         string
	ClusterName      string
	NodePool         string
	Status           string
	Provisioning     string
	CloudProject     string
	CloudZone        string
	CloudMachineType string
	InternalIP       string
	ExternalIP       string
	CollectedAt      time.Time
	FirstCollectedAt time.Time
	MergeKeys        map[string][]string
}

// ResourceID returns the deterministic resource ID: "{provider}:{bronze_resource_id}".
func (n *NormalizedK8sNode) ResourceID() string {
	return n.Provider + ":" + n.BronzeResourceID
}

// Provider loads bronze data and normalizes it into NormalizedK8sNode records.
type Provider interface {
	Key() string
	Label() string
	IsBase() bool
	Load(ctx context.Context, db *sql.DB) ([]NormalizedK8sNode, error)
}
