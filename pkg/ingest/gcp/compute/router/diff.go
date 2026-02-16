package router

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// RouterDiff represents changes between old and new router states.
type RouterDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffRouterData compares existing Ent entity with new RouterData and returns differences.
func DiffRouterData(old *ent.BronzeGCPComputeRouter, new *RouterData) *RouterDiff {
	diff := &RouterDiff{}

	// New router
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.Network != new.Network ||
		old.Region != new.Region ||
		old.BgpAsn != new.BgpAsn ||
		old.BgpAdvertiseMode != new.BgpAdvertiseMode ||
		old.BgpKeepaliveInterval != new.BgpKeepaliveInterval ||
		old.EncryptedInterconnectRouter != new.EncryptedInterconnectRouter ||
		!bytes.Equal(old.BgpAdvertisedGroupsJSON, new.BgpAdvertisedGroupsJSON) ||
		!bytes.Equal(old.BgpAdvertisedIPRangesJSON, new.BgpAdvertisedIPRangesJSON) ||
		!bytes.Equal(old.BgpPeersJSON, new.BgpPeersJSON) ||
		!bytes.Equal(old.InterfacesJSON, new.InterfacesJSON) ||
		!bytes.Equal(old.NatsJSON, new.NatsJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the router changed.
func (d *RouterDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
