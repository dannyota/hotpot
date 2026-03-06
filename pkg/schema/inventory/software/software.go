package software

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/inventory/mixin"
)

// InventorySoftware is the final merged installed software table.
type InventorySoftware struct {
	ent.Schema
}

func (InventorySoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (InventorySoftware) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("machine_id").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("version").Optional(),
		field.String("publisher").Optional(),
	}
}

func (InventorySoftware) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bronze_links", InventorySoftwareBronzeLink.Type),
	}
}

func (InventorySoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("machine_id"),
		index.Fields("name"),
		index.Fields("machine_id", "name").Unique(),
	}
}

func (InventorySoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "software"},
	}
}
