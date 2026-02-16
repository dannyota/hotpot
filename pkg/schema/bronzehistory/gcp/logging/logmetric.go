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

// BronzeHistoryGCPLoggingLogMetric stores historical snapshots of GCP Cloud Logging log-based metrics.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPLoggingLogMetric struct {
	ent.Schema
}

func (BronzeHistoryGCPLoggingLogMetric) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPLoggingLogMetric) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze log metric by resource_id"),

		// All log metric fields (same as bronze.BronzeGCPLoggingLogMetric)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Text("filter").
			Optional(),

		// JSONB fields
		field.JSON("metric_descriptor_json", json.RawMessage{}).
			Optional(),
		field.JSON("label_extractors_json", json.RawMessage{}).
			Optional(),
		field.JSON("bucket_options_json", json.RawMessage{}).
			Optional(),

		field.String("value_extractor").
			Optional(),
		field.String("version").
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

func (BronzeHistoryGCPLoggingLogMetric) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPLoggingLogMetric) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_log_metrics_history"},
	}
}
