package monitoring

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPMonitoringAlertPolicy represents a GCP Monitoring alert policy in the bronze layer.
// Fields preserve raw API response data from monitoring.ListAlertPolicies.
type BronzeGCPMonitoringAlertPolicy struct {
	ent.Schema
}

func (BronzeGCPMonitoringAlertPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPMonitoringAlertPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Alert policy resource name (e.g., projects/123/alertPolicies/456)"),
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),

		// DocumentationJSON contains the documentation included with notifications.
		//
		//	{"content": "...", "mimeType": "text/markdown", ...}
		field.JSON("documentation_json", json.RawMessage{}).
			Optional(),

		// UserLabelsJSON contains user-supplied key/value labels.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("user_labels_json", json.RawMessage{}).
			Optional(),

		// ConditionsJSON contains the list of conditions for the policy.
		//
		//	[{"name": "...", "displayName": "...", "conditionThreshold": {...}}, ...]
		field.JSON("conditions_json", json.RawMessage{}).
			Optional(),

		// Combiner is the ConditionCombinerType enum value.
		// 0=COMBINE_UNSPECIFIED, 1=AND, 2=OR, 3=AND_WITH_MATCHING_RESOURCE
		field.Int("combiner").
			Default(0),

		field.Bool("enabled").
			Default(false),

		// NotificationChannelsJSON contains the list of notification channel names.
		//
		//	["projects/123/notificationChannels/456", ...]
		field.JSON("notification_channels_json", json.RawMessage{}).
			Optional(),

		// CreationRecordJSON contains the MutationRecord for when the policy was created.
		//
		//	{"mutateTime": "...", "mutatedBy": "..."}
		field.JSON("creation_record_json", json.RawMessage{}).
			Optional(),

		// MutationRecordJSON contains the MutationRecord for the most recent change.
		//
		//	{"mutateTime": "...", "mutatedBy": "..."}
		field.JSON("mutation_record_json", json.RawMessage{}).
			Optional(),

		// AlertStrategyJSON contains notification rate control and auto-close settings.
		//
		//	{"notificationRateLimit": {...}, "autoClose": "..."}
		field.JSON("alert_strategy_json", json.RawMessage{}).
			Optional(),

		// Severity is the AlertPolicy_Severity enum value.
		// 0=SEVERITY_UNSPECIFIED, 1=CRITICAL, 2=ERROR, 3=WARNING
		field.Int("severity").
			Default(0),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPMonitoringAlertPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPMonitoringAlertPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_monitoring_alert_policies"},
	}
}
