package dnspolicy

import (
	"encoding/json"
	"fmt"
	"time"

	dnsv1 "google.golang.org/api/dns/v1"
)

// PolicyData holds converted DNS policy data ready for Ent insertion.
type PolicyData struct {
	ID                             string
	Name                           string
	Description                    string
	EnableInboundForwarding        bool
	EnableLogging                  bool
	NetworksJSON                   json.RawMessage
	AlternativeNameServerConfigJSON json.RawMessage
	ProjectID                      string
	CollectedAt                    time.Time
}

// ConvertPolicy converts a GCP API Policy to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertPolicy(policy *dnsv1.Policy, projectID string, collectedAt time.Time) (*PolicyData, error) {
	data := &PolicyData{
		ID:                      fmt.Sprintf("%d", policy.Id),
		Name:                    policy.Name,
		Description:             policy.Description,
		EnableInboundForwarding: policy.EnableInboundForwarding,
		EnableLogging:           policy.EnableLogging,
		ProjectID:               projectID,
		CollectedAt:             collectedAt,
	}

	// Convert networks to JSONB
	if policy.Networks != nil {
		b, err := json.Marshal(policy.Networks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal networks for policy %s: %w", policy.Name, err)
		}
		data.NetworksJSON = b
	}

	// Convert alternative name server config to JSONB
	if policy.AlternativeNameServerConfig != nil {
		b, err := json.Marshal(policy.AlternativeNameServerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal alternative_name_server_config for policy %s: %w", policy.Name, err)
		}
		data.AlternativeNameServerConfigJSON = b
	}

	return data, nil
}
