package disk

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// DiskData represents a GCP Compute disk in a data structure.
type DiskData struct {
	ID                        string
	Name                      string
	Description               string
	Zone                      string
	Region                    string
	Type                      string
	Status                    string
	SizeGb                    int64
	Architecture              string
	SelfLink                  string
	CreationTimestamp         string
	LastAttachTimestamp       string
	LastDetachTimestamp       string
	SourceImage               string
	SourceImageId             string
	SourceSnapshot            string
	SourceSnapshotId          string
	SourceDisk                string
	SourceDiskId              string
	ProvisionedIops           int64
	ProvisionedThroughput     int64
	PhysicalBlockSizeBytes    int64
	EnableConfidentialCompute bool
	DiskEncryptionKeyJSON     json.RawMessage
	UsersJSON                 json.RawMessage
	ReplicaZonesJSON          json.RawMessage
	ResourcePoliciesJSON      json.RawMessage
	GuestOsFeaturesJSON       json.RawMessage
	ProjectID                 string
	CollectedAt               time.Time

	Labels   []DiskLabelData
	Licenses []DiskLicenseData
}

// DiskLabelData represents a label attached to a disk.
type DiskLabelData struct {
	Key   string
	Value string
}

// DiskLicenseData represents a license attached to a disk.
type DiskLicenseData struct {
	License string
}

// ConvertDisk converts a GCP API Disk to DiskData.
// Preserves raw API data with minimal transformation.
func ConvertDisk(d *computepb.Disk, projectID string, collectedAt time.Time) (*DiskData, error) {
	disk := &DiskData{
		ID:                        fmt.Sprintf("%d", d.GetId()),
		Name:                      d.GetName(),
		Description:               d.GetDescription(),
		Zone:                      d.GetZone(),
		Region:                    d.GetRegion(),
		Type:                      d.GetType(),
		Status:                    d.GetStatus(),
		SizeGb:                    d.GetSizeGb(),
		Architecture:              d.GetArchitecture(),
		SelfLink:                  d.GetSelfLink(),
		CreationTimestamp:         d.GetCreationTimestamp(),
		LastAttachTimestamp:       d.GetLastAttachTimestamp(),
		LastDetachTimestamp:       d.GetLastDetachTimestamp(),
		SourceImage:               d.GetSourceImage(),
		SourceImageId:             d.GetSourceImageId(),
		SourceSnapshot:            d.GetSourceSnapshot(),
		SourceSnapshotId:          d.GetSourceSnapshotId(),
		SourceDisk:                d.GetSourceDisk(),
		SourceDiskId:              d.GetSourceDiskId(),
		ProvisionedIops:           d.GetProvisionedIops(),
		ProvisionedThroughput:     d.GetProvisionedThroughput(),
		PhysicalBlockSizeBytes:    d.GetPhysicalBlockSizeBytes(),
		EnableConfidentialCompute: d.GetEnableConfidentialCompute(),
		ProjectID:                 projectID,
		CollectedAt:               collectedAt,
	}

	// Convert JSONB fields (nil → SQL NULL, data → JSON bytes)
	var err error
	if d.DiskEncryptionKey != nil {
		disk.DiskEncryptionKeyJSON, err = json.Marshal(d.DiskEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.Users != nil {
		disk.UsersJSON, err = json.Marshal(d.Users)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.ReplicaZones != nil {
		disk.ReplicaZonesJSON, err = json.Marshal(d.ReplicaZones)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.ResourcePolicies != nil {
		disk.ResourcePoliciesJSON, err = json.Marshal(d.ResourcePolicies)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.GuestOsFeatures != nil {
		disk.GuestOsFeaturesJSON, err = json.Marshal(d.GuestOsFeatures)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}

	// Convert labels to separate table
	disk.Labels = ConvertLabels(d.Labels)

	// Convert licenses to separate table
	disk.Licenses = ConvertLicenses(d.Licenses)

	return disk, nil
}

// ConvertLabels converts disk labels from GCP API to DiskLabelData.
func ConvertLabels(labels map[string]string) []DiskLabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]DiskLabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, DiskLabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertLicenses converts disk licenses from GCP API to DiskLicenseData.
func ConvertLicenses(licenses []string) []DiskLicenseData {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]DiskLicenseData, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, DiskLicenseData{
			License: license,
		})
	}

	return result
}
