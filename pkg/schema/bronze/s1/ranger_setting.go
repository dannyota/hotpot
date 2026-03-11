package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1RangerSetting represents SentinelOne Ranger settings in the bronze layer.
type BronzeS1RangerSetting struct {
	ent.Schema
}

func (BronzeS1RangerSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1RangerSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Account ID used as resource ID for Ranger settings"),
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

func (BronzeS1RangerSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeS1RangerSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_ranger_settings"},
	}
}
