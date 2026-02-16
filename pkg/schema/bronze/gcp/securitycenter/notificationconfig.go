package securitycenter

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPSecurityCenterNotificationConfig represents a Security Command Center
// notification config in the bronze layer.
// Fields preserve raw API response data from securitycenter.ListNotificationConfigs.
type BronzeGCPSecurityCenterNotificationConfig struct {
	ent.Schema
}

func (BronzeGCPSecurityCenterNotificationConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSecurityCenterNotificationConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Notification config resource name (e.g., organizations/123/notificationConfigs/456)"),
		field.String("name").
			NotEmpty().
			Comment("Notification config resource name"),
		field.String("description").
			Optional().
			Comment("Description of the notification config (max 1024 characters)"),
		field.String("pubsub_topic").
			Optional().
			Comment("Pub/Sub topic for notifications (e.g., projects/123/topics/my-topic)"),
		field.String("streaming_config_json").
			Optional().
			Comment("StreamingConfig filter expression as JSON"),
		field.String("service_account").
			Optional().
			Comment("Service account for Pub/Sub publishing"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPSecurityCenterNotificationConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
	}
}

func (BronzeGCPSecurityCenterNotificationConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_securitycenter_notification_configs"},
	}
}
