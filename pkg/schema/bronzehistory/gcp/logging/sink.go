package logging

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPLoggingSink stores historical snapshots of GCP Cloud Logging sinks.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPLoggingSink struct {
	ent.Schema
}

func (BronzeHistoryGCPLoggingSink) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPLoggingSink) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze sink by resource_id"),

		// All sink fields (same as bronze.BronzeGCPLoggingSink)
		field.String("name").
			NotEmpty(),
		field.String("destination").
			Optional(),
		field.Text("filter").
			Optional(),
		field.String("description").
			Optional(),
		field.Bool("disabled").
			Default(false),
		field.Bool("include_children").
			Default(false),
		field.String("writer_identity").
			Optional(),

		// JSONB fields
		field.JSON("exclusions_json", json.RawMessage{}).
			Optional(),
		field.JSON("bigquery_options_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPLoggingSink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPLoggingSink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_sinks_history"},
	}
}
