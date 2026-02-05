package instance

import (
	"reflect"

	"hotpot/pkg/base/models/bronze"
)

// InstanceDiff represents changes between old and new instance states.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	DisksDiff           ChildDiff
	NICsDiff            ChildDiff
	LabelsDiff          ChildDiff
	TagsDiff            ChildDiff
	MetadataDiff        ChildDiff
	ServiceAccountsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffInstance compares old and new instance states.
// Returns nil if old is nil (new instance).
func DiffInstance(old, new *bronze.GCPComputeInstance) *InstanceDiff {
	if old == nil {
		return &InstanceDiff{
			IsNew:               true,
			DisksDiff:           ChildDiff{Changed: true},
			NICsDiff:            ChildDiff{Changed: true},
			LabelsDiff:          ChildDiff{Changed: true},
			TagsDiff:            ChildDiff{Changed: true},
			MetadataDiff:        ChildDiff{Changed: true},
			ServiceAccountsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &InstanceDiff{}

	// Compare instance-level fields
	diff.IsChanged = hasInstanceFieldsChanged(old, new)

	// Compare children
	diff.DisksDiff = diffDisks(old.Disks, new.Disks)
	diff.NICsDiff = diffNICs(old.NICs, new.NICs)
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)
	diff.TagsDiff = diffTags(old.Tags, new.Tags)
	diff.MetadataDiff = diffMetadata(old.Metadata, new.Metadata)
	diff.ServiceAccountsDiff = diffServiceAccounts(old.ServiceAccounts, new.ServiceAccounts)

	return diff
}

// HasAnyChange returns true if any part of the instance changed.
func (d *InstanceDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.DisksDiff.Changed ||
		d.NICsDiff.Changed ||
		d.LabelsDiff.Changed ||
		d.TagsDiff.Changed ||
		d.MetadataDiff.Changed ||
		d.ServiceAccountsDiff.Changed
}

// hasInstanceFieldsChanged compares instance-level fields (excluding children).
func hasInstanceFieldsChanged(old, new *bronze.GCPComputeInstance) bool {
	return old.Name != new.Name ||
		old.Zone != new.Zone ||
		old.MachineType != new.MachineType ||
		old.Status != new.Status ||
		old.StatusMessage != new.StatusMessage ||
		old.CpuPlatform != new.CpuPlatform ||
		old.Hostname != new.Hostname ||
		old.Description != new.Description ||
		old.DeletionProtection != new.DeletionProtection ||
		old.CanIpForward != new.CanIpForward ||
		old.SchedulingJSON != new.SchedulingJSON
}

func diffDisks(old, new []bronze.GCPComputeInstanceDisk) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if hasDiskChanged(&old[i], &new[i]) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasDiskChanged(old, new *bronze.GCPComputeInstanceDisk) bool {
	return old.Source != new.Source ||
		old.DeviceName != new.DeviceName ||
		old.Index != new.Index ||
		old.Boot != new.Boot ||
		old.AutoDelete != new.AutoDelete ||
		old.Mode != new.Mode ||
		old.Interface != new.Interface ||
		old.Type != new.Type ||
		old.DiskSizeGb != new.DiskSizeGb ||
		old.DiskEncryptionKeyJSON != new.DiskEncryptionKeyJSON ||
		old.InitializeParamsJSON != new.InitializeParamsJSON ||
		!reflect.DeepEqual(old.Licenses, new.Licenses)
}

func diffNICs(old, new []bronze.GCPComputeInstanceNIC) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if hasNICChanged(&old[i], &new[i]) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasNICChanged(old, new *bronze.GCPComputeInstanceNIC) bool {
	return old.Name != new.Name ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.NetworkIP != new.NetworkIP ||
		old.StackType != new.StackType ||
		old.NicType != new.NicType ||
		!reflect.DeepEqual(old.AccessConfigs, new.AccessConfigs) ||
		!reflect.DeepEqual(old.AliasIpRanges, new.AliasIpRanges)
}

func diffLabels(old, new []bronze.GCPComputeInstanceLabel) ChildDiff {
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

func diffTags(old, new []bronze.GCPComputeInstanceTag) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldSet := make(map[string]bool)
	for _, t := range old {
		oldSet[t.Tag] = true
	}
	for _, t := range new {
		if !oldSet[t.Tag] {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffMetadata(old, new []bronze.GCPComputeInstanceMetadata) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]string)
	for _, m := range old {
		oldMap[m.Key] = m.Value
	}
	for _, m := range new {
		if v, ok := oldMap[m.Key]; !ok || v != m.Value {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffServiceAccounts(old, new []bronze.GCPComputeInstanceServiceAccount) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if old[i].Email != new[i].Email || old[i].ScopesJSON != new[i].ScopesJSON {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
