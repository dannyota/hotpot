package spanner

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPSpannerInstance stores historical snapshots of GCP Spanner instances.
type BronzeHistoryGCPSpannerInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPSpannerInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSpannerInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Spanner instance by resource_id"),

		// All instance fields
		field.String("name").
			NotEmpty(),
		field.String("config").
			Optional(),
		field.String("display_name").
			Optional(),
		field.Int32("node_count").
			Optional(),
		field.Int32("processing_units").
			Optional(),
		field.Int("state").
			Optional(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("endpoint_uris_json", json.RawMessage{}).
			Optional(),

		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.Int("edition").
			Optional(),
		field.Int("default_backup_schedule_type").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPSpannerInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPSpannerInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_spanner_instances_history"},
	}
}
