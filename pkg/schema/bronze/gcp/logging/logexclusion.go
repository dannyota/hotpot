package logging

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPLoggingLogExclusion represents a GCP Cloud Logging log exclusion in the bronze layer.
// Fields preserve raw API response data from logging.projects.exclusions.list.
type BronzeGCPLoggingLogExclusion struct {
	ent.Schema
}

func (BronzeGCPLoggingLogExclusion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPLoggingLogExclusion) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP log exclusion name, used as primary key for linking"),
		field.String("name").
			NotEmpty().
			Comment("The client-assigned exclusion identifier"),
		field.String("description").
			Optional().
			Comment("User-supplied description of the exclusion"),
		field.Text("filter").
			Optional().
			Comment("Advanced logs filter for matching log entries to exclude"),
		field.Bool("disabled").
			Default(false).
			Comment("Whether the exclusion is disabled"),
		field.String("create_time").
			Optional().
			Comment("Timestamp when the exclusion was created"),
		field.String("update_time").
			Optional().
			Comment("Timestamp when the exclusion was last updated"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPLoggingLogExclusion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPLoggingLogExclusion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_log_exclusions"},
	}
}
