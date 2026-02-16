package logging

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPLoggingLogMetric represents a GCP Cloud Logging log-based metric in the bronze layer.
// Fields preserve raw API response data from logging.projects.metrics.list.
type BronzeGCPLoggingLogMetric struct {
	ent.Schema
}

func (BronzeGCPLoggingLogMetric) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPLoggingLogMetric) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP log metric name, used as primary key for linking"),
		field.String("name").
			NotEmpty().
			Comment("The client-assigned metric identifier"),
		field.String("description").
			Optional().
			Comment("User-supplied description of the metric"),
		field.Text("filter").
			Optional().
			Comment("Advanced logs filter for matching log entries"),

		// MetricDescriptorJSON contains the metric descriptor definition.
		//
		//	{"name": "...", "metricKind": "DELTA", "valueType": "INT64", ...}
		field.JSON("metric_descriptor_json", json.RawMessage{}).
			Optional(),

		// LabelExtractorsJSON contains label extractor expressions.
		//
		//	{"label_name": "REGEXP_EXTRACT(jsonPayload.request, \"...\")"}
		field.JSON("label_extractors_json", json.RawMessage{}).
			Optional(),

		// BucketOptionsJSON contains distribution bucket options.
		//
		//	{"linearBuckets": {"numFiniteBuckets": 3, "width": 1, "offset": 1}}
		field.JSON("bucket_options_json", json.RawMessage{}).
			Optional(),

		field.String("value_extractor").
			Optional().
			Comment("Value extractor expression for distribution metrics"),
		field.String("version").
			Optional().
			Comment("Logging API version for the metric"),
		field.Bool("disabled").
			Default(false).
			Comment("Whether the metric is disabled"),
		field.String("create_time").
			Optional().
			Comment("Creation timestamp from GCP API"),
		field.String("update_time").
			Optional().
			Comment("Last update timestamp from GCP API"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPLoggingLogMetric) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPLoggingLogMetric) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_log_metrics"},
	}
}
