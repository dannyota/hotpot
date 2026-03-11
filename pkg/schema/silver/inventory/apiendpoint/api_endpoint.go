package apiendpoint

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/silver/inventory/mixin"
)

// InventoryApiEndpoint is the normalized API endpoint catalog in the silver layer.
type InventoryApiEndpoint struct {
	ent.Schema
}

func (InventoryApiEndpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (InventoryApiEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("name").Optional(),
		field.String("service").
			Optional().
			Comment("Normalized upstream code, e.g. \"dbs\""),
		field.String("uri_pattern").
			NotEmpty().
			Comment("URI path pattern for matching"),
		field.JSON("methods", []string{}).
			Optional().
			Comment("HTTP methods, e.g. [\"POST\", \"PUT\"]"),
		field.Bool("is_active").Default(true),
		field.String("access_level").
			Optional().
			Comment("Derived from URI prefix: public, protected, private"),
	}
}

func (InventoryApiEndpoint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bronze_links", InventoryApiEndpointBronzeLink.Type),
	}
}

func (InventoryApiEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uri_pattern"),
		index.Fields("service"),
		index.Fields("access_level"),
	}
}

func (InventoryApiEndpoint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_api_endpoints"},
	}
}
