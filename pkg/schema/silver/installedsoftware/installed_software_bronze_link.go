package installedsoftware

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SilverInstalledSoftwareBronzeLink tracks which bronze records contributed to a software record.
type SilverInstalledSoftwareBronzeLink struct {
	ent.Schema
}

func (SilverInstalledSoftwareBronzeLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").NotEmpty(),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),
	}
}

func (SilverInstalledSoftwareBronzeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("installed_software", SilverInstalledSoftware.Type).Ref("bronze_links").Unique().Required(),
	}
}

func (SilverInstalledSoftwareBronzeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "installed_software_bronze_links"},
	}
}
