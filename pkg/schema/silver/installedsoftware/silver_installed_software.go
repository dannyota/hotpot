package installedsoftware

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	silvermixin "github.com/dannyota/hotpot/pkg/schema/silver/mixin"
)

// SilverInstalledSoftware is the final merged installed software table.
type SilverInstalledSoftware struct {
	ent.Schema
}

func (SilverInstalledSoftware) Mixin() []ent.Mixin {
	return []ent.Mixin{
		silvermixin.Timestamp{},
	}
}

func (SilverInstalledSoftware) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("machine_id").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("version").Optional(),
		field.String("publisher").Optional(),
	}
}

func (SilverInstalledSoftware) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bronze_links", SilverInstalledSoftwareBronzeLink.Type),
	}
}

func (SilverInstalledSoftware) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("machine_id"),
		index.Fields("name"),
		index.Fields("machine_id", "name").Unique(),
	}
}

func (SilverInstalledSoftware) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "installed_software"},
	}
}
