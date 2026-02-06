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
func ConvertSubnetwork(s *computepb.Subnetwork, projectID string, collectedAt time.Time) (bronze.GCPComputeSubnetwork, error) {
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

	// Convert log config to JSONB (nil → SQL NULL, data → JSON bytes)
	if s.LogConfig != nil {
		var err error
		subnet.LogConfigJSON, err = json.Marshal(s.LogConfig)
		if err != nil {
			return bronze.GCPComputeSubnetwork{}, fmt.Errorf("failed to marshal log config for subnetwork %s: %w", s.GetName(), err)
		}
	}

	// Convert secondary IP ranges to separate table
	subnet.SecondaryIpRanges = ConvertSecondaryRanges(s.SecondaryIpRanges)

	return subnet, nil
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
