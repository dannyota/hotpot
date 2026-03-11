package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeNetworkPeering represents a GreenNode network peering in the bronze layer.
type BronzeGreenNodeNetworkPeering struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkPeering) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkPeering) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Peering ID"),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.String("from_vpc_id").
			Optional(),
		field.String("from_cidr").
			Optional(),
		field.String("end_vpc_id").
			Optional(),
		field.String("end_cidr").
			Optional(),
		field.String("created_at").
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkPeering) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeNetworkPeering) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_peerings"},
	}
}
