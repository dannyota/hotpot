package healthcheck

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// HealthCheckData holds converted health check data ready for Ent insertion.
type HealthCheckData struct {
	ID                   string
	Name                 string
	Description          string
	CreationTimestamp    string
	SelfLink             string
	Type                 string
	Region               string
	CheckIntervalSec     int32
	TimeoutSec           int32
	HealthyThreshold     int32
	UnhealthyThreshold   int32
	TcpHealthCheckJSON   map[string]interface{}
	HttpHealthCheckJSON  map[string]interface{}
	HttpsHealthCheckJSON map[string]interface{}
	Http2HealthCheckJSON map[string]interface{}
	SslHealthCheckJSON   map[string]interface{}
	GrpcHealthCheckJSON  map[string]interface{}
	LogConfigJSON        map[string]interface{}
	ProjectID            string
	CollectedAt          time.Time
}

// ConvertHealthCheck converts a GCP API HealthCheck to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertHealthCheck(h *computepb.HealthCheck, projectID string, collectedAt time.Time) (*HealthCheckData, error) {
	data := &HealthCheckData{
		ID:                 fmt.Sprintf("%d", h.GetId()),
		Name:               h.GetName(),
		Description:        h.GetDescription(),
		CreationTimestamp:  h.GetCreationTimestamp(),
		SelfLink:           h.GetSelfLink(),
		Type:               h.GetType(),
		Region:             h.GetRegion(),
		CheckIntervalSec:   h.GetCheckIntervalSec(),
		TimeoutSec:         h.GetTimeoutSec(),
		HealthyThreshold:   h.GetHealthyThreshold(),
		UnhealthyThreshold: h.GetUnhealthyThreshold(),
		ProjectID:          projectID,
		CollectedAt:        collectedAt,
	}

	// Convert JSONB fields
	if h.TcpHealthCheck != nil {
		var tcpMap map[string]interface{}
		tcpBytes, err := json.Marshal(h.TcpHealthCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal TCP health check JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(tcpBytes, &tcpMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal TCP health check JSON for %s: %w", h.GetName(), err)
		}
		data.TcpHealthCheckJSON = tcpMap
	}
	if h.HttpHealthCheck != nil {
		var httpMap map[string]interface{}
		httpBytes, err := json.Marshal(h.HttpHealthCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal HTTP health check JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(httpBytes, &httpMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal HTTP health check JSON for %s: %w", h.GetName(), err)
		}
		data.HttpHealthCheckJSON = httpMap
	}
	if h.HttpsHealthCheck != nil {
		var httpsMap map[string]interface{}
		httpsBytes, err := json.Marshal(h.HttpsHealthCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal HTTPS health check JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(httpsBytes, &httpsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal HTTPS health check JSON for %s: %w", h.GetName(), err)
		}
		data.HttpsHealthCheckJSON = httpsMap
	}
	if h.Http2HealthCheck != nil {
		var http2Map map[string]interface{}
		http2Bytes, err := json.Marshal(h.Http2HealthCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal HTTP2 health check JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(http2Bytes, &http2Map); err != nil {
			return nil, fmt.Errorf("failed to unmarshal HTTP2 health check JSON for %s: %w", h.GetName(), err)
		}
		data.Http2HealthCheckJSON = http2Map
	}
	if h.SslHealthCheck != nil {
		var sslMap map[string]interface{}
		sslBytes, err := json.Marshal(h.SslHealthCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal SSL health check JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(sslBytes, &sslMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SSL health check JSON for %s: %w", h.GetName(), err)
		}
		data.SslHealthCheckJSON = sslMap
	}
	if h.GrpcHealthCheck != nil {
		var grpcMap map[string]interface{}
		grpcBytes, err := json.Marshal(h.GrpcHealthCheck)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal gRPC health check JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(grpcBytes, &grpcMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal gRPC health check JSON for %s: %w", h.GetName(), err)
		}
		data.GrpcHealthCheckJSON = grpcMap
	}
	if h.LogConfig != nil {
		var logMap map[string]interface{}
		logBytes, err := json.Marshal(h.LogConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal log config JSON for %s: %w", h.GetName(), err)
		}
		if err := json.Unmarshal(logBytes, &logMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal log config JSON for %s: %w", h.GetName(), err)
		}
		data.LogConfigJSON = logMap
	}

	return data, nil
}
