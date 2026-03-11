package managedzone

import (
	"encoding/json"
	"fmt"
	"time"

	dnsv1 "google.golang.org/api/dns/v1"
)

// ManagedZoneData holds converted managed zone data ready for Ent insertion.
type ManagedZoneData struct {
	ID                          string
	Name                        string
	DnsName                     string
	Description                 string
	Visibility                  string
	CreationTime                string
	DnssecConfigJSON            json.RawMessage
	PrivateVisibilityConfigJSON json.RawMessage
	ForwardingConfigJSON        json.RawMessage
	PeeringConfigJSON           json.RawMessage
	CloudLoggingConfigJSON      json.RawMessage
	Labels                      []LabelData
	ProjectID                   string
	CollectedAt                 time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertManagedZone converts a GCP API ManagedZone to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertManagedZone(zone *dnsv1.ManagedZone, projectID string, collectedAt time.Time) (*ManagedZoneData, error) {
	data := &ManagedZoneData{
		ID:           fmt.Sprintf("%d", zone.Id),
		Name:         zone.Name,
		DnsName:      zone.DnsName,
		Description:  zone.Description,
		Visibility:   zone.Visibility,
		CreationTime: zone.CreationTime,
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}

	// Convert DNSSEC config to JSONB
	if zone.DnssecConfig != nil {
		b, err := json.Marshal(zone.DnssecConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dnssec_config for zone %s: %w", zone.Name, err)
		}
		data.DnssecConfigJSON = b
	}

	// Convert private visibility config to JSONB
	if zone.PrivateVisibilityConfig != nil {
		b, err := json.Marshal(zone.PrivateVisibilityConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private_visibility_config for zone %s: %w", zone.Name, err)
		}
		data.PrivateVisibilityConfigJSON = b
	}

	// Convert forwarding config to JSONB
	if zone.ForwardingConfig != nil {
		b, err := json.Marshal(zone.ForwardingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal forwarding_config for zone %s: %w", zone.Name, err)
		}
		data.ForwardingConfigJSON = b
	}

	// Convert peering config to JSONB
	if zone.PeeringConfig != nil {
		b, err := json.Marshal(zone.PeeringConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal peering_config for zone %s: %w", zone.Name, err)
		}
		data.PeeringConfigJSON = b
	}

	// Convert cloud logging config to JSONB
	if zone.CloudLoggingConfig != nil {
		b, err := json.Marshal(zone.CloudLoggingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cloud_logging_config for zone %s: %w", zone.Name, err)
		}
		data.CloudLoggingConfigJSON = b
	}

	// Convert labels
	data.Labels = ConvertLabels(zone.Labels)

	return data, nil
}

// ConvertLabels converts GCP managed zone labels to label data structs.
func ConvertLabels(labels map[string]string) []LabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(labels))
	for k, v := range labels {
		result = append(result, LabelData{
			Key:   k,
			Value: v,
		})
	}

	return result
}
