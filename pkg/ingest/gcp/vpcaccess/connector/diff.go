package connector

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// ConnectorDiff represents changes between old and new connector states.
type ConnectorDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffConnector compares old and new connector states.
func DiffConnector(old, new *bronze.GCPVpcAccessConnector) *ConnectorDiff {
	if old == nil {
		return &ConnectorDiff{IsNew: true}
	}
	return &ConnectorDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the connector changed.
func (d *ConnectorDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old, new *bronze.GCPVpcAccessConnector) bool {
	return old.Network != new.Network ||
		old.IpCidrRange != new.IpCidrRange ||
		old.State != new.State ||
		old.MinThroughput != new.MinThroughput ||
		old.MaxThroughput != new.MaxThroughput ||
		old.MinInstances != new.MinInstances ||
		old.MaxInstances != new.MaxInstances ||
		old.MachineType != new.MachineType ||
		old.Region != new.Region ||
		jsonb.Changed(old.SubnetJSON, new.SubnetJSON) ||
		jsonb.Changed(old.ConnectedProjectsJSON, new.ConnectedProjectsJSON)
}
