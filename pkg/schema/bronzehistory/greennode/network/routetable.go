package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeNetworkRouteTable stores historical snapshots of GreenNode route tables.
type BronzeHistoryGreenNodeNetworkRouteTable struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkRouteTable) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeNetworkRouteTable) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.String("network_id").
			Optional(),
		field.String("created_at").
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeNetworkRouteTable) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeNetworkRouteTable) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_route_tables_history"},
	}
}

// BronzeHistoryGreenNodeNetworkRouteTableRoute stores historical snapshots of route table routes.
type BronzeHistoryGreenNodeNetworkRouteTableRoute struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkRouteTableRoute) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("route_table_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeNetworkRouteTable"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("route_id").
			NotEmpty(),
		field.String("route_table_id").
			Optional(),
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

func (BronzeHistoryGreenNodeNetworkRouteTableRoute) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("route_table_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeNetworkRouteTableRoute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_route_table_routes_history"},
	}
}
