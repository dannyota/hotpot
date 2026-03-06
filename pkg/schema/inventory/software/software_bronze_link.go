package software

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// InventorySoftwareBronzeLink tracks which bronze records contributed to a software record.
type InventorySoftwareBronzeLink struct {
	ent.Schema
}

func (InventorySoftwareBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (InventorySoftwareBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("software", InventorySoftware.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (InventorySoftwareBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "software_bronze_links"},
	}
}
