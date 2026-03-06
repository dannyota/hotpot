package lb

import (
	"bytes"

	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// LBDiff represents changes between old and new load balancer states.
type LBDiff struct {
	IsNew     bool
	IsChanged bool

	ListenersDiff ChildDiff
	PoolsDiff     ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffLBData compares old Ent entity and new LBData.
func DiffLBData(old *entlb.BronzeGreenNodeLoadBalancerLB, new *LBData) *LBDiff {
	if old == nil {
		return &LBDiff{
			IsNew:         true,
			ListenersDiff: ChildDiff{Changed: true},
			PoolsDiff:     ChildDiff{Changed: true},
		}
	}

	diff := &LBDiff{}
	diff.IsChanged = hasLBFieldsChanged(old, new)
	diff.ListenersDiff = diffListeners(old.Edges.Listeners, new.Listeners)
	diff.PoolsDiff = diffPools(old.Edges.Pools, new.Pools)

	return diff
}

// HasAnyChange returns true if any part of the load balancer changed.
func (d *LBDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.ListenersDiff.Changed || d.PoolsDiff.Changed
}

func hasLBFieldsChanged(old *entlb.BronzeGreenNodeLoadBalancerLB, new *LBData) bool {
	return old.Name != new.Name ||
		old.DisplayStatus != new.DisplayStatus ||
		old.Address != new.Address ||
		old.PrivateSubnetID != new.PrivateSubnetID ||
		old.PrivateSubnetCidr != new.PrivateSubnetCidr ||
		old.Type != new.Type ||
		old.DisplayType != new.DisplayType ||
		old.LoadBalancerSchema != new.LoadBalancerSchema ||
		old.PackageID != new.PackageID ||
		old.Description != new.Description ||
		old.Location != new.Location ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.UpdatedAtAPI != new.UpdatedAtAPI ||
		old.ProgressStatus != new.ProgressStatus ||
		old.Status != new.Status ||
		old.BackendSubnetID != new.BackendSubnetID ||
		old.Internal != new.Internal ||
		old.AutoScalable != new.AutoScalable ||
		old.ZoneID != new.ZoneID ||
		old.MinSize != new.MinSize ||
		old.MaxSize != new.MaxSize ||
		old.TotalNodes != new.TotalNodes ||
		!bytes.Equal(old.NodesJSON, new.NodesJSON)
}

func diffListeners(old []*entlb.BronzeGreenNodeLoadBalancerListener, new []ListenerData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*entlb.BronzeGreenNodeLoadBalancerListener)
	for _, l := range old {
		oldMap[l.ListenerID] = l
	}
	for _, l := range new {
		existing, ok := oldMap[l.ListenerID]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if hasListenerFieldsChanged(existing, &l) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasListenerFieldsChanged(old *entlb.BronzeGreenNodeLoadBalancerListener, new *ListenerData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Protocol != new.Protocol ||
		old.ProtocolPort != new.ProtocolPort ||
		old.ConnectionLimit != new.ConnectionLimit ||
		old.DefaultPoolID != new.DefaultPoolID ||
		old.DefaultPoolName != new.DefaultPoolName ||
		old.TimeoutClient != new.TimeoutClient ||
		old.TimeoutMember != new.TimeoutMember ||
		old.TimeoutConnection != new.TimeoutConnection ||
		old.AllowedCidrs != new.AllowedCidrs ||
		!bytes.Equal(old.CertificateAuthoritiesJSON, new.CertificateAuthoritiesJSON) ||
		old.DisplayStatus != new.DisplayStatus ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.UpdatedAtAPI != new.UpdatedAtAPI ||
		!optionalStringEqual(old.DefaultCertificateAuthority, new.DefaultCertificateAuthority) ||
		!optionalStringEqual(old.ClientCertificateAuthentication, new.ClientCertificateAuthentication) ||
		old.ProgressStatus != new.ProgressStatus ||
		!bytes.Equal(old.InsertHeadersJSON, new.InsertHeadersJSON) ||
		!bytes.Equal(old.PoliciesJSON, new.PoliciesJSON)
}

func diffPools(old []*entlb.BronzeGreenNodeLoadBalancerPool, new []PoolData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*entlb.BronzeGreenNodeLoadBalancerPool)
	for _, p := range old {
		oldMap[p.PoolID] = p
	}
	for _, p := range new {
		existing, ok := oldMap[p.PoolID]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if hasPoolFieldsChanged(existing, &p) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasPoolFieldsChanged(old *entlb.BronzeGreenNodeLoadBalancerPool, new *PoolData) bool {
	return old.Name != new.Name ||
		old.Protocol != new.Protocol ||
		old.Description != new.Description ||
		old.LoadBalanceMethod != new.LoadBalanceMethod ||
		old.Status != new.Status ||
		old.Stickiness != new.Stickiness ||
		old.TLSEncryption != new.TLSEncryption ||
		!bytes.Equal(old.MembersJSON, new.MembersJSON) ||
		!bytes.Equal(old.HealthMonitorJSON, new.HealthMonitorJSON)
}

func optionalStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
