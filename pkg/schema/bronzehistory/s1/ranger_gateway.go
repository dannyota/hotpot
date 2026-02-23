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

// BronzeHistoryS1RangerGateway stores historical snapshots of SentinelOne Ranger gateways.
type BronzeHistoryS1RangerGateway struct {
	ent.Schema
}

func (BronzeHistoryS1RangerGateway) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1RangerGateway) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze ranger gateway by resource_id"),

		field.String("ip").
			Optional(),
		field.String("mac_address").
			Optional(),
		field.String("external_ip").
			Optional(),
		field.String("manufacturer").
			Optional(),
		field.String("network_name").
			Optional(),
		field.String("account_id").
			Optional(),
		field.String("account_name").
			Optional(),
		field.String("site_id").
			Optional(),
		field.Int("number_of_agents").
			Optional(),
		field.Int("number_of_rangers").
			Optional(),
		field.Int("connected_rangers").
			Optional(),
		field.Int("total_agents").
			Optional(),
		field.Float("agent_percentage").
			Optional(),
		field.Bool("allow_scan").
			Default(false),
		field.Bool("archived").
			Default(false),
		field.Bool("new_network").
			Default(false),
		field.Bool("inherit_settings").
			Default(false),
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
		field.Bool("scan_only_local_subnets").
			Default(false),
		field.Time("created_at_api").
			Optional().
			Nillable(),
		field.Time("expiry_date").
			Optional().
			Nillable(),
		field.JSON("tcp_ports_json", json.RawMessage{}).
			Optional(),
		field.JSON("udp_ports_json", json.RawMessage{}).
			Optional(),
		field.JSON("restrictions_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryS1RangerGateway) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("account_id"),
	}
}

func (BronzeHistoryS1RangerGateway) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_ranger_gateways_history"},
	}
}
