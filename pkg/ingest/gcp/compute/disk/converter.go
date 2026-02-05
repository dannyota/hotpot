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
func ConvertDisk(d *computepb.Disk, projectID string, collectedAt time.Time) bronze.GCPComputeDisk {
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

	// Convert encryption key to JSON
	if d.DiskEncryptionKey != nil {
		if data, err := json.Marshal(d.DiskEncryptionKey); err == nil {
			disk.DiskEncryptionKeyJSON = string(data)
		}
	}

	// Convert users array to JSON
	if len(d.Users) > 0 {
		if data, err := json.Marshal(d.Users); err == nil {
			disk.UsersJSON = string(data)
		}
	}

	// Convert replica zones to JSON
	if len(d.ReplicaZones) > 0 {
		if data, err := json.Marshal(d.ReplicaZones); err == nil {
			disk.ReplicaZonesJSON = string(data)
		}
	}

	// Convert resource policies to JSON
	if len(d.ResourcePolicies) > 0 {
		if data, err := json.Marshal(d.ResourcePolicies); err == nil {
			disk.ResourcePoliciesJSON = string(data)
		}
	}

	// Convert guest OS features to JSON
	if len(d.GuestOsFeatures) > 0 {
		if data, err := json.Marshal(d.GuestOsFeatures); err == nil {
			disk.GuestOsFeaturesJSON = string(data)
		}
	}

	// Convert labels to separate table
	disk.Labels = ConvertLabels(d.Labels)

	// Convert licenses to separate table
	disk.Licenses = ConvertLicenses(d.Licenses)

	return disk
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
