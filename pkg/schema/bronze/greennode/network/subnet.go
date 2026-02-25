package network

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeNetworkSubnet represents a GreenNode network subnet in the bronze layer.
type BronzeGreenNodeNetworkSubnet struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkSubnet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkSubnet) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Subnet ID"),
		field.String("name").
			NotEmpty(),
		field.String("network_id").
			NotEmpty(),
		field.String("cidr").
			Optional(),
		field.String("status").
			Optional(),
		field.String("route_table_id").
			Optional(),
		field.String("interface_acl_policy_id").
			Optional(),
		field.String("interface_acl_policy_name").
			Optional(),
		field.String("zone_id").
			Optional(),
		field.JSON("secondary_subnets", json.RawMessage{}).
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkSubnet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
		index.Fields("network_id"),
	}
}

func (BronzeGreenNodeNetworkSubnet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_subnets"},
	}
}
