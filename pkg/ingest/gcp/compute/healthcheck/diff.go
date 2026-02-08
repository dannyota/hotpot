package healthcheck

import (
	"reflect"

	"hotpot/pkg/storage/ent"
)

// HealthCheckDiff represents changes between old and new health check states.
type HealthCheckDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffHealthCheckData compares existing Ent entity with new HealthCheckData and returns differences.
func DiffHealthCheckData(old *ent.BronzeGCPComputeHealthCheck, new *HealthCheckData) *HealthCheckDiff {
	diff := &HealthCheckDiff{}

	// New health check
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Type != new.Type ||
		old.Region != new.Region ||
		old.CheckIntervalSec != new.CheckIntervalSec ||
		old.TimeoutSec != new.TimeoutSec ||
		old.HealthyThreshold != new.HealthyThreshold ||
		old.UnhealthyThreshold != new.UnhealthyThreshold ||
		!reflect.DeepEqual(old.TCPHealthCheckJSON, new.TcpHealthCheckJSON) ||
		!reflect.DeepEqual(old.HTTPHealthCheckJSON, new.HttpHealthCheckJSON) ||
		!reflect.DeepEqual(old.HTTPSHealthCheckJSON, new.HttpsHealthCheckJSON) ||
		!reflect.DeepEqual(old.Http2HealthCheckJSON, new.Http2HealthCheckJSON) ||
		!reflect.DeepEqual(old.SslHealthCheckJSON, new.SslHealthCheckJSON) ||
		!reflect.DeepEqual(old.GrpcHealthCheckJSON, new.GrpcHealthCheckJSON) ||
		!reflect.DeepEqual(old.LogConfigJSON, new.LogConfigJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the health check changed.
func (d *HealthCheckDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
