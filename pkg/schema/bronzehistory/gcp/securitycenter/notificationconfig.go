package securitycenter

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPSecurityCenterNotificationConfig stores historical snapshots
// of SCC notification configs.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPSecurityCenterNotificationConfig struct {
	ent.Schema
}

func (BronzeHistoryGCPSecurityCenterNotificationConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSecurityCenterNotificationConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze SCC notification config by resource_id"),

		// All notification config fields (same as bronze.BronzeGCPSecurityCenterNotificationConfig)
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
			NotEmpty(),
	}
}

func (BronzeHistoryGCPSecurityCenterNotificationConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPSecurityCenterNotificationConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_securitycenter_notification_configs_history"},
	}
}
