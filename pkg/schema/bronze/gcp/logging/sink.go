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

// BronzeGCPLoggingSink represents a GCP Cloud Logging sink in the bronze layer.
// Fields preserve raw API response data from logging.projects.sinks.list.
type BronzeGCPLoggingSink struct {
	ent.Schema
}

func (BronzeGCPLoggingSink) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPLoggingSink) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP sink name, used as primary key for linking"),
		field.String("name").
			NotEmpty().
			Comment("The client-assigned sink identifier"),
		field.String("destination").
			Optional().
			Comment("Export destination (e.g. storage bucket, BigQuery dataset, Pub/Sub topic)"),
		field.Text("filter").
			Optional().
			Comment("Advanced logs filter for matching log entries"),
		field.String("description").
			Optional().
			Comment("User-supplied description of the sink"),
		field.Bool("disabled").
			Default(false).
			Comment("Whether the sink is disabled"),
		field.Bool("include_children").
			Default(false).
			Comment("Whether to include child organizations/folders"),
		field.String("writer_identity").
			Optional().
			Comment("Service account used by the sink to write logs"),

		// ExclusionsJSON contains log exclusion filters.
		//
		//	[{"name": "exclude-debug", "filter": "severity < WARNING", "disabled": false}]
		field.JSON("exclusions_json", json.RawMessage{}).
			Optional(),

		// BigqueryOptionsJSON contains BigQuery-specific options.
		//
		//	{"usePartitionedTables": true}
		field.JSON("bigquery_options_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPLoggingSink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPLoggingSink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_logging_sinks"},
	}
}
