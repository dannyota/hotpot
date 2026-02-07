package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeHealthCheck represents a GCP Compute Engine health check in the bronze layer.
// Fields preserve raw API response data from compute.healthChecks.aggregatedList.
// Protocol-specific checks (TCP, HTTP, etc.) are stored as JSONB — no child tables.
type GCPComputeHealthCheck struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	Type              string `gorm:"column:type;type:varchar(50);index" json:"type"`
	Region            string `gorm:"column:region;type:text" json:"region"`

	// Check parameters
	CheckIntervalSec   int32 `gorm:"column:check_interval_sec" json:"checkIntervalSec"`
	TimeoutSec         int32 `gorm:"column:timeout_sec" json:"timeoutSec"`
	HealthyThreshold   int32 `gorm:"column:healthy_threshold" json:"healthyThreshold"`
	UnhealthyThreshold int32 `gorm:"column:unhealthy_threshold" json:"unhealthyThreshold"`

	// Protocol-specific checks (JSONB — no child tables)

	// TcpHealthCheckJSON contains TCP health check configuration.
	//
	//	{"port":80,"proxyHeader":"NONE"}
	TcpHealthCheckJSON jsonb.JSON `gorm:"column:tcp_health_check_json;type:jsonb" json:"tcpHealthCheck"`

	// HttpHealthCheckJSON contains HTTP health check configuration.
	//
	//	{"port":80,"requestPath":"/","proxyHeader":"NONE"}
	HttpHealthCheckJSON jsonb.JSON `gorm:"column:http_health_check_json;type:jsonb" json:"httpHealthCheck"`

	// HttpsHealthCheckJSON contains HTTPS health check configuration.
	HttpsHealthCheckJSON jsonb.JSON `gorm:"column:https_health_check_json;type:jsonb" json:"httpsHealthCheck"`

	// Http2HealthCheckJSON contains HTTP/2 health check configuration.
	Http2HealthCheckJSON jsonb.JSON `gorm:"column:http2_health_check_json;type:jsonb" json:"http2HealthCheck"`

	// SslHealthCheckJSON contains SSL health check configuration.
	SslHealthCheckJSON jsonb.JSON `gorm:"column:ssl_health_check_json;type:jsonb" json:"sslHealthCheck"`

	// GrpcHealthCheckJSON contains gRPC health check configuration.
	//
	//	{"port":443,"grpcServiceName":"..."}
	GrpcHealthCheckJSON jsonb.JSON `gorm:"column:grpc_health_check_json;type:jsonb" json:"grpcHealthCheck"`

	// LogConfigJSON contains logging configuration.
	//
	//	{"enable":true}
	LogConfigJSON jsonb.JSON `gorm:"column:log_config_json;type:jsonb" json:"logConfig"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`
}

func (GCPComputeHealthCheck) TableName() string {
	return "bronze.gcp_compute_health_checks"
}
