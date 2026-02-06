package disk

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// DiskDiff represents changes between old and new disk states.
type DiskDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelsDiff   ChildDiff
	LicensesDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffDisk compares old and new disk states.
// Returns nil if old is nil (new disk).
func DiffDisk(old, new *bronze.GCPComputeDisk) *DiskDiff {
	if old == nil {
		return &DiskDiff{
			IsNew:        true,
			LabelsDiff:   ChildDiff{Changed: true},
			LicensesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &DiskDiff{}

	// Compare disk-level fields
	diff.IsChanged = hasDiskFieldsChanged(old, new)

	// Compare children
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)
	diff.LicensesDiff = diffLicenses(old.Licenses, new.Licenses)

	return diff
}

// HasAnyChange returns true if any part of the disk changed.
func (d *DiskDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed || d.LicensesDiff.Changed
}

// hasDiskFieldsChanged compares disk-level fields (excluding children).
func hasDiskFieldsChanged(old, new *bronze.GCPComputeDisk) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Zone != new.Zone ||
		old.Region != new.Region ||
		old.Type != new.Type ||
		old.Status != new.Status ||
		old.SizeGb != new.SizeGb ||
		old.Architecture != new.Architecture ||
		old.LastAttachTimestamp != new.LastAttachTimestamp ||
		old.LastDetachTimestamp != new.LastDetachTimestamp ||
		old.SourceImage != new.SourceImage ||
		old.SourceImageId != new.SourceImageId ||
		old.SourceSnapshot != new.SourceSnapshot ||
		old.SourceSnapshotId != new.SourceSnapshotId ||
		old.SourceDisk != new.SourceDisk ||
		old.SourceDiskId != new.SourceDiskId ||
		old.ProvisionedIops != new.ProvisionedIops ||
		old.ProvisionedThroughput != new.ProvisionedThroughput ||
		old.PhysicalBlockSizeBytes != new.PhysicalBlockSizeBytes ||
		old.EnableConfidentialCompute != new.EnableConfidentialCompute ||
		jsonb.Changed(old.DiskEncryptionKeyJSON, new.DiskEncryptionKeyJSON) ||
		jsonb.Changed(old.UsersJSON, new.UsersJSON) ||
		jsonb.Changed(old.ReplicaZonesJSON, new.ReplicaZonesJSON) ||
		jsonb.Changed(old.ResourcePoliciesJSON, new.ResourcePoliciesJSON) ||
		jsonb.Changed(old.GuestOsFeaturesJSON, new.GuestOsFeaturesJSON)
}

func diffLabels(old, new []bronze.GCPComputeDiskLabel) ChildDiff {
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

func diffLicenses(old, new []bronze.GCPComputeDiskLicense) ChildDiff {
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
