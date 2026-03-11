package firewall

import (
	"bytes"

	entdo "danny.vn/hotpot/pkg/storage/ent/do"
)

// FirewallDiff represents changes between old and new Firewall states.
type FirewallDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffFirewallData compares old Ent entity and new data.
func DiffFirewallData(old *entdo.BronzeDOFirewall, new *FirewallData) *FirewallDiff {
	if old == nil {
		return &FirewallDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.Status != new.Status ||
		old.APICreatedAt != new.APICreatedAt ||
		!bytes.Equal(old.InboundRulesJSON, new.InboundRulesJSON) ||
		!bytes.Equal(old.OutboundRulesJSON, new.OutboundRulesJSON) ||
		!bytes.Equal(old.DropletIdsJSON, new.DropletIdsJSON) ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!bytes.Equal(old.PendingChangesJSON, new.PendingChangesJSON)

	return &FirewallDiff{IsChanged: changed}
}
