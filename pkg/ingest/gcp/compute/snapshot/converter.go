package snapshot

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// SnapshotData represents a GCP Compute snapshot in a data structure.
type SnapshotData struct {
	ID                              string
	Name                            string
	Description                     string
	Status                          string
	DiskSizeGB                      int64
	StorageBytes                    int64
	StorageBytesStatus              string
	DownloadBytes                   int64
	SnapshotType                    string
	Architecture                    string
	SelfLink                        string
	CreationTimestamp               string
	LabelFingerprint                string
	SourceDisk                      string
	SourceDiskID                    string
	SourceDiskForRecoveryCheckpoint string
	AutoCreated                     bool
	SatisfiesPzi                    bool
	SatisfiesPzs                    bool
	EnableConfidentialCompute       bool
	SnapshotEncryptionKeyJSON       json.RawMessage
	SourceDiskEncryptionKeyJSON     json.RawMessage
	GuestOsFeaturesJSON             json.RawMessage
	StorageLocationsJSON            json.RawMessage
	ProjectID                       string
	CollectedAt                     time.Time

	Labels   []SnapshotLabelData
	Licenses []SnapshotLicenseData
}

// SnapshotLabelData represents a label attached to a snapshot.
type SnapshotLabelData struct {
	Key   string
	Value string
}

// SnapshotLicenseData represents a license attached to a snapshot.
type SnapshotLicenseData struct {
	License string
}

// ConvertSnapshot converts a GCP API Snapshot to SnapshotData.
// Preserves raw API data with minimal transformation.
func ConvertSnapshot(s *computepb.Snapshot, projectID string, collectedAt time.Time) (*SnapshotData, error) {
	snap := &SnapshotData{
		ID:                              fmt.Sprintf("%d", s.GetId()),
		Name:                            s.GetName(),
		Description:                     s.GetDescription(),
		Status:                          s.GetStatus(),
		DiskSizeGB:                      s.GetDiskSizeGb(),
		StorageBytes:                    s.GetStorageBytes(),
		StorageBytesStatus:              s.GetStorageBytesStatus(),
		DownloadBytes:                   s.GetDownloadBytes(),
		SnapshotType:                    s.GetSnapshotType(),
		Architecture:                    s.GetArchitecture(),
		SelfLink:                        s.GetSelfLink(),
		CreationTimestamp:               s.GetCreationTimestamp(),
		LabelFingerprint:                s.GetLabelFingerprint(),
		SourceDisk:                      s.GetSourceDisk(),
		SourceDiskID:                    s.GetSourceDiskId(),
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
			return nil, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}
	if s.SourceDiskEncryptionKey != nil {
		snap.SourceDiskEncryptionKeyJSON, err = json.Marshal(s.SourceDiskEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}
	if s.GuestOsFeatures != nil {
		snap.GuestOsFeaturesJSON, err = json.Marshal(s.GuestOsFeatures)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}
	if s.StorageLocations != nil {
		snap.StorageLocationsJSON, err = json.Marshal(s.StorageLocations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for snapshot %s: %w", s.GetName(), err)
		}
	}

	// Convert labels to separate table
	snap.Labels = ConvertLabels(s.Labels)

	// Convert licenses to separate table
	snap.Licenses = ConvertLicenses(s.Licenses)

	return snap, nil
}

// ConvertLabels converts snapshot labels from GCP API to SnapshotLabelData.
func ConvertLabels(labels map[string]string) []SnapshotLabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]SnapshotLabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, SnapshotLabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertLicenses converts snapshot licenses from GCP API to SnapshotLicenseData.
func ConvertLicenses(licenses []string) []SnapshotLicenseData {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]SnapshotLicenseData, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, SnapshotLicenseData{
			License: license,
		})
	}

	return result
}
