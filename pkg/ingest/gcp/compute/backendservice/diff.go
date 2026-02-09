package backendservice

import (
	"encoding/json"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// BackendServiceDiff represents changes between old and new backend service states.
type BackendServiceDiff struct {
	IsNew        bool
	IsChanged    bool
	BackendsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if any part of the backend service changed.
func (d *BackendServiceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.BackendsDiff.HasChanges
}

// DiffBackendServiceData compares existing Ent entity with new BackendServiceData.
func DiffBackendServiceData(old *ent.BronzeGCPComputeBackendService, new *BackendServiceData) *BackendServiceDiff {
	diff := &BackendServiceDiff{}

	// New backend service
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	diff.IsChanged = hasBackendServiceFieldsChanged(old, new)

	// Compare backends
	diff.BackendsDiff = diffBackendsData(old.Edges.Backends, new.Backends)

	return diff
}

// hasBackendServiceFieldsChanged compares backend service-level fields (excluding children).
func hasBackendServiceFieldsChanged(old *ent.BronzeGCPComputeBackendService, new *BackendServiceData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Fingerprint != new.Fingerprint ||
		old.LoadBalancingScheme != new.LoadBalancingScheme ||
		old.Protocol != new.Protocol ||
		old.PortName != new.PortName ||
		old.Port != new.Port ||
		old.TimeoutSec != new.TimeoutSec ||
		old.Region != new.Region ||
		old.Network != new.Network ||
		old.SecurityPolicy != new.SecurityPolicy ||
		old.EdgeSecurityPolicy != new.EdgeSecurityPolicy ||
		old.SessionAffinity != new.SessionAffinity ||
		old.AffinityCookieTTLSec != new.AffinityCookieTtlSec ||
		old.LocalityLbPolicy != new.LocalityLbPolicy ||
		old.CompressionMode != new.CompressionMode ||
		old.ServiceLbPolicy != new.ServiceLbPolicy ||
		old.EnableCdn != new.EnableCdn ||
		jsonArrayChanged(old.HealthChecksJSON, new.HealthChecksJSON) ||
		jsonArrayChanged(old.LocalityLbPoliciesJSON, new.LocalityLbPoliciesJSON) ||
		jsonArrayChanged(old.UsedByJSON, new.UsedByJSON) ||
		jsonArrayChanged(old.CustomRequestHeadersJSON, new.CustomRequestHeadersJSON) ||
		jsonArrayChanged(old.CustomResponseHeadersJSON, new.CustomResponseHeadersJSON) ||
		jsonObjChanged(old.CdnPolicyJSON, new.CdnPolicyJSON) ||
		jsonObjChanged(old.CircuitBreakersJSON, new.CircuitBreakersJSON) ||
		jsonObjChanged(old.ConnectionDrainingJSON, new.ConnectionDrainingJSON) ||
		jsonObjChanged(old.ConnectionTrackingPolicyJSON, new.ConnectionTrackingPolicyJSON) ||
		jsonObjChanged(old.ConsistentHashJSON, new.ConsistentHashJSON) ||
		jsonObjChanged(old.FailoverPolicyJSON, new.FailoverPolicyJSON) ||
		jsonObjChanged(old.IapJSON, new.IapJSON) ||
		jsonObjChanged(old.LogConfigJSON, new.LogConfigJSON) ||
		jsonObjChanged(old.MaxStreamDurationJSON, new.MaxStreamDurationJSON) ||
		jsonObjChanged(old.OutlierDetectionJSON, new.OutlierDetectionJSON) ||
		jsonObjChanged(old.SecuritySettingsJSON, new.SecuritySettingsJSON) ||
		jsonObjChanged(old.SubsettingJSON, new.SubsettingJSON) ||
		jsonArrayChanged(old.ServiceBindingsJSON, new.ServiceBindingsJSON)
}

func jsonArrayChanged(old, new []interface{}) bool {
	oldBytes, _ := json.Marshal(old)
	newBytes, _ := json.Marshal(new)
	return string(oldBytes) != string(newBytes)
}

func jsonObjChanged(old, new map[string]interface{}) bool {
	oldBytes, _ := json.Marshal(old)
	newBytes, _ := json.Marshal(new)
	return string(oldBytes) != string(newBytes)
}

func diffBackendsData(old []*ent.BronzeGCPComputeBackendServiceBackend, new []BackendData) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old backends by group
	oldMap := make(map[string]*ent.BronzeGCPComputeBackendServiceBackend, len(old))
	for _, b := range old {
		oldMap[b.Group] = b
	}

	// Compare with new backends
	for _, b := range new {
		oldBackend, ok := oldMap[b.Group]
		if !ok {
			diff.HasChanges = true
			return diff
		}
		if oldBackend.BalancingMode != b.BalancingMode ||
			oldBackend.CapacityScaler != b.CapacityScaler ||
			oldBackend.Description != b.Description ||
			oldBackend.Failover != b.Failover ||
			oldBackend.MaxConnections != b.MaxConnections ||
			oldBackend.MaxConnectionsPerEndpoint != b.MaxConnectionsPerEndpoint ||
			oldBackend.MaxConnectionsPerInstance != b.MaxConnectionsPerInstance ||
			oldBackend.MaxRate != b.MaxRate ||
			oldBackend.MaxRatePerEndpoint != b.MaxRatePerEndpoint ||
			oldBackend.MaxRatePerInstance != b.MaxRatePerInstance ||
			oldBackend.MaxUtilization != b.MaxUtilization ||
			oldBackend.Preference != b.Preference {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}
