package bigquery

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPBigQueryTable stores historical snapshots of GCP BigQuery tables.
type BronzeHistoryGCPBigQueryTable struct {
	ent.Schema
}

func (BronzeHistoryGCPBigQueryTable) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPBigQueryTable) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze table by resource_id"),

		// All table fields
		field.String("dataset_id").
			NotEmpty().
			Comment("Parent dataset resource name"),
		field.String("friendly_name").
			Optional(),
		field.String("description").
			Optional(),

		// JSONB fields
		field.JSON("schema_json", json.RawMessage{}).
			Optional(),

		field.Int64("num_bytes").
			Optional().
			Nillable(),
		field.Int64("num_long_term_bytes").
			Optional().
			Nillable(),
		field.Uint64("num_rows").
			Optional().
			Nillable(),
		field.String("creation_time").
			Optional(),
		field.String("expiration_time").
			Optional(),
		field.String("last_modified_time").
			Optional(),
		field.String("table_type").
			Optional(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("encryption_configuration_json", json.RawMessage{}).
			Optional(),
		field.JSON("time_partitioning_json", json.RawMessage{}).
			Optional(),
		field.JSON("range_partitioning_json", json.RawMessage{}).
			Optional(),
		field.JSON("clustering_json", json.RawMessage{}).
			Optional(),

		field.Bool("require_partition_filter").
			Default(false),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPBigQueryTable) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPBigQueryTable) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigquery_tables_history"},
	}
}
