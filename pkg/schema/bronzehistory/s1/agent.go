package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1Agent stores historical snapshots of SentinelOne agents.
type BronzeHistoryS1Agent struct {
	ent.Schema
}

func (BronzeHistoryS1Agent) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1Agent) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze agent by resource_id"),

		field.String("computer_name").
			NotEmpty(),
		field.String("external_ip").
			Optional(),
		field.String("site_name").
			Optional(),
		field.String("account_id").
			Optional(),
		field.String("account_name").
			Optional(),
		field.String("agent_version").
			Optional(),
		field.String("os_type").
			Optional(),
		field.String("os_name").
			Optional(),
		field.String("os_revision").
			Optional(),
		field.String("os_arch").
			Optional(),
		field.Bool("is_active").
			Default(false),
		field.Bool("is_infected").
			Default(false),
		field.Bool("is_decommissioned").
			Default(false),
		field.String("machine_type").
			Optional(),
		field.String("domain").
			Optional(),
		field.String("uuid").
			Optional(),
		field.String("network_status").
			Optional(),
		field.Time("last_active_date").
			Optional().
			Nillable(),
		field.Time("registered_at").
			Optional().
			Nillable(),
		field.Time("api_updated_at").
			Optional().
			Nillable(),
		field.Time("os_start_time").
			Optional().
			Nillable(),
		field.Int("active_threats").
			Default(0),
		field.Bool("encrypted_applications").
			Default(false),
		field.String("group_name").
			Optional(),
		field.String("group_id").
			Optional(),
		field.Int("cpu_count").
			Default(0),
		field.Int("core_count").
			Default(0),
		field.String("cpu_id").
			Optional(),
		field.Int64("total_memory").
			Default(0),
		field.String("model_name").
			Optional(),
		field.String("serial_number").
			Optional(),
		field.String("storage_encryption_status").
			Optional(),
		field.JSON("network_interfaces_json", json.RawMessage{}).
			Optional(),
		field.String("site_id").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.String("os_username").
			Optional(),
		field.String("group_ip").
			Optional(),
		field.String("scan_status").
			Optional(),
		field.Time("scan_started_at").
			Optional().
			Nillable(),
		field.Time("scan_finished_at").
			Optional().
			Nillable(),
		field.String("mitigation_mode").
			Optional(),
		field.String("mitigation_mode_suspicious").
			Optional(),
		field.String("last_logged_in_user_name").
			Optional(),
		field.String("installer_type").
			Optional(),
		field.String("external_id").
			Optional(),
		field.String("last_ip_to_mgmt").
			Optional(),
		field.Bool("is_up_to_date").
			Default(false),
		field.Bool("is_pending_uninstall").
			Default(false),
		field.Bool("is_uninstalled").
			Default(false),
		field.String("apps_vulnerability_status").
			Optional(),
		field.String("console_migration_status").
			Optional(),
		field.String("ranger_version").
			Optional(),
		field.String("ranger_status").
			Optional(),
		field.JSON("active_directory_json", json.RawMessage{}).
			Optional(),
		field.JSON("locations_json", json.RawMessage{}).
			Optional(),
		field.JSON("user_actions_needed_json", json.RawMessage{}).
			Optional(),
		field.JSON("missing_permissions_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryS1Agent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("account_id"),
	}
}

func (BronzeHistoryS1Agent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_agents_history"},
	}
}

// BronzeHistoryS1AgentNIC stores historical snapshots of agent network interfaces.
type BronzeHistoryS1AgentNIC struct {
	ent.Schema
}

func (BronzeHistoryS1AgentNIC) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("agent_history_id").
			Comment("Links to parent BronzeHistoryS1Agent"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("interface_id").
			Optional(),
		field.String("name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("type").
			Optional(),
		field.JSON("inet_json", json.RawMessage{}).
			Optional(),
		field.JSON("inet6_json", json.RawMessage{}).
			Optional(),
		field.String("physical").
			Optional(),
		field.String("gateway_ip").
			Optional(),
		field.String("gateway_mac").
			Optional(),
	}
}

func (BronzeHistoryS1AgentNIC) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryS1AgentNIC) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_agent_nics_history"},
	}
}
