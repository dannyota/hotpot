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

// BronzeHistoryGCPLoggingBucket stores historical snapshots of GCP Cloud Logging buckets.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPLoggingBucket struct {
	ent.Schema
}

func (BronzeHistoryGCPLoggingBucket) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPLoggingBucket) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze bucket by resource_id"),

		// All bucket fields (same as bronze.BronzeGCPLoggingBucket)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Int32("retention_days").
			Default(0),
		field.Bool("locked").
			Default(false),
		field.String("lifecycle_state").
			Optional(),
		field.Bool("analytics_enabled").
			Default(false),
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),

		// JSONB fields
		field.JSON("cmek_settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("index_configs_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPLoggingBucket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPLoggingBucket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_buckets_history"},
	}
}
