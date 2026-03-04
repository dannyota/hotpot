package k8snode

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SilverK8sNodeNormalized holds per-provider normalized rows before merge.
type SilverK8sNodeNormalized struct {
	ent.Schema
}

func (SilverK8sNodeNormalized) Fields() []ent.Field {
	return []ent.Field{
		// Identity — unique per provider+bronze record.
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("provider").NotEmpty(),
		field.Bool("is_base"),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),

		// Normalized k8s node fields.
		field.String("node_name").Optional(),
		field.String("cluster_name").Optional(),
		field.String("node_pool").Optional(),
		field.String("status"),
		field.String("provisioning").Optional(),
		field.String("cloud_project").Optional(),
		field.String("cloud_zone").Optional(),
		field.String("cloud_machine_type").Optional(),
		field.String("internal_ip").Optional(),
		field.String("external_ip").Optional(),

		// Merge keys — used by merge engine for dedup.
		field.JSON("merge_keys_json", map[string][]string{}),

		// Timestamps from bronze (NOT immutable — upserts may update).
		field.Time("collected_at"),
		field.Time("first_collected_at"),

		// When this row was last normalized.
		field.Time("normalized_at"),
	}
}

func (SilverK8sNodeNormalized) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider"),
		index.Fields("provider", "bronze_resource_id").Unique(),
	}
}

func (SilverK8sNodeNormalized) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "k8s_node_normalized"},
	}
}
