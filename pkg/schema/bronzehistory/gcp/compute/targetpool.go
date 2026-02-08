package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeTargetPool stores historical snapshots of GCP Compute target pools.
type BronzeHistoryGCPComputeTargetPool struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeTargetPool) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeTargetPool) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").Unique().Immutable(),
		field.String("resource_id").NotEmpty().Comment("Link to bronze target pool by resource_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("session_affinity").Optional(),
		field.String("backup_pool").Optional(),
		field.Float32("failover_ratio").Optional(),
		field.String("security_policy").Optional(),
		field.String("region").Optional(),
		field.JSON("health_checks_json", []interface{}{}).Optional(),
		field.JSON("instances_json", []interface{}{}).Optional(),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeTargetPool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeTargetPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_pools_history"},
	}
}
