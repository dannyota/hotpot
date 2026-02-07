package vpngateway

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertVpnGateway converts a GCP API VpnGateway to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertVpnGateway(v *computepb.VpnGateway, projectID string, collectedAt time.Time) (bronze.GCPComputeVpnGateway, error) {
	gw := bronze.GCPComputeVpnGateway{
		ResourceID:        fmt.Sprintf("%d", v.GetId()),
		Name:              v.GetName(),
		Description:       v.GetDescription(),
		Region:            v.GetRegion(),
		Network:           v.GetNetwork(),
		SelfLink:          v.GetSelfLink(),
		CreationTimestamp: v.GetCreationTimestamp(),
		LabelFingerprint:  v.GetLabelFingerprint(),
		GatewayIpVersion:  v.GetGatewayIpVersion(),
		StackType:         v.GetStackType(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Convert JSONB fields (nil -> SQL NULL, data -> JSON bytes)
	if v.VpnInterfaces != nil {
		data, err := json.Marshal(v.VpnInterfaces)
		if err != nil {
			return bronze.GCPComputeVpnGateway{}, fmt.Errorf("failed to marshal JSON for vpn gateway %s: %w", v.GetName(), err)
		}
		gw.VpnInterfacesJSON = data
	}

	// Convert labels to separate table
	gw.Labels = ConvertLabels(v.Labels)

	return gw, nil
}

// ConvertLabels converts VPN gateway labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeVpnGatewayLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeVpnGatewayLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeVpnGatewayLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}
