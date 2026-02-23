package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1RangerDevice represents a SentinelOne Ranger discovered device in the bronze layer.
type BronzeS1RangerDevice struct {
	ent.Schema
}

func (BronzeS1RangerDevice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1RangerDevice) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne Ranger device ID"),
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

func (BronzeS1RangerDevice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("managed_state"),
		index.Fields("agent_id"),
		index.Fields("collected_at"),
		index.Fields("site_name"),
		index.Fields("network_name"),
	}
}

func (BronzeS1RangerDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_ranger_devices"},
	}
}
