package subnetwork

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// SubnetworkData holds converted subnetwork data ready for Ent insertion.
type SubnetworkData struct {
	ID                      string
	Name                    string
	Description             string
	SelfLink                string
	CreationTimestamp       string
	Network                 string
	Region                  string
	IpCidrRange             string
	GatewayAddress          string
	Purpose                 string
	Role                    string
	PrivateIpGoogleAccess   bool
	PrivateIpv6GoogleAccess string
	StackType               string
	Ipv6AccessType          string
	InternalIpv6Prefix      string
	ExternalIpv6Prefix      string
	LogConfigJSON           json.RawMessage
	Fingerprint             string
	SecondaryIpRanges       []SecondaryRangeData
	ProjectID               string
	CollectedAt             time.Time
}

// SecondaryRangeData holds converted secondary range data.
type SecondaryRangeData struct {
	RangeName   string
	IpCidrRange string
}

// ConvertSubnetwork converts a GCP API Subnetwork to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertSubnetwork(s *computepb.Subnetwork, projectID string, collectedAt time.Time) (*SubnetworkData, error) {
	data := &SubnetworkData{
		ID:                      fmt.Sprintf("%d", s.GetId()),
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
		data.LogConfigJSON, err = json.Marshal(s.LogConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal log config for subnetwork %s: %w", s.GetName(), err)
		}
	}

	// Convert secondary IP ranges to separate table
	data.SecondaryIpRanges = ConvertSecondaryRanges(s.SecondaryIpRanges)

	return data, nil
}

// ConvertSecondaryRanges converts secondary IP ranges from GCP API to data structs.
func ConvertSecondaryRanges(ranges []*computepb.SubnetworkSecondaryRange) []SecondaryRangeData {
	if len(ranges) == 0 {
		return nil
	}

	result := make([]SecondaryRangeData, 0, len(ranges))
	for _, r := range ranges {
		result = append(result, SecondaryRangeData{
			RangeName:   r.GetRangeName(),
			IpCidrRange: r.GetIpCidrRange(),
		})
	}

	return result
}
