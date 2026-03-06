package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeNetworkSecgroup represents a GreenNode network security group in the bronze layer.
type BronzeGreenNodeNetworkSecgroup struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkSecgroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkSecgroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Security group ID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("status").
			Optional(),
		field.String("created_at").
			Optional().
			Comment("API creation timestamp"),
		field.Bool("is_system").
			Default(false),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkSecgroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("rules", BronzeGreenNodeNetworkSecgroupRule.Type),
	}
}

func (BronzeGreenNodeNetworkSecgroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeNetworkSecgroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_secgroups"},
	}
}

// BronzeGreenNodeNetworkSecgroupRule represents a rule within a security group.
type BronzeGreenNodeNetworkSecgroupRule struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkSecgroupRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("rule_id").
			NotEmpty().
			Comment("SDK rule ID"),
		field.String("direction").
			Optional(),
		field.String("ether_type").
			Optional(),
		field.String("protocol").
			Optional(),
		field.String("description").
			Optional(),
		field.String("remote_ip_prefix").
			Optional(),
		field.Int("port_range_max").
			Optional(),
		field.Int("port_range_min").
			Optional(),
	}
}

func (BronzeGreenNodeNetworkSecgroupRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("secgroup", BronzeGreenNodeNetworkSecgroup.Type).
			Ref("rules").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeNetworkSecgroupRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_secgroup_rules"},
	}
}
