package subnetwork

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertSubnetwork converts a GCP API Subnetwork to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertSubnetwork(s *computepb.Subnetwork, projectID string, collectedAt time.Time) bronze.GCPComputeSubnetwork {
	subnet := bronze.GCPComputeSubnetwork{
		ResourceID:              fmt.Sprintf("%d", s.GetId()),
		Name:                    s.GetName(),
		Description:             s.GetDescription(),
		SelfLink:                s.GetSelfLink(),
		CreationTimestamp:       s.GetCreationTimestamp(),
		Network:                 s.GetNetwork(),
		Region:                  s.GetRegion(),
		IpCidrRange:             s.GetIpCidrRange(),
		GatewayAddress:          s.GetGatewayAddress(),
		Purpose:                 s.GetPurpose(),
		Role:                    s.GetRole(),
		PrivateIpGoogleAccess:   s.GetPrivateIpGoogleAccess(),
		PrivateIpv6GoogleAccess: s.GetPrivateIpv6GoogleAccess(),
		StackType:               s.GetStackType(),
		Ipv6AccessType:          s.GetIpv6AccessType(),
		InternalIpv6Prefix:      s.GetInternalIpv6Prefix(),
		ExternalIpv6Prefix:      s.GetExternalIpv6Prefix(),
		Fingerprint:             s.GetFingerprint(),
		ProjectID:               projectID,
		CollectedAt:             collectedAt,
	}

	// Convert log config to JSON
	if s.LogConfig != nil {
		if data, err := json.Marshal(s.LogConfig); err == nil {
			subnet.LogConfigJSON = string(data)
		}
	}

	// Convert secondary IP ranges to separate table
	subnet.SecondaryIpRanges = ConvertSecondaryRanges(s.SecondaryIpRanges)

	return subnet
}

// ConvertSecondaryRanges converts secondary IP ranges from GCP API to Bronze models.
func ConvertSecondaryRanges(ranges []*computepb.SubnetworkSecondaryRange) []bronze.GCPComputeSubnetworkSecondaryRange {
	if len(ranges) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeSubnetworkSecondaryRange, 0, len(ranges))
	for _, r := range ranges {
		result = append(result, bronze.GCPComputeSubnetworkSecondaryRange{
			RangeName:   r.GetRangeName(),
			IpCidrRange: r.GetIpCidrRange(),
		})
	}

	return result
}
