package logging

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPLoggingLogExclusion stores historical snapshots of GCP Cloud Logging log exclusions.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPLoggingLogExclusion struct {
	ent.Schema
}

func (BronzeHistoryGCPLoggingLogExclusion) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPLoggingLogExclusion) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze log exclusion by resource_id"),

		// All log exclusion fields (same as bronze.BronzeGCPLoggingLogExclusion)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Text("filter").
			Optional(),
		field.Bool("disabled").
			Default(false),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPLoggingLogExclusion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPLoggingLogExclusion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_log_exclusions_history"},
	}
}
