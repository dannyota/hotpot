package monitoring

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPMonitoringAlertPolicy stores historical snapshots of GCP Monitoring alert policies.
type BronzeHistoryGCPMonitoringAlertPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPMonitoringAlertPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPMonitoringAlertPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze alert policy by resource_id"),

		// All alert policy fields (same as bronze.BronzeGCPMonitoringAlertPolicy)
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.JSON("documentation_json", json.RawMessage{}).
			Optional(),
		field.JSON("user_labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("conditions_json", json.RawMessage{}).
			Optional(),
		field.Int("combiner").
			Default(0),
		field.Bool("enabled").
			Default(false),
		field.JSON("notification_channels_json", json.RawMessage{}).
			Optional(),
		field.JSON("creation_record_json", json.RawMessage{}).
			Optional(),
		field.JSON("mutation_record_json", json.RawMessage{}).
			Optional(),
		field.JSON("alert_strategy_json", json.RawMessage{}).
			Optional(),
		field.Int("severity").
			Default(0),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPMonitoringAlertPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPMonitoringAlertPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_monitoring_alert_policies_history"},
	}
}
