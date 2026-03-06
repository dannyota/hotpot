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

// BronzeGreenNodeNetworkRouteTable represents a GreenNode network route table in the bronze layer.
type BronzeGreenNodeNetworkRouteTable struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkRouteTable) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkRouteTable) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Route table ID"),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.String("network_id").
			Optional(),
		field.String("created_at").
			Optional().
			Comment("API creation timestamp"),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkRouteTable) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("routes", BronzeGreenNodeNetworkRouteTableRoute.Type),
	}
}

func (BronzeGreenNodeNetworkRouteTable) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeNetworkRouteTable) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_route_tables"},
	}
}

// BronzeGreenNodeNetworkRouteTableRoute represents a route within a route table.
type BronzeGreenNodeNetworkRouteTableRoute struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkRouteTableRoute) Fields() []ent.Field {
	return []ent.Field{
		field.String("route_id").
			NotEmpty().
			Comment("SDK route ID"),
		field.String("routing_type").
			Optional(),
		field.String("destination_cidr_block").
			Optional(),
		field.String("target").
			Optional(),
		field.String("status").
			Optional(),
	}
}

func (BronzeGreenNodeNetworkRouteTableRoute) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("route_table", BronzeGreenNodeNetworkRouteTable.Type).
			Ref("routes").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeNetworkRouteTableRoute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_route_table_routes"},
	}
}
