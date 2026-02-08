package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeBackendService stores historical snapshots of backend services.
type BronzeHistoryGCPComputeBackendService struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeBackendService) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeBackendService) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze backend service by resource_id"),

		// All backend service fields (same as bronze.BronzeGCPComputeBackendService)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("fingerprint").
			Optional(),
		field.String("load_balancing_scheme").
			Optional(),
		field.String("protocol").
			Optional(),
		field.String("port_name").
			Optional(),
		field.String("port").
			Optional(),
		field.String("timeout_sec").
			Optional(),
		field.String("region").
			Optional(),
		field.String("network").
			Optional(),
		field.String("security_policy").
			Optional(),
		field.String("edge_security_policy").
			Optional(),
		field.String("session_affinity").
			Optional(),
		field.String("affinity_cookie_ttl_sec").
			Optional(),
		field.String("locality_lb_policy").
			Optional(),
		field.String("compression_mode").
			Optional(),
		field.String("service_lb_policy").
			Optional(),
		field.Bool("enable_cdn").
			Default(false),

		// JSON arrays
		field.JSON("health_checks_json", []interface{}{}).
			Optional(),
		field.JSON("locality_lb_policies_json", []interface{}{}).
			Optional(),
		field.JSON("used_by_json", []interface{}{}).
			Optional(),
		field.JSON("custom_request_headers_json", []interface{}{}).
			Optional(),
		field.JSON("custom_response_headers_json", []interface{}{}).
			Optional(),

		// JSON objects
		field.JSON("cdn_policy_json", map[string]interface{}{}).
			Optional(),
		field.JSON("circuit_breakers_json", map[string]interface{}{}).
			Optional(),
		field.JSON("connection_draining_json", map[string]interface{}{}).
			Optional(),
		field.JSON("connection_tracking_policy_json", map[string]interface{}{}).
			Optional(),
		field.JSON("consistent_hash_json", map[string]interface{}{}).
			Optional(),
		field.JSON("failover_policy_json", map[string]interface{}{}).
			Optional(),
		field.JSON("iap_json", map[string]interface{}{}).
			Optional(),
		field.JSON("log_config_json", map[string]interface{}{}).
			Optional(),
		field.JSON("max_stream_duration_json", map[string]interface{}{}).
			Optional(),
		field.JSON("outlier_detection_json", map[string]interface{}{}).
			Optional(),
		field.JSON("security_settings_json", map[string]interface{}{}).
			Optional(),
		field.JSON("subsetting_json", map[string]interface{}{}).
			Optional(),
		field.JSON("service_bindings_json", []interface{}{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeBackendService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeBackendService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_backend_services_history"},
	}
}

// BronzeHistoryGCPComputeBackendServiceBackend stores historical backend entries.
type BronzeHistoryGCPComputeBackendServiceBackend struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeBackendServiceBackend) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("backend_service_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeBackendService"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Backend fields
		field.String("group").
			NotEmpty(),
		field.String("balancing_mode").
			Optional(),
		field.String("capacity_scaler").
			Optional(),
		field.String("description").
			Optional(),
		field.Bool("failover").
			Default(false),
		field.String("max_connections").
			Optional(),
		field.String("max_connections_per_endpoint").
			Optional(),
		field.String("max_connections_per_instance").
			Optional(),
		field.String("max_rate").
			Optional(),
		field.String("max_rate_per_endpoint").
			Optional(),
		field.String("max_rate_per_instance").
			Optional(),
		field.String("max_utilization").
			Optional(),
		field.String("preference").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeBackendServiceBackend) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("backend_service_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeBackendServiceBackend) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_backend_service_backends_history"},
	}
}
