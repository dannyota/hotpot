package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigRpmCoreRepo stores RPM repository names considered OS core for software
// lifecycle detection. Packages from these repos are classified as OS components
// rather than independently tracked software. Manageable via admin UI.
type ConfigRpmCoreRepo struct {
	ent.Schema
}

func (ConfigRpmCoreRepo) Fields() []ent.Field {
	return []ent.Field{
		field.String("repo_name").NotEmpty().Unique().
			Comment("RPM repo name, e.g. rhel9-baseos, epel9"),
		field.String("description").Optional().
			Comment("Human-readable note about this repo"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigRpmCoreRepo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}

func (ConfigRpmCoreRepo) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "rpm_core_repos"},
	}
}
