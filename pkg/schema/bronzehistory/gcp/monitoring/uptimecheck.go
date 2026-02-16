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

// BronzeHistoryGCPMonitoringUptimeCheckConfig stores historical snapshots of GCP Monitoring uptime check configs.
type BronzeHistoryGCPMonitoringUptimeCheckConfig struct {
	ent.Schema
}

func (BronzeHistoryGCPMonitoringUptimeCheckConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPMonitoringUptimeCheckConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze uptime check config by resource_id"),

		// All uptime check config fields (same as bronze.BronzeGCPMonitoringUptimeCheckConfig)
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.JSON("monitored_resource_json", json.RawMessage{}).
			Optional(),
		field.JSON("resource_group_json", json.RawMessage{}).
			Optional(),
		field.JSON("http_check_json", json.RawMessage{}).
			Optional(),
		field.JSON("tcp_check_json", json.RawMessage{}).
			Optional(),
		field.String("period").
			Optional(),
		field.String("timeout").
			Optional(),
		field.JSON("content_matchers_json", json.RawMessage{}).
			Optional(),
		field.Int("checker_type").
			Default(0),
		field.JSON("selected_regions_json", json.RawMessage{}).
			Optional(),
		field.Bool("is_internal").
			Default(false),
		field.JSON("internal_checkers_json", json.RawMessage{}).
			Optional(),
		field.JSON("user_labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPMonitoringUptimeCheckConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPMonitoringUptimeCheckConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_monitoring_uptime_check_configs_history"},
	}
}
