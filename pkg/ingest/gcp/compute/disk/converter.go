package disk

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertDisk converts a GCP API Disk to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertDisk(d *computepb.Disk, projectID string, collectedAt time.Time) (bronze.GCPComputeDisk, error) {
	disk := bronze.GCPComputeDisk{
		ResourceID:                fmt.Sprintf("%d", d.GetId()),
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
			return bronze.GCPComputeDisk{}, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.Users != nil {
		disk.UsersJSON, err = json.Marshal(d.Users)
		if err != nil {
			return bronze.GCPComputeDisk{}, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.ReplicaZones != nil {
		disk.ReplicaZonesJSON, err = json.Marshal(d.ReplicaZones)
		if err != nil {
			return bronze.GCPComputeDisk{}, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.ResourcePolicies != nil {
		disk.ResourcePoliciesJSON, err = json.Marshal(d.ResourcePolicies)
		if err != nil {
			return bronze.GCPComputeDisk{}, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}
	if d.GuestOsFeatures != nil {
		disk.GuestOsFeaturesJSON, err = json.Marshal(d.GuestOsFeatures)
		if err != nil {
			return bronze.GCPComputeDisk{}, fmt.Errorf("failed to marshal JSON for disk %s: %w", d.GetName(), err)
		}
	}

	// Convert labels to separate table
	disk.Labels = ConvertLabels(d.Labels)

	// Convert licenses to separate table
	disk.Licenses = ConvertLicenses(d.Licenses)

	return disk, nil
}

// ConvertLabels converts disk labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeDiskLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeDiskLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeDiskLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertLicenses converts disk licenses from GCP API to Bronze models.
func ConvertLicenses(licenses []string) []bronze.GCPComputeDiskLicense {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeDiskLicense, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, bronze.GCPComputeDiskLicense{
			License: license,
		})
	}

	return result
}
