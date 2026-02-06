package forwardingrule

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertForwardingRule converts a GCP API ForwardingRule to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertForwardingRule(fr *computepb.ForwardingRule, projectID string, collectedAt time.Time) (bronze.GCPComputeForwardingRule, error) {
	rule := bronze.GCPComputeForwardingRule{
		ResourceID:                                          fmt.Sprintf("%d", fr.GetId()),
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
		rule.PortsJSON, err = json.Marshal(fr.Ports)
		if err != nil {
			return bronze.GCPComputeForwardingRule{}, fmt.Errorf("failed to marshal ports for forwarding rule %s: %w", fr.GetName(), err)
		}
	}

	if fr.SourceIpRanges != nil {
		var err error
		rule.SourceIpRangesJSON, err = json.Marshal(fr.SourceIpRanges)
		if err != nil {
			return bronze.GCPComputeForwardingRule{}, fmt.Errorf("failed to marshal source ip ranges for forwarding rule %s: %w", fr.GetName(), err)
		}
	}

	if fr.MetadataFilters != nil {
		var err error
		rule.MetadataFiltersJSON, err = json.Marshal(fr.MetadataFilters)
		if err != nil {
			return bronze.GCPComputeForwardingRule{}, fmt.Errorf("failed to marshal metadata filters for forwarding rule %s: %w", fr.GetName(), err)
		}
	}

	if fr.ServiceDirectoryRegistrations != nil {
		var err error
		rule.ServiceDirectoryRegistrationsJSON, err = json.Marshal(fr.ServiceDirectoryRegistrations)
		if err != nil {
			return bronze.GCPComputeForwardingRule{}, fmt.Errorf("failed to marshal service directory registrations for forwarding rule %s: %w", fr.GetName(), err)
		}
	}

	// Convert labels to separate table
	rule.Labels = ConvertLabels(fr.Labels)

	return rule, nil
}

// ConvertLabels converts forwarding rule labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeForwardingRuleLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeForwardingRuleLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeForwardingRuleLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}
