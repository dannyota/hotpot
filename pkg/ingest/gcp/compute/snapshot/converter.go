package snapshot

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertSnapshot converts a GCP API Snapshot to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertSnapshot(s *computepb.Snapshot, projectID string, collectedAt time.Time) (bronze.GCPComputeSnapshot, error) {
	snap := bronze.GCPComputeSnapshot{
		ResourceID:                      fmt.Sprintf("%d", s.GetId()),
		Name:                            s.GetName(),
		Description:                     s.GetDescription(),
		Status:                          s.GetStatus(),
		DiskSizeGb:                      s.GetDiskSizeGb(),
		StorageBytes:                    s.GetStorageBytes(),
		StorageBytesStatus:              s.GetStorageBytesStatus(),
		DownloadBytes:                   s.GetDownloadBytes(),
		SnapshotType:                    s.GetSnapshotType(),
		Architecture:                    s.GetArchitecture(),
		SelfLink:                        s.GetSelfLink(),
		CreationTimestamp:               s.GetCreationTimestamp(),
		LabelFingerprint:                s.GetLabelFingerprint(),
		SourceDisk:                      s.GetSourceDisk(),
		SourceDiskId:                    s.GetSourceDiskId(),
		SourceDiskForRecoveryCheckpoint: s.GetSourceDiskForRecoveryCheckpoint(),
		AutoCreated:                     s.GetAutoCreated(),
		SatisfiesPzi:                    s.GetSatisfiesPzi(),
		SatisfiesPzs:                    s.GetSatisfiesPzs(),
		EnableConfidentialCompute:       s.GetEnableConfidentialCompute(),
		ProjectID:                       projectID,
		CollectedAt:                     collectedAt,
	}

	// Convert JSONB fields (nil → SQL NULL, data → JSON bytes)
	var err error
	if s.SnapshotEncryptionKey != nil {
		snap.SnapshotEncryptionKeyJSON, err = json.Marshal(s.SnapshotEncryptionKey)
		if err != nil {
			return bronze.GCPComputeSnapshot{}, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}
	if s.SourceDiskEncryptionKey != nil {
		snap.SourceDiskEncryptionKeyJSON, err = json.Marshal(s.SourceDiskEncryptionKey)
		if err != nil {
			return bronze.GCPComputeSnapshot{}, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}
	if s.GuestOsFeatures != nil {
		snap.GuestOsFeaturesJSON, err = json.Marshal(s.GuestOsFeatures)
		if err != nil {
			return bronze.GCPComputeSnapshot{}, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}
	if s.StorageLocations != nil {
		snap.StorageLocationsJSON, err = json.Marshal(s.StorageLocations)
		if err != nil {
			return bronze.GCPComputeSnapshot{}, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}

	// Convert labels to separate table
	snap.Labels = ConvertLabels(s.Labels)

	// Convert licenses to separate table
	snap.Licenses = ConvertLicenses(s.Licenses)

	return snap, nil
}

// ConvertLabels converts snapshot labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeSnapshotLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeSnapshotLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeSnapshotLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertLicenses converts snapshot licenses from GCP API to Bronze models.
func ConvertLicenses(licenses []string) []bronze.GCPComputeSnapshotLicense {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeSnapshotLicense, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, bronze.GCPComputeSnapshotLicense{
			License: license,
		})
	}

	return result
}
