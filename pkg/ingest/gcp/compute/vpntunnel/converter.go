package vpntunnel

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// VpnTunnelData holds converted VPN tunnel data ready for Ent insertion.
type VpnTunnelData struct {
	ID                           string
	Name                         string
	Description                  string
	Status                       string
	DetailedStatus               string
	Region                       string
	SelfLink                     string
	CreationTimestamp            string
	LabelFingerprint             string
	IkeVersion                   int32
	PeerIp                       string
	PeerExternalGateway          string
	PeerExternalGatewayInterface int32
	PeerGcpGateway               string
	Router                       string
	SharedSecretHash             string
	VpnGateway                   string
	TargetVpnGateway             string
	VpnGatewayInterface          int32
	LocalTrafficSelectorJSON     json.RawMessage
	RemoteTrafficSelectorJSON    json.RawMessage
	Labels                       []LabelData
	ProjectID                    string
	CollectedAt                  time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertVpnTunnel converts a GCP API VpnTunnel to VpnTunnelData.
// Preserves raw API data with minimal transformation.
func ConvertVpnTunnel(t *computepb.VpnTunnel, projectID string, collectedAt time.Time) (*VpnTunnelData, error) {
	tunnel := &VpnTunnelData{
		ID:                           fmt.Sprintf("%d", t.GetId()),
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
	if t.LocalTrafficSelector != nil {
		localBytes, err := json.Marshal(t.LocalTrafficSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal local traffic selector for vpn tunnel %s: %w", t.GetName(), err)
		}
		tunnel.LocalTrafficSelectorJSON = localBytes
	}
	if t.RemoteTrafficSelector != nil {
		remoteBytes, err := json.Marshal(t.RemoteTrafficSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal remote traffic selector for vpn tunnel %s: %w", t.GetName(), err)
		}
		tunnel.RemoteTrafficSelectorJSON = remoteBytes
	}

	// Convert labels to separate table
	tunnel.Labels = ConvertLabels(t.Labels)

	return tunnel, nil
}

// ConvertLabels converts VPN tunnel labels from GCP API to label data.
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
