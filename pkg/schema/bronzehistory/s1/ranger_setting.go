package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1RangerSetting stores historical snapshots of SentinelOne Ranger settings.
type BronzeHistoryS1RangerSetting struct {
	ent.Schema
}

func (BronzeHistoryS1RangerSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1RangerSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze ranger setting by resource_id (account_id)"),

		field.String("account_id").
			Optional(),
		field.String("scope_id").
			Optional(),
		field.Bool("enabled").
			Default(false),
		field.Bool("use_periodic_snapshots").
			Default(false),
		field.Int("snapshot_period").
			Optional(),
		field.Int("network_decommission_value").
			Optional(),
		field.Int("min_agents_in_network_to_scan").
			Optional(),
		field.Bool("tcp_port_scan").
			Default(false),
		field.Bool("udp_port_scan").
			Default(false),
		field.Bool("icmp_scan").
			Default(false),
		field.Bool("smb_scan").
			Default(false),
		field.Bool("mdns_scan").
			Default(false),
		field.Bool("rdns_scan").
			Default(false),
		field.Bool("snmp_scan").
			Default(false),
		field.Bool("multi_scan_ssdp").
			Default(false),
		field.Bool("use_full_dns_scan").
			Default(false),
		field.Bool("scan_only_local_subnets").
			Default(false),
		field.Bool("auto_enable_networks").
			Default(false),
		field.Bool("combine_devices").
			Default(false),
		field.Int("new_network_in_hours").
			Optional(),
		field.JSON("tcp_ports_json", json.RawMessage{}).
			Optional(),
		field.JSON("udp_ports_json", json.RawMessage{}).
			Optional(),
		field.JSON("restrictions_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryS1RangerSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("account_id"),
	}
}

func (BronzeHistoryS1RangerSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_ranger_settings_history"},
	}
}
