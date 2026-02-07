package targetvpngateway

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertTargetVpnGateway converts a GCP API TargetVpnGateway to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertTargetVpnGateway(t *computepb.TargetVpnGateway, projectID string, collectedAt time.Time) (bronze.GCPComputeTargetVpnGateway, error) {
	gw := bronze.GCPComputeTargetVpnGateway{
		ResourceID:        fmt.Sprintf("%d", t.GetId()),
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
	var err error
	if t.ForwardingRules != nil {
		gw.ForwardingRulesJSON, err = json.Marshal(t.ForwardingRules)
		if err != nil {
			return bronze.GCPComputeTargetVpnGateway{}, fmt.Errorf("failed to marshal forwarding rules JSON for target vpn gateway %s: %w", t.GetName(), err)
		}
	}
	if t.Tunnels != nil {
		gw.TunnelsJSON, err = json.Marshal(t.Tunnels)
		if err != nil {
			return bronze.GCPComputeTargetVpnGateway{}, fmt.Errorf("failed to marshal tunnels JSON for target vpn gateway %s: %w", t.GetName(), err)
		}
	}

	// Convert labels to separate table
	gw.Labels = ConvertLabels(t.Labels)

	return gw, nil
}

// ConvertLabels converts target VPN gateway labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeTargetVpnGatewayLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeTargetVpnGatewayLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeTargetVpnGatewayLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}
