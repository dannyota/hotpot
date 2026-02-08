package backendservice

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// BackendServiceData holds converted backend service data ready for Ent insertion.
type BackendServiceData struct {
	ID                   string
	Name                 string
	Description          string
	CreationTimestamp    string
	SelfLink             string
	Fingerprint          string
	LoadBalancingScheme  string
	Protocol             string
	PortName             string
	Port                 string
	TimeoutSec           string
	Region               string
	Network              string
	SecurityPolicy       string
	EdgeSecurityPolicy   string
	SessionAffinity      string
	AffinityCookieTtlSec string
	LocalityLbPolicy     string
	CompressionMode      string
	ServiceLbPolicy      string
	EnableCdn            bool
	// JSON arrays
	HealthChecksJSON          []interface{}
	LocalityLbPoliciesJSON    []interface{}
	UsedByJSON                []interface{}
	CustomRequestHeadersJSON  []interface{}
	CustomResponseHeadersJSON []interface{}
	// JSON objects
	CdnPolicyJSON                map[string]interface{}
	CircuitBreakersJSON          map[string]interface{}
	ConnectionDrainingJSON       map[string]interface{}
	ConnectionTrackingPolicyJSON map[string]interface{}
	ConsistentHashJSON           map[string]interface{}
	FailoverPolicyJSON           map[string]interface{}
	IapJSON                      map[string]interface{}
	LogConfigJSON                map[string]interface{}
	MaxStreamDurationJSON        map[string]interface{}
	OutlierDetectionJSON         map[string]interface{}
	SecuritySettingsJSON         map[string]interface{}
	SubsettingJSON               map[string]interface{}
	ServiceBindingsJSON          []interface{}
	// Child entities
	Backends    []BackendData
	ProjectID   string
	CollectedAt time.Time
}

// BackendData holds converted backend data.
type BackendData struct {
	Group                     string
	BalancingMode             string
	CapacityScaler            string
	Description               string
	Failover                  bool
	MaxConnections            string
	MaxConnectionsPerEndpoint string
	MaxConnectionsPerInstance string
	MaxRate                   string
	MaxRatePerEndpoint        string
	MaxRatePerInstance        string
	MaxUtilization            string
	Preference                string
}

// ConvertBackendService converts a GCP API BackendService to BackendServiceData.
// Preserves raw API data with minimal transformation.
func ConvertBackendService(bs *computepb.BackendService, projectID string, collectedAt time.Time) (*BackendServiceData, error) {
	data := &BackendServiceData{
		ID:                   fmt.Sprintf("%d", bs.GetId()),
		Name:                 bs.GetName(),
		Description:          bs.GetDescription(),
		CreationTimestamp:    bs.GetCreationTimestamp(),
		SelfLink:             bs.GetSelfLink(),
		Fingerprint:          bs.GetFingerprint(),
		LoadBalancingScheme:  bs.GetLoadBalancingScheme(),
		Protocol:             bs.GetProtocol(),
		PortName:             bs.GetPortName(),
		Port:                 fmt.Sprintf("%d", bs.GetPort()),
		TimeoutSec:           fmt.Sprintf("%d", bs.GetTimeoutSec()),
		Region:               bs.GetRegion(),
		Network:              bs.GetNetwork(),
		SecurityPolicy:       bs.GetSecurityPolicy(),
		EdgeSecurityPolicy:   bs.GetEdgeSecurityPolicy(),
		SessionAffinity:      bs.GetSessionAffinity(),
		AffinityCookieTtlSec: fmt.Sprintf("%d", bs.GetAffinityCookieTtlSec()),
		LocalityLbPolicy:     bs.GetLocalityLbPolicy(),
		CompressionMode:      bs.GetCompressionMode(),
		ServiceLbPolicy:      bs.GetServiceLbPolicy(),
		EnableCdn:            bs.GetEnableCDN(),
		ProjectID:            projectID,
		CollectedAt:          collectedAt,
	}

	// Convert JSON array fields
	if bs.GetHealthChecks() != nil {
		b, err := json.Marshal(bs.GetHealthChecks())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal health checks for backend service %s: %w", bs.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.HealthChecksJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal health checks: %w", err)
		}
	}

	if bs.GetLocalityLbPolicies() != nil {
		b, err := json.Marshal(bs.GetLocalityLbPolicies())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal locality lb policies for backend service %s: %w", bs.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.LocalityLbPoliciesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal locality lb policies: %w", err)
		}
	}

	if bs.GetUsedBy() != nil {
		b, err := json.Marshal(bs.GetUsedBy())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal used by for backend service %s: %w", bs.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.UsedByJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal used by: %w", err)
		}
	}

	if bs.GetCustomRequestHeaders() != nil {
		b, err := json.Marshal(bs.GetCustomRequestHeaders())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom request headers for backend service %s: %w", bs.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.CustomRequestHeadersJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom request headers: %w", err)
		}
	}

	if bs.GetCustomResponseHeaders() != nil {
		b, err := json.Marshal(bs.GetCustomResponseHeaders())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom response headers for backend service %s: %w", bs.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.CustomResponseHeadersJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom response headers: %w", err)
		}
	}

	if bs.GetServiceBindings() != nil {
		b, err := json.Marshal(bs.GetServiceBindings())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal service bindings for backend service %s: %w", bs.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.ServiceBindingsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service bindings: %w", err)
		}
	}

	// Convert JSON object fields
	jsonObjFields := []struct {
		name   string
		getter func() interface{}
		target *map[string]interface{}
	}{
		{"cdn_policy", func() interface{} { return bs.GetCdnPolicy() }, &data.CdnPolicyJSON},
		{"circuit_breakers", func() interface{} { return bs.GetCircuitBreakers() }, &data.CircuitBreakersJSON},
		{"connection_draining", func() interface{} { return bs.GetConnectionDraining() }, &data.ConnectionDrainingJSON},
		{"connection_tracking_policy", func() interface{} { return bs.GetConnectionTrackingPolicy() }, &data.ConnectionTrackingPolicyJSON},
		{"consistent_hash", func() interface{} { return bs.GetConsistentHash() }, &data.ConsistentHashJSON},
		{"failover_policy", func() interface{} { return bs.GetFailoverPolicy() }, &data.FailoverPolicyJSON},
		{"iap", func() interface{} { return bs.GetIap() }, &data.IapJSON},
		{"log_config", func() interface{} { return bs.GetLogConfig() }, &data.LogConfigJSON},
		{"max_stream_duration", func() interface{} { return bs.GetMaxStreamDuration() }, &data.MaxStreamDurationJSON},
		{"outlier_detection", func() interface{} { return bs.GetOutlierDetection() }, &data.OutlierDetectionJSON},
		{"security_settings", func() interface{} { return bs.GetSecuritySettings() }, &data.SecuritySettingsJSON},
		{"subsetting", func() interface{} { return bs.GetSubsetting() }, &data.SubsettingJSON},
	}

	for _, f := range jsonObjFields {
		val := f.getter()
		if val == nil {
			continue
		}
		b, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s for backend service %s: %w", f.name, bs.GetName(), err)
		}
		if err := json.Unmarshal(b, f.target); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %w", f.name, err)
		}
	}

	// Convert child backends
	data.Backends = ConvertBackends(bs.GetBackends())

	return data, nil
}

// ConvertBackends converts GCP API Backend entries to BackendData.
func ConvertBackends(backends []*computepb.Backend) []BackendData {
	if len(backends) == 0 {
		return nil
	}

	result := make([]BackendData, 0, len(backends))
	for _, b := range backends {
		result = append(result, BackendData{
			Group:                     b.GetGroup(),
			BalancingMode:             b.GetBalancingMode(),
			CapacityScaler:            fmt.Sprintf("%g", b.GetCapacityScaler()),
			Description:               b.GetDescription(),
			Failover:                  b.GetFailover(),
			MaxConnections:            fmt.Sprintf("%d", b.GetMaxConnections()),
			MaxConnectionsPerEndpoint: fmt.Sprintf("%d", b.GetMaxConnectionsPerEndpoint()),
			MaxConnectionsPerInstance: fmt.Sprintf("%d", b.GetMaxConnectionsPerInstance()),
			MaxRate:                   fmt.Sprintf("%d", b.GetMaxRate()),
			MaxRatePerEndpoint:        fmt.Sprintf("%g", b.GetMaxRatePerEndpoint()),
			MaxRatePerInstance:        fmt.Sprintf("%g", b.GetMaxRatePerInstance()),
			MaxUtilization:            fmt.Sprintf("%g", b.GetMaxUtilization()),
			Preference:                b.GetPreference(),
		})
	}

	return result
}
