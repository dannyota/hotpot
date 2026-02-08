package instance

import (
	"bytes"
	"hotpot/pkg/storage/ent"
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

// DiffInstanceData compares old Ent entity and new data.
func DiffInstanceData(old *ent.BronzeGCPComputeInstance, new *InstanceData) *InstanceDiff {
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

	// Compare children (need to load edges from old)
	diff.DisksDiff = diffDisksData(old.Edges.Disks, new.Disks)
	diff.NICsDiff = diffNICsData(old.Edges.Nics, new.NICs)
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)
	diff.TagsDiff = diffTagsData(old.Edges.Tags, new.Tags)
	diff.MetadataDiff = diffMetadataData(old.Edges.Metadata, new.Metadata)
	diff.ServiceAccountsDiff = diffServiceAccountsData(old.Edges.ServiceAccounts, new.ServiceAccounts)

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
func hasInstanceFieldsChanged(old *ent.BronzeGCPComputeInstance, new *InstanceData) bool {
	return old.Name != new.Name ||
		old.Zone != new.Zone ||
		old.MachineType != new.MachineType ||
		old.Status != new.Status ||
		old.StatusMessage != new.StatusMessage ||
		old.CPUPlatform != new.CpuPlatform ||
		old.Hostname != new.Hostname ||
		old.Description != new.Description ||
		old.DeletionProtection != new.DeletionProtection ||
		old.CanIPForward != new.CanIpForward ||
		!bytes.Equal(old.SchedulingJSON, new.SchedulingJSON)
}

func diffDisksData(old []*ent.BronzeGCPComputeInstanceDisk, new []DiskData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if hasDiskChangedData(old[i], &new[i]) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasDiskChangedData(old *ent.BronzeGCPComputeInstanceDisk, new *DiskData) bool {
	// Compare fields
	if old.Source != new.Source ||
		old.DeviceName != new.DeviceName ||
		old.Index != new.Index ||
		old.Boot != new.Boot ||
		old.AutoDelete != new.AutoDelete ||
		old.Mode != new.Mode ||
		old.Interface != new.Interface ||
		old.Type != new.Type ||
		old.DiskSizeGB != new.DiskSizeGb ||
		!bytes.Equal(old.DiskEncryptionKeyJSON, new.DiskEncryptionKeyJSON) ||
		!bytes.Equal(old.InitializeParamsJSON, new.InitializeParamsJSON) {
		return true
	}

	// Compare licenses
	if len(old.Edges.Licenses) != len(new.Licenses) {
		return true
	}
	for i := range old.Edges.Licenses {
		if old.Edges.Licenses[i].License != new.Licenses[i].License {
			return true
		}
	}
	return false
}

func diffNICsData(old []*ent.BronzeGCPComputeInstanceNIC, new []NICData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if hasNICChangedData(old[i], &new[i]) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasNICChangedData(old *ent.BronzeGCPComputeInstanceNIC, new *NICData) bool {
	if old.Name != new.Name ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.NetworkIP != new.NetworkIP ||
		old.StackType != new.StackType ||
		old.NicType != new.NicType {
		return true
	}

	// Compare access configs
	if len(old.Edges.AccessConfigs) != len(new.AccessConfigs) {
		return true
	}
	for i := range old.Edges.AccessConfigs {
		if old.Edges.AccessConfigs[i].Type != new.AccessConfigs[i].Type ||
			old.Edges.AccessConfigs[i].Name != new.AccessConfigs[i].Name ||
			old.Edges.AccessConfigs[i].NatIP != new.AccessConfigs[i].NatIP ||
			old.Edges.AccessConfigs[i].NetworkTier != new.AccessConfigs[i].NetworkTier {
			return true
		}
	}

	// Compare alias ranges
	if len(old.Edges.AliasIPRanges) != len(new.AliasIPRanges) {
		return true
	}
	for i := range old.Edges.AliasIPRanges {
		if old.Edges.AliasIPRanges[i].IPCidrRange != new.AliasIPRanges[i].IPCidrRange ||
			old.Edges.AliasIPRanges[i].SubnetworkRangeName != new.AliasIPRanges[i].SubnetworkRangeName {
			return true
		}
	}

	return false
}

func diffLabelsData(old []*ent.BronzeGCPComputeInstanceLabel, new []LabelData) ChildDiff {
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

func diffTagsData(old []*ent.BronzeGCPComputeInstanceTag, new []TagData) ChildDiff {
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

func diffMetadataData(old []*ent.BronzeGCPComputeInstanceMetadata, new []MetadataData) ChildDiff {
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

func diffServiceAccountsData(old []*ent.BronzeGCPComputeInstanceServiceAccount, new []ServiceAccountData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if old[i].Email != new[i].Email || !bytes.Equal(old[i].ScopesJSON, new[i].ScopesJSON) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
