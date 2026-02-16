package packetmirroring

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// PacketMirroringDiff represents changes between old and new packet mirroring states.
type PacketMirroringDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffPacketMirroringData compares existing Ent entity with new PacketMirroringData and returns differences.
func DiffPacketMirroringData(old *ent.BronzeGCPComputePacketMirroring, new *PacketMirroringData) *PacketMirroringDiff {
	diff := &PacketMirroringDiff{}

	// New packet mirroring
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.SelfLink != new.SelfLink ||
		old.Region != new.Region ||
		old.Network != new.Network ||
		old.Priority != new.Priority ||
		old.Enable != new.Enable ||
		old.CreationTimestamp != new.CreationTimestamp ||
		!bytes.Equal(old.CollectorIlbJSON, new.CollectorIlbJSON) ||
		!bytes.Equal(old.MirroredResourcesJSON, new.MirroredResourcesJSON) ||
		!bytes.Equal(old.FilterJSON, new.FilterJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the packet mirroring changed.
func (d *PacketMirroringDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
