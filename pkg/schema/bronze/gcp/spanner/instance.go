package spanner

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPSpannerInstance represents a GCP Cloud Spanner instance in the bronze layer.
type BronzeGCPSpannerInstance struct {
	ent.Schema
}

func (BronzeGCPSpannerInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSpannerInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Spanner instance resource name"),
		field.String("name").
			NotEmpty(),
		field.String("config").
			Optional().
			Comment("Instance configuration resource name"),
		field.String("display_name").
			Optional(),
		field.Int32("node_count").
			Optional(),
		field.Int32("processing_units").
			Optional(),
		field.Int("state").
			Optional().
			Comment("Instance state enum value"),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "backend"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// EndpointURIsJSON contains the endpoint URIs for the instance.
		//
		//	["https://spanner.googleapis.com"]
		field.JSON("endpoint_uris_json", json.RawMessage{}).
			Optional(),

		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.Int("edition").
			Optional().
			Comment("Instance edition enum value"),
		field.Int("default_backup_schedule_type").
			Optional().
			Comment("Default backup schedule type enum value"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPSpannerInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("databases", BronzeGCPSpannerDatabase.Type),
	}
}

func (BronzeGCPSpannerInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPSpannerInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_spanner_instances"},
	}
}
