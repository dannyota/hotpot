package image

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// ImageData represents a GCP Compute image in a data structure.
type ImageData struct {
	ID                               string
	Name                             string
	Description                      string
	Status                           string
	Architecture                     string
	SelfLink                         string
	CreationTimestamp                string
	LabelFingerprint                 string
	Family                           string
	SourceDisk                       string
	SourceDiskId                     string
	SourceImage                      string
	SourceImageId                    string
	SourceSnapshot                   string
	SourceSnapshotId                 string
	SourceType                       string
	DiskSizeGb                       int64
	ArchiveSizeBytes                 int64
	SatisfiesPzi                     bool
	SatisfiesPzs                     bool
	EnableConfidentialCompute        bool
	ImageEncryptionKeyJSON           json.RawMessage
	SourceDiskEncryptionKeyJSON      json.RawMessage
	SourceImageEncryptionKeyJSON     json.RawMessage
	SourceSnapshotEncryptionKeyJSON  json.RawMessage
	DeprecatedJSON                   json.RawMessage
	GuestOsFeaturesJSON              json.RawMessage
	ShieldedInstanceInitialStateJSON json.RawMessage
	RawDiskJSON                      json.RawMessage
	StorageLocationsJSON             json.RawMessage
	LicenseCodesJSON                 json.RawMessage
	ProjectID                        string
	CollectedAt                      time.Time

	Labels   []ImageLabelData
	Licenses []ImageLicenseData
}

// ImageLabelData represents a label attached to an image.
type ImageLabelData struct {
	Key   string
	Value string
}

// ImageLicenseData represents a license attached to an image.
type ImageLicenseData struct {
	License string
}

// ConvertImage converts a GCP API Image to ImageData.
// Preserves raw API data with minimal transformation.
func ConvertImage(s *computepb.Image, projectID string, collectedAt time.Time) (*ImageData, error) {
	img := &ImageData{
		ID:                        fmt.Sprintf("%d", s.GetId()),
		Name:                      s.GetName(),
		Description:               s.GetDescription(),
		Status:                    s.GetStatus(),
		Architecture:              s.GetArchitecture(),
		SelfLink:                  s.GetSelfLink(),
		CreationTimestamp:         s.GetCreationTimestamp(),
		LabelFingerprint:          s.GetLabelFingerprint(),
		Family:                    s.GetFamily(),
		SourceDisk:                s.GetSourceDisk(),
		SourceDiskId:              s.GetSourceDiskId(),
		SourceImage:               s.GetSourceImage(),
		SourceImageId:             s.GetSourceImageId(),
		SourceSnapshot:            s.GetSourceSnapshot(),
		SourceSnapshotId:          s.GetSourceSnapshotId(),
		SourceType:                s.GetSourceType(),
		DiskSizeGb:                s.GetDiskSizeGb(),
		ArchiveSizeBytes:          s.GetArchiveSizeBytes(),
		SatisfiesPzi:              s.GetSatisfiesPzi(),
		SatisfiesPzs:              s.GetSatisfiesPzs(),
		EnableConfidentialCompute: s.GetEnableConfidentialCompute(),
		ProjectID:                 projectID,
		CollectedAt:               collectedAt,
	}

	// Convert JSONB fields (nil -> SQL NULL, data -> JSON bytes)
	var err error
	if s.ImageEncryptionKey != nil {
		img.ImageEncryptionKeyJSON, err = json.Marshal(s.ImageEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.SourceDiskEncryptionKey != nil {
		img.SourceDiskEncryptionKeyJSON, err = json.Marshal(s.SourceDiskEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.SourceImageEncryptionKey != nil {
		img.SourceImageEncryptionKeyJSON, err = json.Marshal(s.SourceImageEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.SourceSnapshotEncryptionKey != nil {
		img.SourceSnapshotEncryptionKeyJSON, err = json.Marshal(s.SourceSnapshotEncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.Deprecated != nil {
		img.DeprecatedJSON, err = json.Marshal(s.Deprecated)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.GuestOsFeatures != nil {
		img.GuestOsFeaturesJSON, err = json.Marshal(s.GuestOsFeatures)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.ShieldedInstanceInitialState != nil {
		img.ShieldedInstanceInitialStateJSON, err = json.Marshal(s.ShieldedInstanceInitialState)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.RawDisk != nil {
		img.RawDiskJSON, err = json.Marshal(s.RawDisk)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.StorageLocations != nil {
		img.StorageLocationsJSON, err = json.Marshal(s.StorageLocations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.LicenseCodes != nil {
		img.LicenseCodesJSON, err = json.Marshal(s.LicenseCodes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}

	// Convert labels to separate table
	img.Labels = ConvertLabels(s.Labels)

	// Convert licenses to separate table
	img.Licenses = ConvertLicenses(s.Licenses)

	return img, nil
}

// ConvertLabels converts image labels from GCP API to ImageLabelData.
func ConvertLabels(labels map[string]string) []ImageLabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]ImageLabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, ImageLabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertLicenses converts image licenses from GCP API to ImageLicenseData.
func ConvertLicenses(licenses []string) []ImageLicenseData {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]ImageLicenseData, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, ImageLicenseData{
			License: license,
		})
	}

	return result
}
