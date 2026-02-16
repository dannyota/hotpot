package storage

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"encoding/json"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPStorageBucket stores historical snapshots of GCP Storage buckets.
type BronzeHistoryGCPStorageBucket struct {
	ent.Schema
}

func (BronzeHistoryGCPStorageBucket) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPStorageBucket) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze bucket by resource_id"),

		// All bucket fields
		field.String("name").
			NotEmpty(),
		field.String("location").
			Optional(),
		field.String("storage_class").
			Optional(),
		field.String("project_number").
			Optional(),
		field.String("time_created").
			Optional(),
		field.String("updated").
			Optional(),
		field.Bool("default_event_based_hold").
			Default(false),
		field.String("metageneration").
			Optional(),
		field.String("etag").
			Optional(),

		// JSONB fields
		field.JSON("iam_configuration_json", json.RawMessage{}).
			Optional(),
		field.JSON("encryption_json", json.RawMessage{}).
			Optional(),
		field.JSON("lifecycle_json", json.RawMessage{}).
			Optional(),
		field.JSON("versioning_json", json.RawMessage{}).
			Optional(),
		field.JSON("retention_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("logging_json", json.RawMessage{}).
			Optional(),
		field.JSON("cors_json", json.RawMessage{}).
			Optional(),
		field.JSON("website_json", json.RawMessage{}).
			Optional(),
		field.JSON("autoclass_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPStorageBucket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPStorageBucket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_buckets_history"},
	}
}

// BronzeHistoryGCPStorageBucketLabel stores historical snapshots of bucket labels.
type BronzeHistoryGCPStorageBucketLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPStorageBucketLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("bucket_history_id").
			Comment("Links to parent BronzeHistoryGCPStorageBucket"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPStorageBucketLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("bucket_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPStorageBucketLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_bucket_labels_history"},
	}
}
