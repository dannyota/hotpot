package urlmap

import (
	"reflect"

	"hotpot/pkg/storage/ent"
)

// UrlMapDiff represents changes between old and new URL map states.
type UrlMapDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffUrlMapData compares existing Ent entity with new UrlMapData.
func DiffUrlMapData(old *ent.BronzeGCPComputeUrlMap, new *UrlMapData) *UrlMapDiff {
	diff := &UrlMapDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Fingerprint != new.Fingerprint ||
		old.DefaultService != new.DefaultService ||
		old.Region != new.Region ||
		!reflect.DeepEqual(old.HostRulesJSON, new.HostRulesJSON) ||
		!reflect.DeepEqual(old.PathMatchersJSON, new.PathMatchersJSON) ||
		!reflect.DeepEqual(old.TestsJSON, new.TestsJSON) ||
		!reflect.DeepEqual(old.DefaultRouteActionJSON, new.DefaultRouteActionJSON) ||
		!reflect.DeepEqual(old.DefaultURLRedirectJSON, new.DefaultUrlRedirectJSON) ||
		!reflect.DeepEqual(old.HeaderActionJSON, new.HeaderActionJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the URL map changed.
func (d *UrlMapDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
