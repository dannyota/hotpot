package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeUrlMap stores historical snapshots of GCP Compute URL maps.
type BronzeHistoryGCPComputeUrlMap struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeUrlMap) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeUrlMap) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").Unique().Immutable(),
		field.String("resource_id").NotEmpty().Comment("Link to bronze URL map by resource_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("fingerprint").Optional(),
		field.String("default_service").Optional(),
		field.String("region").Optional(),
		field.JSON("host_rules_json", []interface{}{}).Optional(),
		field.JSON("path_matchers_json", []interface{}{}).Optional(),
		field.JSON("tests_json", []interface{}{}).Optional(),
		field.JSON("default_route_action_json", map[string]interface{}{}).Optional(),
		field.JSON("default_url_redirect_json", map[string]interface{}{}).Optional(),
		field.JSON("header_action_json", map[string]interface{}{}).Optional(),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeUrlMap) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeUrlMap) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_url_maps_history"},
	}
}
