package address

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// AddressData represents a GCP Compute address in a data structure.
type AddressData struct {
	ID                string
	Name              string
	Description       string
	Address           string
	AddressType       string
	IpVersion         string
	Ipv6EndpointType  string
	IpCollection      string
	Region            string
	Status            string
	Purpose           string
	Network           string
	Subnetwork        string
	NetworkTier       string
	PrefixLength      int32
	SelfLink          string
	CreationTimestamp string
	LabelFingerprint  string
	UsersJSON         json.RawMessage
	ProjectID         string
	CollectedAt       time.Time

	Labels []AddressLabelData
}

// AddressLabelData represents a label attached to an address.
type AddressLabelData struct {
	Key   string
	Value string
}

// ConvertAddress converts a GCP API Address to AddressData.
// Preserves raw API data with minimal transformation.
func ConvertAddress(a *computepb.Address, projectID string, collectedAt time.Time) (*AddressData, error) {
	addr := &AddressData{
		ID:                fmt.Sprintf("%d", a.GetId()),
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
			return nil, fmt.Errorf("failed to marshal users for address %s: %w", a.GetName(), err)
		}
	}

	// Convert labels to separate table
	addr.Labels = ConvertLabels(a.Labels)

	return addr, nil
}

// ConvertLabels converts address labels from GCP API to AddressLabelData.
func ConvertLabels(labels map[string]string) []AddressLabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]AddressLabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, AddressLabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}
