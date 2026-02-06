package address

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertAddress converts a GCP API Address to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertAddress(a *computepb.Address, projectID string, collectedAt time.Time) (bronze.GCPComputeAddress, error) {
	addr := bronze.GCPComputeAddress{
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
			return bronze.GCPComputeAddress{}, fmt.Errorf("failed to marshal users for address %s: %w", a.GetName(), err)
		}
	}

	// Convert labels to separate table
	addr.Labels = ConvertLabels(a.Labels)

	return addr, nil
}

// ConvertLabels converts address labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeAddressLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeAddressLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeAddressLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}
