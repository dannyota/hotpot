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

// BronzeHistoryS1RangerDevice stores historical snapshots of SentinelOne Ranger devices.
type BronzeHistoryS1RangerDevice struct {
	ent.Schema
}

func (BronzeHistoryS1RangerDevice) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1RangerDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze ranger device by resource_id"),

		field.String("local_ip").
			Optional(),
		field.String("external_ip").
			Optional(),
		field.String("mac_address").
			Optional(),
		field.String("os_type").
			Optional(),
		field.String("os_name").
			Optional(),
		field.String("os_version").
			Optional(),
		field.String("device_type").
			Optional(),
		field.String("device_function").
			Optional(),
		field.String("manufacturer").
			Optional(),
		field.String("managed_state").
			Optional(),
		field.String("agent_id").
			Optional(),
		field.Time("first_seen").
			Optional().
			Nillable(),
		field.Time("last_seen").
			Optional().
			Nillable(),
		field.String("subnet_address").
			Optional(),
		field.String("gateway_ip_address").
			Optional(),
		field.String("gateway_mac_address").
			Optional(),
		field.String("network_name").
			Optional(),
		field.String("domain").
			Optional(),
		field.String("site_name").
			Optional(),
		field.String("device_review").
			Optional(),
		field.Bool("has_identity").
			Default(false),
		field.Bool("has_user_label").
			Default(false),
		field.Int("fingerprint_score").
			Optional(),
		field.JSON("tcp_ports_json", json.RawMessage{}).
			Optional(),
		field.JSON("udp_ports_json", json.RawMessage{}).
			Optional(),
		field.JSON("hostnames_json", json.RawMessage{}).
			Optional(),
		field.JSON("discovery_methods_json", json.RawMessage{}).
			Optional(),
		field.JSON("networks_json", json.RawMessage{}).
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryS1RangerDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("managed_state"),
		index.Fields("agent_id"),
	}
}

func (BronzeHistoryS1RangerDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_ranger_devices_history"},
	}
}
