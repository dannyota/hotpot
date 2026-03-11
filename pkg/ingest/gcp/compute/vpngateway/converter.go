package vpngateway

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// VpnGatewayData holds converted VPN gateway data ready for Ent insertion.
type VpnGatewayData struct {
	ID                string
	Name              string
	Description       string
	Region            string
	Network           string
	SelfLink          string
	CreationTimestamp string
	LabelFingerprint  string
	GatewayIpVersion  string
	StackType         string
	VpnInterfacesJSON json.RawMessage
	Labels            []LabelData
	ProjectID         string
	CollectedAt       time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertVpnGateway converts a GCP API VpnGateway to VpnGatewayData.
// Preserves raw API data with minimal transformation.
func ConvertVpnGateway(v *computepb.VpnGateway, projectID string, collectedAt time.Time) (*VpnGatewayData, error) {
	gw := &VpnGatewayData{
		ID:                fmt.Sprintf("%d", v.GetId()),
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
		interfacesBytes, err := json.Marshal(v.VpnInterfaces)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vpn interfaces for vpn gateway %s: %w", v.GetName(), err)
		}
		gw.VpnInterfacesJSON = interfacesBytes
	}

	// Convert labels to separate table
	gw.Labels = ConvertLabels(v.Labels)

	return gw, nil
}

// ConvertLabels converts VPN gateway labels from GCP API to label data.
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
