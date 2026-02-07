package vpntunnel

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertVpnTunnel converts a GCP API VpnTunnel to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertVpnTunnel(t *computepb.VpnTunnel, projectID string, collectedAt time.Time) (bronze.GCPComputeVpnTunnel, error) {
	tunnel := bronze.GCPComputeVpnTunnel{
		ResourceID:                   fmt.Sprintf("%d", t.GetId()),
		Name:                         t.GetName(),
		Description:                  t.GetDescription(),
		Status:                       t.GetStatus(),
		DetailedStatus:               t.GetDetailedStatus(),
		Region:                       t.GetRegion(),
		SelfLink:                     t.GetSelfLink(),
		CreationTimestamp:            t.GetCreationTimestamp(),
		LabelFingerprint:             t.GetLabelFingerprint(),
		IkeVersion:                   t.GetIkeVersion(),
		PeerIp:                       t.GetPeerIp(),
		PeerExternalGateway:          t.GetPeerExternalGateway(),
		PeerExternalGatewayInterface: t.GetPeerExternalGatewayInterface(),
		PeerGcpGateway:               t.GetPeerGcpGateway(),
		Router:                       t.GetRouter(),
		SharedSecretHash:             t.GetSharedSecretHash(),
		VpnGateway:                   t.GetVpnGateway(),
		TargetVpnGateway:             t.GetTargetVpnGateway(),
		VpnGatewayInterface:          t.GetVpnGatewayInterface(),
		ProjectID:                    projectID,
		CollectedAt:                  collectedAt,
	}

	// Convert JSONB fields (nil -> SQL NULL, data -> JSON bytes)
	var err error
	if t.LocalTrafficSelector != nil {
		tunnel.LocalTrafficSelectorJSON, err = json.Marshal(t.LocalTrafficSelector)
		if err != nil {
			return bronze.GCPComputeVpnTunnel{}, fmt.Errorf("failed to marshal JSON for vpn tunnel %s: %w", t.GetName(), err)
		}
	}
	if t.RemoteTrafficSelector != nil {
		tunnel.RemoteTrafficSelectorJSON, err = json.Marshal(t.RemoteTrafficSelector)
		if err != nil {
			return bronze.GCPComputeVpnTunnel{}, fmt.Errorf("failed to marshal JSON for vpn tunnel %s: %w", t.GetName(), err)
		}
	}

	// Convert labels to separate table
	tunnel.Labels = ConvertLabels(t.Labels)

	return tunnel, nil
}

// ConvertLabels converts VPN tunnel labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeVpnTunnelLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeVpnTunnelLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeVpnTunnelLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}
