package instance

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InstanceDiff represents changes between old and new instance states.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffInstanceData compares old Ent entity and new data.
func DiffInstanceData(old *ent.BronzeGCPSQLInstance, new *InstanceData) *InstanceDiff {
	if old == nil {
		return &InstanceDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &InstanceDiff{}

	// Compare instance-level fields
	diff.IsChanged = hasInstanceFieldsChanged(old, new)

	// Compare labels
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the instance changed.
func (d *InstanceDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

// hasInstanceFieldsChanged compares instance-level fields (excluding children).
func hasInstanceFieldsChanged(old *ent.BronzeGCPSQLInstance, new *InstanceData) bool {
	return old.Name != new.Name ||
		old.DatabaseVersion != new.DatabaseVersion ||
		old.State != new.State ||
		old.Region != new.Region ||
		old.GceZone != new.GceZone ||
		old.SecondaryGceZone != new.SecondaryGceZone ||
		old.InstanceType != new.InstanceType ||
		old.ConnectionName != new.ConnectionName ||
		old.ServiceAccountEmailAddress != new.ServiceAccountEmailAddress ||
		old.SelfLink != new.SelfLink ||
		!bytes.Equal(old.SettingsJSON, new.SettingsJSON) ||
		!bytes.Equal(old.ServerCaCertJSON, new.ServerCaCertJSON) ||
		!bytes.Equal(old.IPAddressesJSON, new.IpAddressesJSON) ||
		!bytes.Equal(old.ReplicaConfigurationJSON, new.ReplicaConfigurationJSON) ||
		!bytes.Equal(old.FailoverReplicaJSON, new.FailoverReplicaJSON) ||
		!bytes.Equal(old.DiskEncryptionConfigurationJSON, new.DiskEncryptionConfigurationJSON) ||
		!bytes.Equal(old.DiskEncryptionStatusJSON, new.DiskEncryptionStatusJSON)
}

func diffLabelsData(old []*ent.BronzeGCPSQLInstanceLabel, new []LabelData) ChildDiff {
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
