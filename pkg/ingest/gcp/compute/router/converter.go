package router

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// RouterData holds converted router data ready for Ent insertion.
type RouterData struct {
	ID                         string
	Name                       string
	Description                string
	SelfLink                   string
	CreationTimestamp          string
	Network                    string
	Region                     string
	BgpAsn                     int
	BgpAdvertiseMode           string
	BgpAdvertisedGroupsJSON    json.RawMessage
	BgpAdvertisedIPRangesJSON  json.RawMessage
	BgpKeepaliveInterval       int
	BgpPeersJSON               json.RawMessage
	InterfacesJSON             json.RawMessage
	NatsJSON                   json.RawMessage
	EncryptedInterconnectRouter bool
	ProjectID                  string
	CollectedAt                time.Time
}

// ConvertRouter converts a GCP API Router to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertRouter(r *computepb.Router, projectID string, collectedAt time.Time) (*RouterData, error) {
	data := &RouterData{
		ID:                          fmt.Sprintf("%d", r.GetId()),
		Name:                        r.GetName(),
		Description:                 r.GetDescription(),
		SelfLink:                    r.GetSelfLink(),
		CreationTimestamp:           r.GetCreationTimestamp(),
		Network:                     r.GetNetwork(),
		Region:                      r.GetRegion(),
		EncryptedInterconnectRouter: r.GetEncryptedInterconnectRouter(),
		ProjectID:                   projectID,
		CollectedAt:                 collectedAt,
	}

	// Extract BGP configuration
	if r.Bgp != nil {
		data.BgpAsn = int(r.GetBgp().GetAsn())
		data.BgpAdvertiseMode = r.GetBgp().GetAdvertiseMode()
		data.BgpKeepaliveInterval = int(r.GetBgp().GetKeepaliveInterval())

		if groups := r.GetBgp().GetAdvertisedGroups(); len(groups) > 0 {
			groupsBytes, err := json.Marshal(groups)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal BGP advertised groups for %s: %w", r.GetName(), err)
			}
			data.BgpAdvertisedGroupsJSON = groupsBytes
		}

		if ipRanges := r.GetBgp().GetAdvertisedIpRanges(); len(ipRanges) > 0 {
			ipRangesBytes, err := json.Marshal(ipRanges)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal BGP advertised IP ranges for %s: %w", r.GetName(), err)
			}
			data.BgpAdvertisedIPRangesJSON = ipRangesBytes
		}
	}

	// Marshal BGP peers
	if peers := r.GetBgpPeers(); len(peers) > 0 {
		peersBytes, err := json.Marshal(peers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal BGP peers for %s: %w", r.GetName(), err)
		}
		data.BgpPeersJSON = peersBytes
	}

	// Marshal interfaces
	if ifaces := r.GetInterfaces(); len(ifaces) > 0 {
		ifacesBytes, err := json.Marshal(ifaces)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal interfaces for %s: %w", r.GetName(), err)
		}
		data.InterfacesJSON = ifacesBytes
	}

	// Marshal NATs
	if nats := r.GetNats(); len(nats) > 0 {
		natsBytes, err := json.Marshal(nats)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal NATs for %s: %w", r.GetName(), err)
		}
		data.NatsJSON = natsBytes
	}

	return data, nil
}
