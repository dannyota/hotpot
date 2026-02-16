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

// BronzeGCPMonitoringUptimeCheckConfig represents a GCP Monitoring uptime check config in the bronze layer.
// Fields preserve raw API response data from monitoring.ListUptimeCheckConfigs.
type BronzeGCPMonitoringUptimeCheckConfig struct {
	ent.Schema
}

func (BronzeGCPMonitoringUptimeCheckConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPMonitoringUptimeCheckConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Uptime check config resource name (e.g., projects/123/uptimeCheckConfigs/456)"),
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),

		// MonitoredResourceJSON contains the monitored resource to check.
		//
		//	{"type": "uptime_url", "labels": {"host": "example.com"}}
		field.JSON("monitored_resource_json", json.RawMessage{}).
			Optional(),

		// ResourceGroupJSON contains the resource group to check.
		//
		//	{"groupId": "...", "resourceType": "INSTANCE"}
		field.JSON("resource_group_json", json.RawMessage{}).
			Optional(),

		// HttpCheckJSON contains the HTTP check configuration.
		//
		//	{"requestMethod": "GET", "path": "/", "port": 443, "useSsl": true, ...}
		field.JSON("http_check_json", json.RawMessage{}).
			Optional(),

		// TcpCheckJSON contains the TCP check configuration.
		//
		//	{"port": 443}
		field.JSON("tcp_check_json", json.RawMessage{}).
			Optional(),

		// Period as a duration string (e.g., "60s", "300s").
		field.String("period").
			Optional(),

		// Timeout as a duration string (e.g., "10s").
		field.String("timeout").
			Optional(),

		// ContentMatchersJSON contains the expected content matchers.
		//
		//	[{"content": "...", "matcher": "CONTAINS_STRING"}]
		field.JSON("content_matchers_json", json.RawMessage{}).
			Optional(),

		// CheckerType is the UptimeCheckConfig_CheckerType enum value.
		// 0=CHECKER_TYPE_UNSPECIFIED, 1=STATIC_IP_CHECKERS, 2=VPC_CHECKERS
		field.Int("checker_type").
			Default(0),

		// SelectedRegionsJSON contains the list of UptimeCheckRegion enum values.
		//
		//	[1, 2, 4]
		field.JSON("selected_regions_json", json.RawMessage{}).
			Optional(),

		field.Bool("is_internal").
			Default(false),

		// InternalCheckersJSON contains deprecated internal checker configs.
		//
		//	[{"name": "...", "displayName": "...", "network": "...", ...}]
		field.JSON("internal_checkers_json", json.RawMessage{}).
			Optional(),

		// UserLabelsJSON contains user-supplied key/value labels.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("user_labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPMonitoringUptimeCheckConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPMonitoringUptimeCheckConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_monitoring_uptime_check_configs"},
	}
}
