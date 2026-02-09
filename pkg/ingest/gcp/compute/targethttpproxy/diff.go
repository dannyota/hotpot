package targethttpproxy

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// TargetHttpProxyDiff represents changes between old and new target HTTP proxy states.
type TargetHttpProxyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetHttpProxyData compares existing Ent entity with new TargetHttpProxyData and returns differences.
func DiffTargetHttpProxyData(old *ent.BronzeGCPComputeTargetHttpProxy, new *TargetHttpProxyData) *TargetHttpProxyDiff {
	diff := &TargetHttpProxyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Fingerprint != new.Fingerprint ||
		old.URLMap != new.UrlMap ||
		old.ProxyBind != new.ProxyBind ||
		old.HTTPKeepAliveTimeoutSec != new.HttpKeepAliveTimeoutSec ||
		old.Region != new.Region {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the target HTTP proxy changed.
func (d *TargetHttpProxyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
