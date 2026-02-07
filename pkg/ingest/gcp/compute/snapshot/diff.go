package snapshot

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// SnapshotDiff represents changes between old and new snapshot states.
type SnapshotDiff struct {
	IsNew     bool
	IsChanged bool

	LabelsDiff   ChildDiff
	LicensesDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffSnapshot compares old and new snapshot states.
func DiffSnapshot(old, new *bronze.GCPComputeSnapshot) *SnapshotDiff {
	if old == nil {
		return &SnapshotDiff{
			IsNew:        true,
			LabelsDiff:   ChildDiff{Changed: true},
			LicensesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &SnapshotDiff{}
	diff.IsChanged = hasSnapshotFieldsChanged(old, new)
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)
	diff.LicensesDiff = diffLicenses(old.Licenses, new.Licenses)

	return diff
}

// HasAnyChange returns true if any part of the snapshot changed.
func (d *SnapshotDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed || d.LicensesDiff.Changed
}

func hasSnapshotFieldsChanged(old, new *bronze.GCPComputeSnapshot) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.DiskSizeGb != new.DiskSizeGb ||
		old.StorageBytes != new.StorageBytes ||
		old.StorageBytesStatus != new.StorageBytesStatus ||
		old.DownloadBytes != new.DownloadBytes ||
		old.SnapshotType != new.SnapshotType ||
		old.Architecture != new.Architecture ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.SourceDisk != new.SourceDisk ||
		old.SourceDiskId != new.SourceDiskId ||
		old.SourceDiskForRecoveryCheckpoint != new.SourceDiskForRecoveryCheckpoint ||
		old.AutoCreated != new.AutoCreated ||
		old.SatisfiesPzi != new.SatisfiesPzi ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.EnableConfidentialCompute != new.EnableConfidentialCompute ||
		jsonb.Changed(old.SnapshotEncryptionKeyJSON, new.SnapshotEncryptionKeyJSON) ||
		jsonb.Changed(old.SourceDiskEncryptionKeyJSON, new.SourceDiskEncryptionKeyJSON) ||
		jsonb.Changed(old.GuestOsFeaturesJSON, new.GuestOsFeaturesJSON) ||
		jsonb.Changed(old.StorageLocationsJSON, new.StorageLocationsJSON)
}

func diffLabels(old, new []bronze.GCPComputeSnapshotLabel) ChildDiff {
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

func diffLicenses(old, new []bronze.GCPComputeSnapshotLicense) ChildDiff {
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
