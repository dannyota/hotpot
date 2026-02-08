package targetvpngateway

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetVpnGatewayData holds converted target VPN gateway data ready for Ent insertion.
type TargetVpnGatewayData struct {
	ID                  string
	Name                string
	Description         string
	Status              string
	Region              string
	Network             string
	SelfLink            string
	CreationTimestamp   string
	LabelFingerprint    string
	ForwardingRulesJSON json.RawMessage
	TunnelsJSON         json.RawMessage
	Labels              []LabelData
	ProjectID           string
	CollectedAt         time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertTargetVpnGateway converts a GCP API TargetVpnGateway to TargetVpnGatewayData.
// Preserves raw API data with minimal transformation.
func ConvertTargetVpnGateway(t *computepb.TargetVpnGateway, projectID string, collectedAt time.Time) (*TargetVpnGatewayData, error) {
	gw := &TargetVpnGatewayData{
		ID:                fmt.Sprintf("%d", t.GetId()),
		Name:              t.GetName(),
		Description:       t.GetDescription(),
		Status:            t.GetStatus(),
		Region:            t.GetRegion(),
		Network:           t.GetNetwork(),
		SelfLink:          t.GetSelfLink(),
		CreationTimestamp: t.GetCreationTimestamp(),
		LabelFingerprint:  t.GetLabelFingerprint(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Convert JSONB fields (nil -> SQL NULL, data -> JSON bytes)
	if t.ForwardingRules != nil {
		rulesBytes, err := json.Marshal(t.ForwardingRules)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal forwarding rules for target vpn gateway %s: %w", t.GetName(), err)
		}
		gw.ForwardingRulesJSON = rulesBytes
	}

	if t.Tunnels != nil {
		tunnelsBytes, err := json.Marshal(t.Tunnels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tunnels for target vpn gateway %s: %w", t.GetName(), err)
		}
		gw.TunnelsJSON = tunnelsBytes
	}

	// Convert labels to separate table
	gw.Labels = ConvertLabels(t.Labels)

	return gw, nil
}

// ConvertLabels converts target VPN gateway labels from GCP API to label data.
func ConvertLabels(labels map[string]string) []LabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, LabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}
