package healthcheck

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// HealthCheckDiff represents changes between old and new health check states.
type HealthCheckDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffHealthCheck compares old and new health check states.
func DiffHealthCheck(old, new *bronze.GCPComputeHealthCheck) *HealthCheckDiff {
	if old == nil {
		return &HealthCheckDiff{
			IsNew: true,
		}
	}

	diff := &HealthCheckDiff{}
	diff.IsChanged = hasHealthCheckFieldsChanged(old, new)

	return diff
}

// HasAnyChange returns true if any part of the health check changed.
func (d *HealthCheckDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasHealthCheckFieldsChanged(old, new *bronze.GCPComputeHealthCheck) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Type != new.Type ||
		old.Region != new.Region ||
		old.CheckIntervalSec != new.CheckIntervalSec ||
		old.TimeoutSec != new.TimeoutSec ||
		old.HealthyThreshold != new.HealthyThreshold ||
		old.UnhealthyThreshold != new.UnhealthyThreshold ||
		jsonb.Changed(old.TcpHealthCheckJSON, new.TcpHealthCheckJSON) ||
		jsonb.Changed(old.HttpHealthCheckJSON, new.HttpHealthCheckJSON) ||
		jsonb.Changed(old.HttpsHealthCheckJSON, new.HttpsHealthCheckJSON) ||
		jsonb.Changed(old.Http2HealthCheckJSON, new.Http2HealthCheckJSON) ||
		jsonb.Changed(old.SslHealthCheckJSON, new.SslHealthCheckJSON) ||
		jsonb.Changed(old.GrpcHealthCheckJSON, new.GrpcHealthCheckJSON) ||
		jsonb.Changed(old.LogConfigJSON, new.LogConfigJSON)
}
