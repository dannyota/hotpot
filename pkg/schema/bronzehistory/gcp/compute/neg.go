package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

type BronzeHistoryGCPComputeNeg struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeNeg) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeNeg) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").Unique().Immutable(),
		field.String("resource_id").NotEmpty().Comment("Link to bronze NEG"),
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
		field.JSON("annotations_json", map[string]interface{}{}).Optional(),
		field.JSON("app_engine_json", map[string]interface{}{}).Optional(),
		field.JSON("cloud_function_json", map[string]interface{}{}).Optional(),
		field.JSON("cloud_run_json", map[string]interface{}{}).Optional(),
		field.JSON("psc_data_json", map[string]interface{}{}).Optional(),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeNeg) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeNeg) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_negs_history"},
	}
}
