package connector

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ConnectorDiff represents changes between old and new connector states.
type ConnectorDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffConnectorData compares existing Ent entity with new ConnectorData and returns differences.
func DiffConnectorData(old *ent.BronzeGCPVPCAccessConnector, new *ConnectorData) *ConnectorDiff {
	diff := &ConnectorDiff{}

	// New connector
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Network != new.Network ||
		old.IPCidrRange != new.IpCidrRange ||
		old.State != new.State ||
		old.MinThroughput != new.MinThroughput ||
		old.MaxThroughput != new.MaxThroughput ||
		old.MinInstances != new.MinInstances ||
		old.MaxInstances != new.MaxInstances ||
		old.MachineType != new.MachineType ||
		old.Region != new.Region ||
		!bytes.Equal(old.SubnetJSON, new.SubnetJSON) ||
		!bytes.Equal(old.ConnectedProjectsJSON, new.ConnectedProjectsJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the connector changed.
func (d *ConnectorDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
