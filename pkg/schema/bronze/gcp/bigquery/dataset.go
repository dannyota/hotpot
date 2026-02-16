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

// BronzeGCPBigQueryDataset represents a GCP BigQuery dataset in the bronze layer.
// Fields preserve raw API response data from bigquery.Dataset.Metadata.
type BronzeGCPBigQueryDataset struct {
	ent.Schema
}

func (BronzeGCPBigQueryDataset) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPBigQueryDataset) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Dataset resource name (projects/{project}/datasets/{dataset})"),
		field.String("friendly_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("location").
			Optional(),
		field.Int64("default_table_expiration_ms").
			Optional().
			Nillable().
			Comment("Default table expiration in milliseconds"),
		field.Int64("default_partition_expiration_ms").
			Optional().
			Nillable().
			Comment("Default partition expiration in milliseconds"),

		// LabelsJSON contains user-provided labels.
		//
		//	{"env": "prod", "team": "analytics"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// AccessJSON contains access control entries for the dataset.
		//
		//	[{"role": "OWNER", "userByEmail": "user@example.com"}, ...]
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
			Nillable().
			Comment("Max time travel duration in hours"),

		// DefaultEncryptionConfigurationJSON contains CMEK configuration.
		//
		//	{"kmsKeyName": "projects/.../cryptoKeys/..."}
		field.JSON("default_encryption_configuration_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPBigQueryDataset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPBigQueryDataset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigquery_datasets"},
	}
}
