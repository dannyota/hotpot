package instancegroup

import (
	"hotpot/pkg/base/models/bronze"
)

// InstanceGroupDiff represents changes between old and new instance group states.
type InstanceGroupDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	NamedPortsDiff ChildDiff
	MembersDiff    ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffInstanceGroup compares old and new instance group states.
// Returns nil if old is nil (new instance group).
func DiffInstanceGroup(old, new *bronze.GCPComputeInstanceGroup) *InstanceGroupDiff {
	if old == nil {
		return &InstanceGroupDiff{
			IsNew:          true,
			NamedPortsDiff: ChildDiff{Changed: true},
			MembersDiff:    ChildDiff{Changed: true},
		}
	}

	diff := &InstanceGroupDiff{}

	// Compare group-level fields
	diff.IsChanged = hasGroupFieldsChanged(old, new)

	// Compare children
	diff.NamedPortsDiff = diffNamedPorts(old.NamedPorts, new.NamedPorts)
	diff.MembersDiff = diffMembers(old.Members, new.Members)

	return diff
}

// HasAnyChange returns true if any part of the instance group changed.
func (d *InstanceGroupDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.NamedPortsDiff.Changed || d.MembersDiff.Changed
}

// hasGroupFieldsChanged compares group-level fields (excluding children).
func hasGroupFieldsChanged(old, new *bronze.GCPComputeInstanceGroup) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Zone != new.Zone ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.Size != new.Size ||
		old.Fingerprint != new.Fingerprint
}

func diffNamedPorts(old, new []bronze.GCPComputeInstanceGroupNamedPort) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]int32)
	for _, p := range old {
		oldMap[p.Name] = p.Port
	}
	for _, p := range new {
		if port, ok := oldMap[p.Name]; !ok || port != p.Port {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffMembers(old, new []bronze.GCPComputeInstanceGroupMember) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]string)
	for _, m := range old {
		oldMap[m.InstanceURL] = m.Status
	}
	for _, m := range new {
		if status, ok := oldMap[m.InstanceURL]; !ok || status != m.Status {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
