package servergroup

import (
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
)

// ServerGroupDiff represents changes between old and new server group states.
type ServerGroupDiff struct {
	IsNew     bool
	IsChanged bool

	MembersDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffServerGroupData compares old Ent entity and new ServerGroupData.
func DiffServerGroupData(old *entcompute.BronzeGreenNodeComputeServerGroup, new *ServerGroupData) *ServerGroupDiff {
	if old == nil {
		return &ServerGroupDiff{
			IsNew:       true,
			MembersDiff: ChildDiff{Changed: true},
		}
	}

	diff := &ServerGroupDiff{}
	diff.IsChanged = old.Name != new.Name ||
		old.Description != new.Description ||
		old.PolicyID != new.PolicyID ||
		old.PolicyName != new.PolicyName

	diff.MembersDiff = diffMembers(old.Edges.Members, new.Members)

	return diff
}

// HasAnyChange returns true if any part of the server group changed.
func (d *ServerGroupDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.MembersDiff.Changed
}

func diffMembers(old []*entcompute.BronzeGreenNodeComputeServerGroupMember, new []MemberData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]string)
	for _, m := range old {
		oldMap[m.UUID] = m.Name
	}
	for _, m := range new {
		if name, ok := oldMap[m.UUID]; !ok || name != m.Name {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
