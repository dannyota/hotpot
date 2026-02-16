package redis

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPRedisInstance stores historical snapshots of GCP Memorystore Redis instances.
type BronzeHistoryGCPRedisInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPRedisInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPRedisInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Redis instance by resource_id"),

		// All instance fields
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.String("location_id").
			Optional(),
		field.String("alternative_location_id").
			Optional(),
		field.String("redis_version").
			Optional(),
		field.String("reserved_ip_range").
			Optional(),
		field.String("secondary_ip_range").
			Optional(),
		field.String("host").
			Optional(),
		field.Int32("port").
			Optional(),
		field.String("current_location_id").
			Optional(),
		field.String("create_time").
			Optional(),
		field.Int32("state").
			Optional(),
		field.String("status_message").
			Optional(),

		field.JSON("redis_configs_json", json.RawMessage{}).
			Optional(),

		field.Int32("tier").
			Optional(),
		field.Int32("memory_size_gb").
			Optional(),
		field.String("authorized_network").
			Optional(),
		field.String("persistence_iam_identity").
			Optional(),
		field.Int32("connect_mode").
			Optional(),
		field.Bool("auth_enabled").
			Default(false),

		field.JSON("server_ca_certs_json", json.RawMessage{}).
			Optional(),

		field.Int32("transit_encryption_mode").
			Optional(),

		field.JSON("maintenance_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_schedule_json", json.RawMessage{}).
			Optional(),

		field.Int32("replica_count").
			Optional(),

		field.JSON("nodes_json", json.RawMessage{}).
			Optional(),

		field.String("read_endpoint").
			Optional(),
		field.Int32("read_endpoint_port").
			Optional(),
		field.Int32("read_replicas_mode").
			Optional(),
		field.String("customer_managed_key").
			Optional(),

		field.JSON("persistence_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("suspension_reasons_json", json.RawMessage{}).
			Optional(),

		field.String("maintenance_version").
			Optional(),

		field.JSON("available_maintenance_versions_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPRedisInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPRedisInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_redis_instances_history"},
	}
}
