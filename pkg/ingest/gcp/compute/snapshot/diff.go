package snapshot

import (
	"bytes"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SnapshotDiff represents changes between old and new snapshot states.
type SnapshotDiff struct {
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

// DiffSnapshotData compares old Ent entity and new SnapshotData.
func DiffSnapshotData(old *ent.BronzeGCPComputeSnapshot, new *SnapshotData) *SnapshotDiff {
	if old == nil {
		return &SnapshotDiff{
			IsNew:        true,
			LabelsDiff:   ChildDiff{Changed: true},
			LicensesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &SnapshotDiff{}

	// Compare snapshot-level fields
	diff.IsChanged = hasSnapshotFieldsChanged(old, new)

	// Compare children
	diff.LabelsDiff = diffLabels(old.Edges.Labels, new.Labels)
	diff.LicensesDiff = diffLicenses(old.Edges.Licenses, new.Licenses)

	return diff
}

// HasAnyChange returns true if any part of the snapshot changed.
func (d *SnapshotDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed || d.LicensesDiff.Changed
}

// hasSnapshotFieldsChanged compares snapshot-level fields (excluding children).
func hasSnapshotFieldsChanged(old *ent.BronzeGCPComputeSnapshot, new *SnapshotData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.DiskSizeGB != new.DiskSizeGB ||
		old.StorageBytes != new.StorageBytes ||
		old.StorageBytesStatus != new.StorageBytesStatus ||
		old.DownloadBytes != new.DownloadBytes ||
		old.SnapshotType != new.SnapshotType ||
		old.Architecture != new.Architecture ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.SourceDisk != new.SourceDisk ||
		old.SourceDiskID != new.SourceDiskID ||
		old.SourceDiskForRecoveryCheckpoint != new.SourceDiskForRecoveryCheckpoint ||
		old.AutoCreated != new.AutoCreated ||
		old.SatisfiesPzi != new.SatisfiesPzi ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.EnableConfidentialCompute != new.EnableConfidentialCompute ||
		!bytes.Equal(old.SnapshotEncryptionKeyJSON, new.SnapshotEncryptionKeyJSON) ||
		!bytes.Equal(old.SourceDiskEncryptionKeyJSON, new.SourceDiskEncryptionKeyJSON) ||
		!bytes.Equal(old.GuestOsFeaturesJSON, new.GuestOsFeaturesJSON) ||
		!bytes.Equal(old.StorageLocationsJSON, new.StorageLocationsJSON)
}

func diffLabels(old []*ent.BronzeGCPComputeSnapshotLabel, new []SnapshotLabelData) ChildDiff {
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

func diffLicenses(old []*ent.BronzeGCPComputeSnapshotLicense, new []SnapshotLicenseData) ChildDiff {
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
