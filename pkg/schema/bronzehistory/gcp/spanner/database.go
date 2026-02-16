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

// BronzeHistoryGCPSpannerDatabase stores historical snapshots of GCP Spanner databases.
type BronzeHistoryGCPSpannerDatabase struct {
	ent.Schema
}

func (BronzeHistoryGCPSpannerDatabase) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSpannerDatabase) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Spanner database by resource_id"),

		// All database fields
		field.String("name").
			NotEmpty(),
		field.Int("state").
			Optional(),
		field.String("create_time").
			Optional(),

		// JSONB fields
		field.JSON("restore_info_json", json.RawMessage{}).
			Optional(),
		field.JSON("encryption_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("encryption_info_json", json.RawMessage{}).
			Optional(),

		field.String("version_retention_period").
			Optional(),
		field.String("earliest_version_time").
			Optional(),
		field.String("default_leader").
			Optional(),
		field.Int("database_dialect").
			Optional(),
		field.Bool("enable_drop_protection").
			Default(false),
		field.Bool("reconciling").
			Default(false),

		// Collection metadata
		field.String("instance_name").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPSpannerDatabase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPSpannerDatabase) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_spanner_databases_history"},
	}
}
