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

// BronzeGCPPubSubSubscription represents a GCP Pub/Sub subscription in the bronze layer.
type BronzeGCPPubSubSubscription struct {
	ent.Schema
}

func (BronzeGCPPubSubSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPPubSubSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Subscription resource name (projects/{project}/subscriptions/{subscription})"),
		field.String("name").
			NotEmpty(),
		field.String("topic").
			Optional().
			Comment("Topic resource name this subscription receives messages from"),

		// PushConfigJSON contains push delivery configuration.
		//
		//	{"pushEndpoint": "https://...", "attributes": {...}}
		field.JSON("push_config_json", json.RawMessage{}).
			Optional(),

		// BigqueryConfigJSON contains BigQuery delivery configuration.
		//
		//	{"table": "projects/.../datasets/.../tables/...", "writeMetadata": true}
		field.JSON("bigquery_config_json", json.RawMessage{}).
			Optional(),

		// CloudStorageConfigJSON contains Cloud Storage delivery configuration.
		//
		//	{"bucket": "my-bucket", "filenamePrefix": "..."}
		field.JSON("cloud_storage_config_json", json.RawMessage{}).
			Optional(),

		field.Int("ack_deadline_seconds").
			Optional().
			Comment("Approximate time Pub/Sub waits for acknowledgment before resending"),

		field.Bool("retain_acked_messages").
			Default(false).
			Comment("Whether to retain acknowledged messages"),

		field.String("message_retention_duration").
			Optional().
			Comment("How long to retain unacknowledged messages in the backlog"),

		// LabelsJSON contains user-defined labels.
		//
		//	{"env": "prod", "team": "data"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.Bool("enable_message_ordering").
			Default(false).
			Comment("Whether messages with the same ordering_key are delivered in order"),

		// ExpirationPolicyJSON contains the subscription expiration policy.
		//
		//	{"ttl": "2678400s"}
		field.JSON("expiration_policy_json", json.RawMessage{}).
			Optional(),

		field.String("filter").
			Optional().
			Comment("Filter expression for message delivery"),

		// DeadLetterPolicyJSON contains the dead letter policy configuration.
		//
		//	{"deadLetterTopic": "projects/.../topics/...", "maxDeliveryAttempts": 5}
		field.JSON("dead_letter_policy_json", json.RawMessage{}).
			Optional(),

		// RetryPolicyJSON contains the message delivery retry policy.
		//
		//	{"minimumBackoff": "10s", "maximumBackoff": "600s"}
		field.JSON("retry_policy_json", json.RawMessage{}).
			Optional(),

		field.Bool("detached").
			Default(false).
			Comment("Whether the subscription is detached from its topic"),

		field.Bool("enable_exactly_once_delivery").
			Default(false).
			Comment("Whether exactly-once delivery is enabled"),

		field.Int("state").
			Optional().
			Comment("Subscription state (0=UNSPECIFIED, 1=ACTIVE, 2=RESOURCE_ERROR)"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPPubSubSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPPubSubSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_pubsub_subscriptions"},
	}
}
