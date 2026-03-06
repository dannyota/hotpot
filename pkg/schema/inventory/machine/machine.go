package machine

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	inventorymixin "danny.vn/hotpot/pkg/schema/inventory/mixin"
)

// InventoryMachine is the final merged machine table.
type InventoryMachine struct {
	ent.Schema
}

func (InventoryMachine) Mixin() []ent.Mixin {
	return []ent.Mixin{
		inventorymixin.Timestamp{},
	}
}

func (InventoryMachine) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("hostname").NotEmpty(),
		field.String("os_type"),
		field.String("os_name").Optional(),
		field.String("status"),
		field.String("internal_ip").Optional(),
		field.String("external_ip").Optional(),
		field.String("environment").Optional(),
		field.String("cloud_project").Optional(),
		field.String("cloud_zone").Optional(),
		field.String("cloud_machine_type").Optional(),
	}
}

func (InventoryMachine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bronze_links", InventoryMachineBronzeLink.Type),
	}
}

func (InventoryMachine) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("os_type"),
		index.Fields("environment"),
		index.Fields("collected_at"),
	}
}

func (InventoryMachine) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "machines"},
	}
}
