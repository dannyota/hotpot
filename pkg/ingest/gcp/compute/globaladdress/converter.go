package globaladdress

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// GlobalAddressData represents a GCP Compute global address in a data structure.
type GlobalAddressData struct {
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

	Labels []GlobalAddressLabelData
}

// GlobalAddressLabelData represents a label attached to a global address.
type GlobalAddressLabelData struct {
	Key   string
	Value string
}

// ConvertGlobalAddress converts a GCP API Address to GlobalAddressData.
// Preserves raw API data with minimal transformation.
func ConvertGlobalAddress(a *computepb.Address, projectID string, collectedAt time.Time) (*GlobalAddressData, error) {
	addr := &GlobalAddressData{
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
			return nil, fmt.Errorf("failed to marshal users for global address %s: %w", a.GetName(), err)
		}
	}

	// Convert labels to separate table
	addr.Labels = ConvertLabels(a.Labels)

	return addr, nil
}

// ConvertLabels converts global address labels from GCP API to GlobalAddressLabelData.
func ConvertLabels(labels map[string]string) []GlobalAddressLabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]GlobalAddressLabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, GlobalAddressLabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}
