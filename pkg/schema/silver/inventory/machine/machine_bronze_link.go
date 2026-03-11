package machine

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// InventoryMachineBronzeLink tracks which bronze records contributed to a machine.
type InventoryMachineBronzeLink struct {
	ent.Schema
}

func (InventoryMachineBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (InventoryMachineBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("machine", InventoryMachine.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (InventoryMachineBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_machine_links"},
	}
}
