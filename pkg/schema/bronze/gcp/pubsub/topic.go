package pubsub

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPPubSubTopic represents a GCP Pub/Sub topic in the bronze layer.
type BronzeGCPPubSubTopic struct {
	ent.Schema
}

func (BronzeGCPPubSubTopic) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPPubSubTopic) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Topic resource name (projects/{project}/topics/{topic})"),
		field.String("name").
			NotEmpty(),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "data"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// MessageStoragePolicyJSON contains the policy constraining regions
		// where messages may be stored.
		//
		//	{"allowedPersistenceRegions": ["us-east1", "us-west1"]}
		field.JSON("message_storage_policy_json", json.RawMessage{}).
			Optional(),

		field.String("kms_key_name").
			Optional().
			Comment("Cloud KMS CryptoKey resource name for message encryption"),

		// SchemaSettingsJSON contains settings for validating messages
		// published against a schema.
		//
		//	{"schema": "projects/.../schemas/...", "encoding": "JSON"}
		field.JSON("schema_settings_json", json.RawMessage{}).
			Optional(),

		field.String("message_retention_duration").
			Optional().
			Comment("Minimum duration to retain a message after publishing"),

		field.Int("state").
			Optional().
			Comment("Topic state (0=UNSPECIFIED, 1=ACTIVE, 2=INGESTION_RESOURCE_ERROR)"),

		// IngestionDataSourceSettingsJSON contains settings for ingestion
		// from a data source into this topic.
		//
		//	{"awsKinesis": {...}} or {"cloudStorage": {...}}
		field.JSON("ingestion_data_source_settings_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPPubSubTopic) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPPubSubTopic) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_pubsub_topics"},
	}
}
