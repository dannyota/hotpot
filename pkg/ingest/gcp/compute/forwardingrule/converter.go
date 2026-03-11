package forwardingrule

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// ForwardingRuleData holds converted forwarding rule data ready for Ent insertion.
type ForwardingRuleData struct {
	ID                                                      string
	Name                                                    string
	Description                                             string
	IPAddress                                               string
	IPProtocol                                              string
	AllPorts                                                bool
	AllowGlobalAccess                                       bool
	AllowPscGlobalAccess                                    bool
	BackendService                                          string
	BaseForwardingRule                                      string
	CreationTimestamp                                       string
	ExternalManagedBackendBucketMigrationState              string
	ExternalManagedBackendBucketMigrationTestingPercentage  float32
	Fingerprint                                             string
	IpCollection                                            string
	IpVersion                                               string
	IsMirroringCollector                                    bool
	LabelFingerprint                                        string
	LoadBalancingScheme                                     string
	Network                                                 string
	NetworkTier                                             string
	NoAutomateDnsZone                                       bool
	PortRange                                               string
	PscConnectionId                                         string
	PscConnectionStatus                                     string
	Region                                                  string
	SelfLink                                                string
	SelfLinkWithId                                          string
	ServiceLabel                                            string
	ServiceName                                             string
	Subnetwork                                              string
	Target                                                  string
	PortsJSON                                               []interface{}
	SourceIpRangesJSON                                      []interface{}
	MetadataFiltersJSON                                     []interface{}
	ServiceDirectoryRegistrationsJSON                       []interface{}
	Labels                                                  []LabelData
	ProjectID                                               string
	CollectedAt                                             time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertForwardingRule converts a GCP API ForwardingRule to ForwardingRuleData.
// Preserves raw API data with minimal transformation.
func ConvertForwardingRule(fr *computepb.ForwardingRule, projectID string, collectedAt time.Time) (*ForwardingRuleData, error) {
	rule := &ForwardingRuleData{
		ID:                                                  fmt.Sprintf("%d", fr.GetId()),
		Name:                                                fr.GetName(),
		Description:                                         fr.GetDescription(),
		IPAddress:                                           fr.GetIPAddress(),
		IPProtocol:                                          fr.GetIPProtocol(),
		AllPorts:                                            fr.GetAllPorts(),
		AllowGlobalAccess:                                   fr.GetAllowGlobalAccess(),
		AllowPscGlobalAccess:                                fr.GetAllowPscGlobalAccess(),
		BackendService:                                      fr.GetBackendService(),
		BaseForwardingRule:                                  fr.GetBaseForwardingRule(),
		CreationTimestamp:                                   fr.GetCreationTimestamp(),
		ExternalManagedBackendBucketMigrationState:          fr.GetExternalManagedBackendBucketMigrationState(),
		ExternalManagedBackendBucketMigrationTestingPercentage: fr.GetExternalManagedBackendBucketMigrationTestingPercentage(),
		Fingerprint:                                         fr.GetFingerprint(),
		IpCollection:                                        fr.GetIpCollection(),
		IpVersion:                                           fr.GetIpVersion(),
		IsMirroringCollector:                                fr.GetIsMirroringCollector(),
		LabelFingerprint:                                    fr.GetLabelFingerprint(),
		LoadBalancingScheme:                                 fr.GetLoadBalancingScheme(),
		Network:                                             fr.GetNetwork(),
		NetworkTier:                                         fr.GetNetworkTier(),
		NoAutomateDnsZone:                                   fr.GetNoAutomateDnsZone(),
		PortRange:                                           fr.GetPortRange(),
		PscConnectionId:                                     fmt.Sprintf("%d", fr.GetPscConnectionId()),
		PscConnectionStatus:                                 fr.GetPscConnectionStatus(),
		Region:                                              fr.GetRegion(),
		SelfLink:                                            fr.GetSelfLink(),
		SelfLinkWithId:                                      fr.GetSelfLinkWithId(),
		ServiceLabel:                                        fr.GetServiceLabel(),
		ServiceName:                                         fr.GetServiceName(),
		Subnetwork:                                          fr.GetSubnetwork(),
		Target:                                              fr.GetTarget(),
		ProjectID:                                           projectID,
		CollectedAt:                                         collectedAt,
	}

	// Convert JSONB fields (nil → SQL NULL, data → JSON bytes)
	if fr.Ports != nil {
		var err error
		portsBytes, err := json.Marshal(fr.Ports)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ports for forwarding rule %s: %w", fr.GetName(), err)
		}
		if err := json.Unmarshal(portsBytes, &rule.PortsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ports: %w", err)
		}
	}

	if fr.SourceIpRanges != nil {
		var err error
		sourceBytes, err := json.Marshal(fr.SourceIpRanges)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source ip ranges for forwarding rule %s: %w", fr.GetName(), err)
		}
		if err := json.Unmarshal(sourceBytes, &rule.SourceIpRangesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal source ip ranges: %w", err)
		}
	}

	if fr.MetadataFilters != nil {
		var err error
		metaBytes, err := json.Marshal(fr.MetadataFilters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata filters for forwarding rule %s: %w", fr.GetName(), err)
		}
		if err := json.Unmarshal(metaBytes, &rule.MetadataFiltersJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata filters: %w", err)
		}
	}

	if fr.ServiceDirectoryRegistrations != nil {
		var err error
		serviceBytes, err := json.Marshal(fr.ServiceDirectoryRegistrations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal service directory registrations for forwarding rule %s: %w", fr.GetName(), err)
		}
		if err := json.Unmarshal(serviceBytes, &rule.ServiceDirectoryRegistrationsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service directory registrations: %w", err)
		}
	}

	// Convert labels to separate table
	rule.Labels = ConvertLabels(fr.Labels)

	return rule, nil
}

// ConvertLabels converts forwarding rule labels from GCP API to label data.
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
