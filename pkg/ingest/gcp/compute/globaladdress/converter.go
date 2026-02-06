package globaladdress

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertGlobalAddress converts a GCP API Address to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertGlobalAddress(a *computepb.Address, projectID string, collectedAt time.Time) (bronze.GCPComputeGlobalAddress, error) {
	addr := bronze.GCPComputeGlobalAddress{
		ResourceID:        fmt.Sprintf("%d", a.GetId()),
		Name:              a.GetName(),
		Description:       a.GetDescription(),
		Address:           a.GetAddress(),
		AddressType:       a.GetAddressType(),
		IpVersion:         a.GetIpVersion(),
		Ipv6EndpointType:  a.GetIpv6EndpointType(),
		IpCollection:      a.GetIpCollection(),
		Region:            a.GetRegion(),
		Status:            a.GetStatus(),
		Purpose:           a.GetPurpose(),
		Network:           a.GetNetwork(),
		Subnetwork:        a.GetSubnetwork(),
		NetworkTier:       a.GetNetworkTier(),
		PrefixLength:      a.GetPrefixLength(),
		SelfLink:          a.GetSelfLink(),
		CreationTimestamp: a.GetCreationTimestamp(),
		LabelFingerprint:  a.GetLabelFingerprint(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Convert users array to JSONB (nil → SQL NULL, data → JSON bytes)
	if a.Users != nil {
		var err error
		addr.UsersJSON, err = json.Marshal(a.Users)
		if err != nil {
			return bronze.GCPComputeGlobalAddress{}, fmt.Errorf("failed to marshal users for global address %s: %w", a.GetName(), err)
		}
	}

	// Convert labels to separate table
	addr.Labels = ConvertLabels(a.Labels)

	return addr, nil
}

// ConvertLabels converts global address labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeGlobalAddressLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeGlobalAddressLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeGlobalAddressLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}
