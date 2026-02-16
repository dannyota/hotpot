package pubsub

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPPubSubTopic stores historical snapshots of GCP Pub/Sub topics.
type BronzeHistoryGCPPubSubTopic struct {
	ent.Schema
}

func (BronzeHistoryGCPPubSubTopic) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPPubSubTopic) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze topic by resource_id"),

		field.String("name").
			NotEmpty(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("message_storage_policy_json", json.RawMessage{}).
			Optional(),

		field.String("kms_key_name").
			Optional(),

		field.JSON("schema_settings_json", json.RawMessage{}).
			Optional(),

		field.String("message_retention_duration").
			Optional(),

		field.Int("state").
			Optional(),

		field.JSON("ingestion_data_source_settings_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPPubSubTopic) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPPubSubTopic) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_pubsub_topics_history"},
	}
}
