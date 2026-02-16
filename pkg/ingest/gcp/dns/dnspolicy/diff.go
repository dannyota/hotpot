package dnspolicy

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// PolicyDiff represents changes between old and new DNS policy states.
type PolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffPolicyData compares old Ent entity and new data.
func DiffPolicyData(old *ent.BronzeGCPDNSPolicy, new *PolicyData) *PolicyDiff {
	if old == nil {
		return &PolicyDiff{IsNew: true}
	}
	return &PolicyDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the policy changed.
func (d *PolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old *ent.BronzeGCPDNSPolicy, new *PolicyData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.EnableInboundForwarding != new.EnableInboundForwarding ||
		old.EnableLogging != new.EnableLogging ||
		!bytes.Equal(old.NetworksJSON, new.NetworksJSON) ||
		!bytes.Equal(old.AlternativeNameServerConfigJSON, new.AlternativeNameServerConfigJSON)
}
