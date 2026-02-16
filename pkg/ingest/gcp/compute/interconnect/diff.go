package interconnect

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InterconnectDiff represents changes between old and new interconnect states.
type InterconnectDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffInterconnectData compares existing Ent entity with new InterconnectData and returns differences.
func DiffInterconnectData(old *ent.BronzeGCPComputeInterconnect, new *InterconnectData) *InterconnectDiff {
	diff := &InterconnectDiff{}

	// New interconnect
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.SelfLink != new.SelfLink ||
		old.Location != new.Location ||
		old.InterconnectType != new.InterconnectType ||
		old.LinkType != new.LinkType ||
		old.AdminEnabled != new.AdminEnabled ||
		old.OperationalStatus != new.OperationalStatus ||
		old.ProvisionedLinkCount != new.ProvisionedLinkCount ||
		old.RequestedLinkCount != new.RequestedLinkCount ||
		old.PeerIPAddress != new.PeerIPAddress ||
		old.GoogleIPAddress != new.GoogleIPAddress ||
		old.GoogleReferenceID != new.GoogleReferenceID ||
		old.NocContactEmail != new.NocContactEmail ||
		old.CustomerName != new.CustomerName ||
		old.State != new.State ||
		old.CreationTimestamp != new.CreationTimestamp ||
		!bytes.Equal(old.ExpectedOutagesJSON, new.ExpectedOutagesJSON) ||
		!bytes.Equal(old.CircuitInfosJSON, new.CircuitInfosJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the interconnect changed.
func (d *InterconnectDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
