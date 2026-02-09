package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeTargetPool represents a GCP Compute Engine target pool in the bronze layer.
type BronzeGCPComputeTargetPool struct {
	ent.Schema
}

func (BronzeGCPComputeTargetPool) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.Timestamp{}}
}

func (BronzeGCPComputeTargetPool) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable().Comment("GCP API ID"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("session_affinity").Optional().Comment("Session affinity option"),
		field.String("backup_pool").Optional().Comment("URL of the backup target pool"),
		field.Float32("failover_ratio").Optional().Comment("Ratio of healthy VMs to trigger failover"),
		field.String("security_policy").Optional(),
		field.String("region").Optional(),
		field.JSON("health_checks_json", []interface{}{}).Optional().Comment("URLs to HttpHealthCheck resources"),
		field.JSON("instances_json", []interface{}{}).Optional().Comment("URLs to instance resources"),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeGCPComputeTargetPool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeTargetPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_pools"},
	}
}
