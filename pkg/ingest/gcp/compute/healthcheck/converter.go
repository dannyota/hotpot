package healthcheck

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertHealthCheck converts a GCP API HealthCheck to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertHealthCheck(h *computepb.HealthCheck, projectID string, collectedAt time.Time) (bronze.GCPComputeHealthCheck, error) {
	check := bronze.GCPComputeHealthCheck{
		ResourceID:         fmt.Sprintf("%d", h.GetId()),
		Name:               h.GetName(),
		Description:        h.GetDescription(),
		CreationTimestamp:  h.GetCreationTimestamp(),
		SelfLink:           h.GetSelfLink(),
		Type:               h.GetType(),
		Region:             h.GetRegion(),
		CheckIntervalSec:  h.GetCheckIntervalSec(),
		TimeoutSec:         h.GetTimeoutSec(),
		HealthyThreshold:   h.GetHealthyThreshold(),
		UnhealthyThreshold: h.GetUnhealthyThreshold(),
		ProjectID:          projectID,
		CollectedAt:        collectedAt,
	}

	// Convert JSONB fields (nil -> SQL NULL, data -> JSON bytes)
	var err error
	if h.TcpHealthCheck != nil {
		check.TcpHealthCheckJSON, err = json.Marshal(h.TcpHealthCheck)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal TCP health check JSON for %s: %w", h.GetName(), err)
		}
	}
	if h.HttpHealthCheck != nil {
		check.HttpHealthCheckJSON, err = json.Marshal(h.HttpHealthCheck)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal HTTP health check JSON for %s: %w", h.GetName(), err)
		}
	}
	if h.HttpsHealthCheck != nil {
		check.HttpsHealthCheckJSON, err = json.Marshal(h.HttpsHealthCheck)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal HTTPS health check JSON for %s: %w", h.GetName(), err)
		}
	}
	if h.Http2HealthCheck != nil {
		check.Http2HealthCheckJSON, err = json.Marshal(h.Http2HealthCheck)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal HTTP2 health check JSON for %s: %w", h.GetName(), err)
		}
	}
	if h.SslHealthCheck != nil {
		check.SslHealthCheckJSON, err = json.Marshal(h.SslHealthCheck)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal SSL health check JSON for %s: %w", h.GetName(), err)
		}
	}
	if h.GrpcHealthCheck != nil {
		check.GrpcHealthCheckJSON, err = json.Marshal(h.GrpcHealthCheck)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal gRPC health check JSON for %s: %w", h.GetName(), err)
		}
	}
	if h.LogConfig != nil {
		check.LogConfigJSON, err = json.Marshal(h.LogConfig)
		if err != nil {
			return bronze.GCPComputeHealthCheck{}, fmt.Errorf("failed to marshal log config JSON for %s: %w", h.GetName(), err)
		}
	}

	return check, nil
}
