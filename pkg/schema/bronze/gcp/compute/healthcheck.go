package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeHealthCheck represents a GCP Compute Engine health check in the bronze layer.
// Fields preserve raw API response data from compute.healthChecks.aggregatedList.
// Protocol-specific checks (TCP, HTTP, etc.) are stored as JSON - no child tables.
type BronzeGCPComputeHealthCheck struct {
	ent.Schema
}

func (BronzeGCPComputeHealthCheck) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeHealthCheck) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("type").
			Optional().
			Comment("Health check type (TCP, HTTP, HTTPS, etc.)"),
		field.String("region").
			Optional(),

		// Check parameters
		field.Int32("check_interval_sec").
			Optional(),
		field.Int32("timeout_sec").
			Optional(),
		field.Int32("healthy_threshold").
			Optional(),
		field.Int32("unhealthy_threshold").
			Optional(),

		// Protocol-specific checks (JSON - no child tables)
		field.JSON("tcp_health_check_json", map[string]interface{}{}).
			Optional().
			Comment("TCP health check configuration"),
		field.JSON("http_health_check_json", map[string]interface{}{}).
			Optional().
			Comment("HTTP health check configuration"),
		field.JSON("https_health_check_json", map[string]interface{}{}).
			Optional().
			Comment("HTTPS health check configuration"),
		field.JSON("http2_health_check_json", map[string]interface{}{}).
			Optional().
			Comment("HTTP/2 health check configuration"),
		field.JSON("ssl_health_check_json", map[string]interface{}{}).
			Optional().
			Comment("SSL health check configuration"),
		field.JSON("grpc_health_check_json", map[string]interface{}{}).
			Optional().
			Comment("gRPC health check configuration"),
		field.JSON("log_config_json", map[string]interface{}{}).
			Optional().
			Comment("Logging configuration"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeHealthCheck) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("type"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeHealthCheck) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_health_checks"},
	}
}
