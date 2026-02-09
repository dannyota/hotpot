package image

import (
	"bytes"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ImageDiff represents changes between old and new image states.
type ImageDiff struct {
	IsNew     bool
	IsChanged bool

	LabelsDiff   ChildDiff
	LicensesDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffImageData compares old Ent entity and new ImageData.
func DiffImageData(old *ent.BronzeGCPComputeImage, new *ImageData) *ImageDiff {
	if old == nil {
		return &ImageDiff{
			IsNew:        true,
			LabelsDiff:   ChildDiff{Changed: true},
			LicensesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &ImageDiff{}
	diff.IsChanged = hasImageFieldsChanged(old, new)
	diff.LabelsDiff = diffLabels(old.Edges.Labels, new.Labels)
	diff.LicensesDiff = diffLicenses(old.Edges.Licenses, new.Licenses)

	return diff
}

// HasAnyChange returns true if any part of the image changed.
func (d *ImageDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed || d.LicensesDiff.Changed
}

func hasImageFieldsChanged(old *ent.BronzeGCPComputeImage, new *ImageData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.Architecture != new.Architecture ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.Family != new.Family ||
		old.SourceDisk != new.SourceDisk ||
		old.SourceDiskID != new.SourceDiskId ||
		old.SourceImage != new.SourceImage ||
		old.SourceImageID != new.SourceImageId ||
		old.SourceSnapshot != new.SourceSnapshot ||
		old.SourceSnapshotID != new.SourceSnapshotId ||
		old.SourceType != new.SourceType ||
		old.DiskSizeGB != new.DiskSizeGb ||
		old.ArchiveSizeBytes != new.ArchiveSizeBytes ||
		old.SatisfiesPzi != new.SatisfiesPzi ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.EnableConfidentialCompute != new.EnableConfidentialCompute ||
		!bytes.Equal(old.ImageEncryptionKeyJSON, new.ImageEncryptionKeyJSON) ||
		!bytes.Equal(old.SourceDiskEncryptionKeyJSON, new.SourceDiskEncryptionKeyJSON) ||
		!bytes.Equal(old.SourceImageEncryptionKeyJSON, new.SourceImageEncryptionKeyJSON) ||
		!bytes.Equal(old.SourceSnapshotEncryptionKeyJSON, new.SourceSnapshotEncryptionKeyJSON) ||
		!bytes.Equal(old.DeprecatedJSON, new.DeprecatedJSON) ||
		!bytes.Equal(old.GuestOsFeaturesJSON, new.GuestOsFeaturesJSON) ||
		!bytes.Equal(old.ShieldedInstanceInitialStateJSON, new.ShieldedInstanceInitialStateJSON) ||
		!bytes.Equal(old.RawDiskJSON, new.RawDiskJSON) ||
		!bytes.Equal(old.StorageLocationsJSON, new.StorageLocationsJSON) ||
		!bytes.Equal(old.LicenseCodesJSON, new.LicenseCodesJSON)
}

func diffLabels(old []*ent.BronzeGCPComputeImageLabel, new []ImageLabelData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	oldMap := make(map[string]string)
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}
	for _, l := range new {
		if v, ok := oldMap[l.Key]; !ok || v != l.Value {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}

func diffLicenses(old []*ent.BronzeGCPComputeImageLicense, new []ImageLicenseData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	oldSet := make(map[string]bool)
	for _, l := range old {
		oldSet[l.License] = true
	}
	for _, l := range new {
		if !oldSet[l.License] {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
