package machine

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SilverMachineBronzeLink tracks which bronze records contributed to a machine.
type SilverMachineBronzeLink struct {
	ent.Schema
}

func (SilverMachineBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (SilverMachineBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("machine", SilverMachine.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (SilverMachineBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "machine_bronze_links"},
	}
}
