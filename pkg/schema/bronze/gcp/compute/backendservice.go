package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeBackendService represents a GCP Compute backend service in the bronze layer.
type BronzeGCPComputeBackendService struct {
	ent.Schema
}

func (BronzeGCPComputeBackendService) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeBackendService) Fields() []ent.Field {
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

		// JSON fields (arrays)
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

		// JSON fields (objects)
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

func (BronzeGCPComputeBackendService) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("backends", BronzeGCPComputeBackendServiceBackend.Type),
	}
}

func (BronzeGCPComputeBackendService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeBackendService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_backend_services"},
	}
}

// BronzeGCPComputeBackendServiceBackend represents a backend within a backend service.
type BronzeGCPComputeBackendServiceBackend struct {
	ent.Schema
}

func (BronzeGCPComputeBackendServiceBackend) Fields() []ent.Field {
	return []ent.Field{
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

func (BronzeGCPComputeBackendServiceBackend) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("backend_service", BronzeGCPComputeBackendService.Type).
			Ref("backends").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeBackendServiceBackend) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_backend_service_backends"},
	}
}
