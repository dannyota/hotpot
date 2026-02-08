package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

type BronzeGCPComputeNeg struct {
	ent.Schema
}

func (BronzeGCPComputeNeg) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeNeg) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("network").Optional(),
		field.String("subnetwork").Optional(),
		field.String("zone").Optional(),
		field.String("network_endpoint_type").Optional(),
		field.String("default_port").Optional(),
		field.String("size").Optional(),
		field.String("region").Optional(),
		// JSON fields
		field.JSON("annotations_json", map[string]interface{}{}).Optional(),
		field.JSON("app_engine_json", map[string]interface{}{}).Optional(),
		field.JSON("cloud_function_json", map[string]interface{}{}).Optional(),
		field.JSON("cloud_run_json", map[string]interface{}{}).Optional(),
		field.JSON("psc_data_json", map[string]interface{}{}).Optional(),
		// Metadata
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeGCPComputeNeg) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("zone"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeNeg) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_negs"},
	}
}
