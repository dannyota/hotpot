package instancegroup

import (
	"hotpot/pkg/storage/ent"
)

// InstanceGroupDiff represents changes between old and new instance group states.
type InstanceGroupDiff struct {
	IsNew          bool
	IsChanged      bool
	NamedPortsDiff ChildDiff
	MembersDiff    ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if any part of the instance group changed.
func (d *InstanceGroupDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.NamedPortsDiff.HasChanges || d.MembersDiff.HasChanges
}

// DiffInstanceGroupData compares existing Ent entity with new InstanceGroupData.
func DiffInstanceGroupData(old *ent.BronzeGCPComputeInstanceGroup, new *InstanceGroupData) *InstanceGroupDiff {
	diff := &InstanceGroupDiff{}

	// New instance group
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	diff.IsChanged = hasGroupFieldsChanged(old, new)

	// Compare children
	diff.NamedPortsDiff = diffNamedPortsData(old.Edges.NamedPorts, new.NamedPorts)
	diff.MembersDiff = diffMembersData(old.Edges.Members, new.Members)

	return diff
}

// hasGroupFieldsChanged compares group-level fields (excluding children).
func hasGroupFieldsChanged(old *ent.BronzeGCPComputeInstanceGroup, new *InstanceGroupData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Zone != new.Zone ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.Size != new.Size ||
		old.Fingerprint != new.Fingerprint
}

func diffNamedPortsData(old []*ent.BronzeGCPComputeInstanceGroupNamedPort, new []NamedPortData) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old named ports
	oldMap := make(map[string]int32, len(old))
	for _, p := range old {
		oldMap[p.Name] = p.Port
	}

	// Compare with new named ports
	for _, p := range new {
		if oldPort, ok := oldMap[p.Name]; !ok || oldPort != p.Port {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}

func diffMembersData(old []*ent.BronzeGCPComputeInstanceGroupMember, new []MemberData) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old members
	oldMap := make(map[string]string, len(old))
	for _, m := range old {
		oldMap[m.InstanceURL] = m.Status
	}

	// Compare with new members
	for _, m := range new {
		if oldStatus, ok := oldMap[m.InstanceURL]; !ok || oldStatus != m.Status {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}
