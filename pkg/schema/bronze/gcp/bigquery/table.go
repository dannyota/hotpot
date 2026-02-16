package bigquery

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPBigQueryTable represents a GCP BigQuery table in the bronze layer.
// Fields preserve raw API response data from bigquery.Table.Metadata.
type BronzeGCPBigQueryTable struct {
	ent.Schema
}

func (BronzeGCPBigQueryTable) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPBigQueryTable) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Table resource name (projects/{project}/datasets/{dataset}/tables/{table})"),
		field.String("dataset_id").
			NotEmpty().
			Comment("Parent dataset resource name"),
		field.String("friendly_name").
			Optional(),
		field.String("description").
			Optional(),

		// SchemaJSON contains the table schema definition.
		//
		//	{"fields": [{"name": "col1", "type": "STRING", "mode": "REQUIRED"}, ...]}
		field.JSON("schema_json", json.RawMessage{}).
			Optional(),

		field.Int64("num_bytes").
			Optional().
			Nillable().
			Comment("Size of the table in bytes"),
		field.Int64("num_long_term_bytes").
			Optional().
			Nillable().
			Comment("Bytes considered long-term storage for billing"),
		field.Uint64("num_rows").
			Optional().
			Nillable().
			Comment("Number of rows in the table"),
		field.String("creation_time").
			Optional(),
		field.String("expiration_time").
			Optional(),
		field.String("last_modified_time").
			Optional(),
		field.String("table_type").
			Optional().
			Comment("TABLE, VIEW, MATERIALIZED_VIEW, EXTERNAL, SNAPSHOT"),

		// LabelsJSON contains user-provided labels.
		//
		//	{"env": "prod", "team": "analytics"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// EncryptionConfigurationJSON contains CMEK configuration.
		//
		//	{"kmsKeyName": "projects/.../cryptoKeys/..."}
		field.JSON("encryption_configuration_json", json.RawMessage{}).
			Optional(),

		// TimePartitioningJSON contains time-based partitioning configuration.
		//
		//	{"type": "DAY", "field": "timestamp", "expirationMs": "7776000000"}
		field.JSON("time_partitioning_json", json.RawMessage{}).
			Optional(),

		// RangePartitioningJSON contains integer range partitioning configuration.
		//
		//	{"field": "id", "range": {"start": "0", "end": "1000000", "interval": "1000"}}
		field.JSON("range_partitioning_json", json.RawMessage{}).
			Optional(),

		// ClusteringJSON contains clustering configuration.
		//
		//	{"fields": ["column1", "column2"]}
		field.JSON("clustering_json", json.RawMessage{}).
			Optional(),

		field.Bool("require_partition_filter").
			Default(false).
			Comment("Whether queries must use a partition filter"),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPBigQueryTable) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("dataset_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPBigQueryTable) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigquery_tables"},
	}
}
