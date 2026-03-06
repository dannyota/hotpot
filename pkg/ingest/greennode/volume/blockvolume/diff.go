package blockvolume

import (
	"bytes"

	entvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
)

// BlockVolumeDiff represents changes between old and new block volume states.
type BlockVolumeDiff struct {
	IsNew     bool
	IsChanged bool

	SnapshotsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffBlockVolumeData compares old Ent entity and new BlockVolumeData.
func DiffBlockVolumeData(old *entvol.BronzeGreenNodeVolumeBlockVolume, new *BlockVolumeData) *BlockVolumeDiff {
	if old == nil {
		return &BlockVolumeDiff{
			IsNew:         true,
			SnapshotsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &BlockVolumeDiff{}
	diff.IsChanged = hasBlockVolumeFieldsChanged(old, new)
	diff.SnapshotsDiff = diffSnapshots(old.Edges.Snapshots, new.Snapshots)

	return diff
}

// HasAnyChange returns true if any part of the block volume changed.
func (d *BlockVolumeDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.SnapshotsDiff.Changed
}

func hasBlockVolumeFieldsChanged(old *entvol.BronzeGreenNodeVolumeBlockVolume, new *BlockVolumeData) bool {
	return old.Name != new.Name ||
		old.VolumeTypeID != new.VolumeTypeID ||
		old.ClusterID != new.ClusterID ||
		old.VMID != new.VMID ||
		old.Size != new.Size ||
		old.IopsID != new.IopsID ||
		old.Status != new.Status ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.UpdatedAtAPI != new.UpdatedAtAPI ||
		old.PersistentVolume != new.PersistentVolume ||
		!bytes.Equal(old.AttachedMachineJSON, new.AttachedMachineJSON) ||
		old.UnderID != new.UnderID ||
		old.MigrateState != new.MigrateState ||
		old.MultiAttach != new.MultiAttach ||
		old.ZoneID != new.ZoneID
}

func diffSnapshots(old []*entvol.BronzeGreenNodeVolumeSnapshot, new []SnapshotData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*entvol.BronzeGreenNodeVolumeSnapshot)
	for _, s := range old {
		oldMap[s.SnapshotID] = s
	}
	for _, s := range new {
		oldSnap, ok := oldMap[s.SnapshotID]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if oldSnap.Name != s.Name ||
			oldSnap.Size != s.Size ||
			oldSnap.VolumeSize != s.VolumeSize ||
			oldSnap.Status != s.Status ||
			oldSnap.CreatedAtAPI != s.CreatedAtAPI {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
