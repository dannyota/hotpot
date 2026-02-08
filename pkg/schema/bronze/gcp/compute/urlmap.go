package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeUrlMap represents a GCP Compute Engine URL map in the bronze layer.
type BronzeGCPComputeUrlMap struct {
	ent.Schema
}

func (BronzeGCPComputeUrlMap) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.Timestamp{}}
}

func (BronzeGCPComputeUrlMap) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable().Comment("GCP API ID"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("fingerprint").Optional(),
		field.String("default_service").Optional().Comment("URL of the BackendService for unmatched traffic"),
		field.String("region").Optional(),

		// JSON fields for complex nested structures
		field.JSON("host_rules_json", []interface{}{}).Optional().Comment("Host rules for URL routing"),
		field.JSON("path_matchers_json", []interface{}{}).Optional().Comment("Path matchers for URL routing"),
		field.JSON("tests_json", []interface{}{}).Optional().Comment("URL map tests"),
		field.JSON("default_route_action_json", map[string]interface{}{}).Optional().Comment("Default route action"),
		field.JSON("default_url_redirect_json", map[string]interface{}{}).Optional().Comment("Default URL redirect"),
		field.JSON("header_action_json", map[string]interface{}{}).Optional().Comment("Header action"),

		field.String("project_id").NotEmpty(),
	}
}

func (BronzeGCPComputeUrlMap) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeUrlMap) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_url_maps"},
	}
}
