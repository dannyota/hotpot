package targettcpproxy

import (
	"hotpot/pkg/storage/ent"
)

// TargetTcpProxyDiff represents changes between old and new target TCP proxy states.
type TargetTcpProxyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetTcpProxyData compares existing Ent entity with new TargetTcpProxyData and returns differences.
func DiffTargetTcpProxyData(old *ent.BronzeGCPComputeTargetTcpProxy, new *TargetTcpProxyData) *TargetTcpProxyDiff {
	diff := &TargetTcpProxyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Service != new.Service ||
		old.ProxyBind != new.ProxyBind ||
		old.ProxyHeader != new.ProxyHeader ||
		old.Region != new.Region {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the target TCP proxy changed.
func (d *TargetTcpProxyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
