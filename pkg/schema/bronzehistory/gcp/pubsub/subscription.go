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

// BronzeHistoryGCPPubSubSubscription stores historical snapshots of GCP Pub/Sub subscriptions.
type BronzeHistoryGCPPubSubSubscription struct {
	ent.Schema
}

func (BronzeHistoryGCPPubSubSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPPubSubSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze subscription by resource_id"),

		field.String("name").
			NotEmpty(),
		field.String("topic").
			Optional(),

		// JSONB fields
		field.JSON("push_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("bigquery_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("cloud_storage_config_json", json.RawMessage{}).
			Optional(),

		field.Int("ack_deadline_seconds").
			Optional(),

		field.Bool("retain_acked_messages").
			Default(false),

		field.String("message_retention_duration").
			Optional(),

		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.Bool("enable_message_ordering").
			Default(false),

		field.JSON("expiration_policy_json", json.RawMessage{}).
			Optional(),

		field.String("filter").
			Optional(),

		field.JSON("dead_letter_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("retry_policy_json", json.RawMessage{}).
			Optional(),

		field.Bool("detached").
			Default(false),

		field.Bool("enable_exactly_once_delivery").
			Default(false),

		field.Int("state").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPPubSubSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPPubSubSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_pubsub_subscriptions_history"},
	}
}
