package image

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertImage converts a GCP API Image to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertImage(s *computepb.Image, projectID string, collectedAt time.Time) (bronze.GCPComputeImage, error) {
	img := bronze.GCPComputeImage{
		ResourceID:                fmt.Sprintf("%d", s.GetId()),
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
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.SourceDiskEncryptionKey != nil {
		img.SourceDiskEncryptionKeyJSON, err = json.Marshal(s.SourceDiskEncryptionKey)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.SourceImageEncryptionKey != nil {
		img.SourceImageEncryptionKeyJSON, err = json.Marshal(s.SourceImageEncryptionKey)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.SourceSnapshotEncryptionKey != nil {
		img.SourceSnapshotEncryptionKeyJSON, err = json.Marshal(s.SourceSnapshotEncryptionKey)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.Deprecated != nil {
		img.DeprecatedJSON, err = json.Marshal(s.Deprecated)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.GuestOsFeatures != nil {
		img.GuestOsFeaturesJSON, err = json.Marshal(s.GuestOsFeatures)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.ShieldedInstanceInitialState != nil {
		img.ShieldedInstanceInitialStateJSON, err = json.Marshal(s.ShieldedInstanceInitialState)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.RawDisk != nil {
		img.RawDiskJSON, err = json.Marshal(s.RawDisk)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.StorageLocations != nil {
		img.StorageLocationsJSON, err = json.Marshal(s.StorageLocations)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}
	if s.LicenseCodes != nil {
		img.LicenseCodesJSON, err = json.Marshal(s.LicenseCodes)
		if err != nil {
			return bronze.GCPComputeImage{}, fmt.Errorf("failed to marshal JSON for image %s: %w", s.GetName(), err)
		}
	}

	// Convert labels to separate table
	img.Labels = ConvertLabels(s.Labels)

	// Convert licenses to separate table
	img.Licenses = ConvertLicenses(s.Licenses)

	return img, nil
}

// ConvertLabels converts image labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPComputeImageLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeImageLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPComputeImageLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertLicenses converts image licenses from GCP API to Bronze models.
func ConvertLicenses(licenses []string) []bronze.GCPComputeImageLicense {
	if len(licenses) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeImageLicense, 0, len(licenses))
	for _, license := range licenses {
		result = append(result, bronze.GCPComputeImageLicense{
			License: license,
		})
	}

	return result
}
