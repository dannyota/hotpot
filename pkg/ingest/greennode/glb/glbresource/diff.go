package glbresource

import (
	"bytes"

	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
)

// GLBDiff represents changes between old and new GLB states.
type GLBDiff struct {
	IsNew     bool
	IsChanged bool

	ListenersDiff ChildDiff
	PoolsDiff     ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffGLBData compares old Ent entity and new GLBData.
func DiffGLBData(old *entglb.BronzeGreenNodeGLBGlobalLoadBalancer, new *GLBData) *GLBDiff {
	if old == nil {
		return &GLBDiff{
			IsNew:         true,
			ListenersDiff: ChildDiff{Changed: true},
			PoolsDiff:     ChildDiff{Changed: true},
		}
	}

	diff := &GLBDiff{}
	diff.IsChanged = hasGLBFieldsChanged(old, new)
	diff.ListenersDiff = diffListeners(old.Edges.Listeners, new.Listeners)
	diff.PoolsDiff = diffPools(old.Edges.Pools, new.Pools)

	return diff
}

// HasAnyChange returns true if any part of the GLB changed.
func (d *GLBDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.ListenersDiff.Changed || d.PoolsDiff.Changed
}

func hasGLBFieldsChanged(old *entglb.BronzeGreenNodeGLBGlobalLoadBalancer, new *GLBData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.Package != new.Package ||
		old.Type != new.Type ||
		old.UserID != new.UserID ||
		!bytes.Equal(old.VipsJSON, new.VipsJSON) ||
		!bytes.Equal(old.DomainsJSON, new.DomainsJSON) ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.UpdatedAtAPI != new.UpdatedAtAPI ||
		old.DeletedAtAPI != new.DeletedAtAPI
}

func diffListeners(old []*entglb.BronzeGreenNodeGLBGlobalListener, new []GLBListenerData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*entglb.BronzeGreenNodeGLBGlobalListener)
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

func hasListenerFieldsChanged(old *entglb.BronzeGreenNodeGLBGlobalListener, new *GLBListenerData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Protocol != new.Protocol ||
		old.Port != new.Port ||
		old.GlobalPoolID != new.GlobalPoolID ||
		old.TimeoutClient != new.TimeoutClient ||
		old.TimeoutMember != new.TimeoutMember ||
		old.TimeoutConnection != new.TimeoutConnection ||
		old.AllowedCidrs != new.AllowedCidrs ||
		!ptrStringEqual(old.Headers, new.Headers) ||
		old.Status != new.Status ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.UpdatedAtAPI != new.UpdatedAtAPI ||
		!ptrStringEqual(old.DeletedAtAPI, new.DeletedAtAPI)
}

func diffPools(old []*entglb.BronzeGreenNodeGLBGlobalPool, new []GLBPoolData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*entglb.BronzeGreenNodeGLBGlobalPool)
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

func hasPoolFieldsChanged(old *entglb.BronzeGreenNodeGLBGlobalPool, new *GLBPoolData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Algorithm != new.Algorithm ||
		!ptrStringEqual(old.StickySession, new.StickySession) ||
		!ptrStringEqual(old.TLSEnabled, new.TLSEnabled) ||
		old.Protocol != new.Protocol ||
		old.Status != new.Status ||
		!bytes.Equal(old.HealthJSON, new.HealthJSON) ||
		!bytes.Equal(old.PoolMembersJSON, new.PoolMembersJSON) ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.UpdatedAtAPI != new.UpdatedAtAPI ||
		!ptrStringEqual(old.DeletedAtAPI, new.DeletedAtAPI)
}

func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
