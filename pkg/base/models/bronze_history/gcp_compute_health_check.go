package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeHealthCheck stores historical snapshots of GCP Compute Engine health checks.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
// No child history tables — protocol-specific checks are JSONB on the parent record.
type GCPComputeHealthCheck struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (HealthCheck has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All health check fields (same as bronze.GCPComputeHealthCheck)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	Type              string `gorm:"column:type;type:varchar(50)" json:"type"`
	Region            string `gorm:"column:region;type:text" json:"region"`

	// Check parameters
	CheckIntervalSec   int32 `gorm:"column:check_interval_sec" json:"checkIntervalSec"`
	TimeoutSec         int32 `gorm:"column:timeout_sec" json:"timeoutSec"`
	HealthyThreshold   int32 `gorm:"column:healthy_threshold" json:"healthyThreshold"`
	UnhealthyThreshold int32 `gorm:"column:unhealthy_threshold" json:"unhealthyThreshold"`

	// Protocol-specific checks (JSONB)
	TcpHealthCheckJSON   jsonb.JSON `gorm:"column:tcp_health_check_json;type:jsonb" json:"tcpHealthCheck"`
	HttpHealthCheckJSON  jsonb.JSON `gorm:"column:http_health_check_json;type:jsonb" json:"httpHealthCheck"`
	HttpsHealthCheckJSON jsonb.JSON `gorm:"column:https_health_check_json;type:jsonb" json:"httpsHealthCheck"`
	Http2HealthCheckJSON jsonb.JSON `gorm:"column:http2_health_check_json;type:jsonb" json:"http2HealthCheck"`
	SslHealthCheckJSON   jsonb.JSON `gorm:"column:ssl_health_check_json;type:jsonb" json:"sslHealthCheck"`
	GrpcHealthCheckJSON  jsonb.JSON `gorm:"column:grpc_health_check_json;type:jsonb" json:"grpcHealthCheck"`
	LogConfigJSON        jsonb.JSON `gorm:"column:log_config_json;type:jsonb" json:"logConfig"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeHealthCheck) TableName() string {
	return "bronze_history.gcp_compute_health_checks"
}
