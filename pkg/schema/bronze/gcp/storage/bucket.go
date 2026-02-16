package storage

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPStorageBucket represents a GCP Cloud Storage bucket in the bronze layer.
// Fields preserve raw API response data from storage.buckets.list.
type BronzeGCPStorageBucket struct {
	ent.Schema
}

func (BronzeGCPStorageBucket) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPStorageBucket) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Bucket name, used as primary key"),
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

		// IamConfigurationJSON contains IAM configuration (uniform bucket-level access).
		//
		//	{"uniformBucketLevelAccess": {"enabled": true}}
		field.JSON("iam_configuration_json", json.RawMessage{}).
			Optional(),

		// EncryptionJSON contains default encryption configuration.
		//
		//	{"defaultKmsKeyName": "projects/..."}
		field.JSON("encryption_json", json.RawMessage{}).
			Optional(),

		// LifecycleJSON contains object lifecycle management rules.
		//
		//	{"rule": [{"action": {"type": "Delete"}, "condition": {"age": 30}}]}
		field.JSON("lifecycle_json", json.RawMessage{}).
			Optional(),

		// VersioningJSON contains versioning configuration.
		//
		//	{"enabled": true}
		field.JSON("versioning_json", json.RawMessage{}).
			Optional(),

		// RetentionPolicyJSON contains retention policy settings.
		//
		//	{"retentionPeriod": "86400", "isLocked": false}
		field.JSON("retention_policy_json", json.RawMessage{}).
			Optional(),

		// LoggingJSON contains access log configuration.
		//
		//	{"logBucket": "bucket-name", "logObjectPrefix": "prefix"}
		field.JSON("logging_json", json.RawMessage{}).
			Optional(),

		// CorsJSON contains cross-origin resource sharing configuration.
		//
		//	[{"origin": ["*"], "method": ["GET"]}]
		field.JSON("cors_json", json.RawMessage{}).
			Optional(),

		// WebsiteJSON contains website configuration.
		//
		//	{"mainPageSuffix": "index.html", "notFoundPage": "404.html"}
		field.JSON("website_json", json.RawMessage{}).
			Optional(),

		// AutoclassJSON contains autoclass configuration.
		//
		//	{"enabled": true, "toggleTime": "2024-01-01T00:00:00Z"}
		field.JSON("autoclass_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPStorageBucket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPStorageBucketLabel.Type),
	}
}

func (BronzeGCPStorageBucket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPStorageBucket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_buckets"},
	}
}

// BronzeGCPStorageBucketLabel represents a label attached to a GCP Storage bucket.
type BronzeGCPStorageBucketLabel struct {
	ent.Schema
}

func (BronzeGCPStorageBucketLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPStorageBucketLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("bucket", BronzeGCPStorageBucket.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPStorageBucketLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_bucket_labels"},
	}
}
