package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1Agent represents a SentinelOne agent/endpoint in the bronze layer.
type BronzeS1Agent struct {
	ent.Schema
}

func (BronzeS1Agent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1Agent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne agent ID"),
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
		field.Int("threat_count").
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

		// NetworkInterfacesJSON stores a snapshot of all NICs as JSONB for quick access.
		//
		//	[{"id": "...", "name": "eth0", "inet": ["10.0.0.1"], ...}]
		field.JSON("network_interfaces_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeS1Agent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("nics", BronzeS1AgentNIC.Type),
	}
}

func (BronzeS1Agent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id"),
		index.Fields("is_active"),
		index.Fields("is_infected"),
		index.Fields("os_type"),
		index.Fields("collected_at"),
	}
}

func (BronzeS1Agent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_agents"},
	}
}

// BronzeS1AgentNIC represents a network interface on a SentinelOne agent.
type BronzeS1AgentNIC struct {
	ent.Schema
}

func (BronzeS1AgentNIC) Fields() []ent.Field {
	return []ent.Field{
		field.String("interface_id").
			Optional(),
		field.String("name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("type").
			Optional(),

		// InetJSON stores IPv4 addresses as JSONB array.
		//
		//	["10.0.0.1", "10.0.0.2"]
		field.JSON("inet_json", json.RawMessage{}).
			Optional(),

		// Inet6JSON stores IPv6 addresses as JSONB array.
		//
		//	["fe80::1", "::1"]
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

func (BronzeS1AgentNIC) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("agent", BronzeS1Agent.Type).
			Ref("nics").
			Unique().
			Required(),
	}
}

func (BronzeS1AgentNIC) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_agent_nics"},
	}
}
