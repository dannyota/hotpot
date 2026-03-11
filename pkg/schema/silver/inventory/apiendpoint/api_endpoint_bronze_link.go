package apiendpoint

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// InventoryApiEndpointBronzeLink tracks which bronze records contributed to an endpoint.
type InventoryApiEndpointBronzeLink struct {
	ent.Schema
}

func (InventoryApiEndpointBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (InventoryApiEndpointBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api_endpoint", InventoryApiEndpoint.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (InventoryApiEndpointBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_api_endpoint_links"},
	}
}
