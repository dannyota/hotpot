package secgroup

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SecgroupDiff represents changes between old and new security group states.
type SecgroupDiff struct {
	IsNew     bool
	IsChanged bool

	RulesDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffSecgroupData compares old Ent entity and new SecgroupData.
func DiffSecgroupData(old *ent.BronzeGreenNodeNetworkSecgroup, new *SecgroupData) *SecgroupDiff {
	if old == nil {
		return &SecgroupDiff{
			IsNew:     true,
			RulesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &SecgroupDiff{}
	diff.IsChanged = hasSecgroupFieldsChanged(old, new)
	diff.RulesDiff = diffRules(old.Edges.Rules, new.Rules)

	return diff
}

// HasAnyChange returns true if any part of the security group changed.
func (d *SecgroupDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.RulesDiff.Changed
}

func hasSecgroupFieldsChanged(old *ent.BronzeGreenNodeNetworkSecgroup, new *SecgroupData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status
}

func diffRules(old []*ent.BronzeGreenNodeNetworkSecgroupRule, new []SecgroupRuleData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*ent.BronzeGreenNodeNetworkSecgroupRule)
	for _, r := range old {
		oldMap[r.RuleID] = r
	}
	for _, r := range new {
		oldRule, ok := oldMap[r.RuleID]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if oldRule.Direction != r.Direction ||
			oldRule.EtherType != r.EtherType ||
			oldRule.Protocol != r.Protocol ||
			oldRule.Description != r.Description ||
			oldRule.RemoteIPPrefix != r.RemoteIPPrefix ||
			oldRule.PortRangeMax != r.PortRangeMax ||
			oldRule.PortRangeMin != r.PortRangeMin {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
