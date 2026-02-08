package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPComputeHealthCheck stores historical snapshots of GCP Compute Engine health checks.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
// No child history tables - protocol-specific checks are JSON on the parent record.
type BronzeHistoryGCPComputeHealthCheck struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeHealthCheck) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze health check by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All health check fields (same as bronze.BronzeGCPComputeHealthCheck)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("type").
			Optional(),
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

		// Protocol-specific checks (JSON)
		field.JSON("tcp_health_check_json", map[string]interface{}{}).
			Optional(),
		field.JSON("http_health_check_json", map[string]interface{}{}).
			Optional(),
		field.JSON("https_health_check_json", map[string]interface{}{}).
			Optional(),
		field.JSON("http2_health_check_json", map[string]interface{}{}).
			Optional(),
		field.JSON("ssl_health_check_json", map[string]interface{}{}).
			Optional(),
		field.JSON("grpc_health_check_json", map[string]interface{}{}).
			Optional(),
		field.JSON("log_config_json", map[string]interface{}{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeHealthCheck) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeHealthCheck) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_health_checks_history"},
	}
}
