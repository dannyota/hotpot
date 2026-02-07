package image

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
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

// DiffImage compares old and new image states.
func DiffImage(old, new *bronze.GCPComputeImage) *ImageDiff {
	if old == nil {
		return &ImageDiff{
			IsNew:        true,
			LabelsDiff:   ChildDiff{Changed: true},
			LicensesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &ImageDiff{}
	diff.IsChanged = hasImageFieldsChanged(old, new)
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)
	diff.LicensesDiff = diffLicenses(old.Licenses, new.Licenses)

	return diff
}

// HasAnyChange returns true if any part of the image changed.
func (d *ImageDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed || d.LicensesDiff.Changed
}

func hasImageFieldsChanged(old, new *bronze.GCPComputeImage) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.Architecture != new.Architecture ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.Family != new.Family ||
		old.SourceDisk != new.SourceDisk ||
		old.SourceDiskId != new.SourceDiskId ||
		old.SourceImage != new.SourceImage ||
		old.SourceImageId != new.SourceImageId ||
		old.SourceSnapshot != new.SourceSnapshot ||
		old.SourceSnapshotId != new.SourceSnapshotId ||
		old.SourceType != new.SourceType ||
		old.DiskSizeGb != new.DiskSizeGb ||
		old.ArchiveSizeBytes != new.ArchiveSizeBytes ||
		old.SatisfiesPzi != new.SatisfiesPzi ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.EnableConfidentialCompute != new.EnableConfidentialCompute ||
		jsonb.Changed(old.ImageEncryptionKeyJSON, new.ImageEncryptionKeyJSON) ||
		jsonb.Changed(old.SourceDiskEncryptionKeyJSON, new.SourceDiskEncryptionKeyJSON) ||
		jsonb.Changed(old.SourceImageEncryptionKeyJSON, new.SourceImageEncryptionKeyJSON) ||
		jsonb.Changed(old.SourceSnapshotEncryptionKeyJSON, new.SourceSnapshotEncryptionKeyJSON) ||
		jsonb.Changed(old.DeprecatedJSON, new.DeprecatedJSON) ||
		jsonb.Changed(old.GuestOsFeaturesJSON, new.GuestOsFeaturesJSON) ||
		jsonb.Changed(old.ShieldedInstanceInitialStateJSON, new.ShieldedInstanceInitialStateJSON) ||
		jsonb.Changed(old.RawDiskJSON, new.RawDiskJSON) ||
		jsonb.Changed(old.StorageLocationsJSON, new.StorageLocationsJSON) ||
		jsonb.Changed(old.LicenseCodesJSON, new.LicenseCodesJSON)
}

func diffLabels(old, new []bronze.GCPComputeImageLabel) ChildDiff {
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

func diffLicenses(old, new []bronze.GCPComputeImageLicense) ChildDiff {
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
