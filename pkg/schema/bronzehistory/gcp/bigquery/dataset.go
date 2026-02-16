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

// BronzeHistoryGCPBigQueryDataset stores historical snapshots of GCP BigQuery datasets.
type BronzeHistoryGCPBigQueryDataset struct {
	ent.Schema
}

func (BronzeHistoryGCPBigQueryDataset) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPBigQueryDataset) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze dataset by resource_id"),

		// All dataset fields
		field.String("friendly_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("location").
			Optional(),
		field.Int64("default_table_expiration_ms").
			Optional().
			Nillable(),
		field.Int64("default_partition_expiration_ms").
			Optional().
			Nillable(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("access_json", json.RawMessage{}).
			Optional(),

		field.String("creation_time").
			Optional(),
		field.String("last_modified_time").
			Optional(),
		field.String("etag").
			Optional(),
		field.String("default_collation").
			Optional(),
		field.Int("max_time_travel_hours").
			Optional().
			Nillable(),

		// JSONB fields
		field.JSON("default_encryption_configuration_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPBigQueryDataset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPBigQueryDataset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigquery_datasets_history"},
	}
}
