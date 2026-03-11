package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigOsCoreRule stores OS core classification patterns for software lifecycle
// detection. Packages matching these rules have their lifecycle tied to the OS
// release, not tracked independently. Manageable via admin UI.
//
// Rule types:
//   - prefix: package name starts with this value
//   - suffix: package name ends with this value
//   - exact:  package name matches exactly
type ConfigOsCoreRule struct {
	ent.Schema
}

func (ConfigOsCoreRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("rule_type").NotEmpty().
			Comment("Rule type: prefix, suffix, exact"),
		field.String("os_type").Default("").
			Comment("OS filter: linux, windows, macos, or empty for all"),
		field.String("value").NotEmpty().
			Comment("The pattern (e.g. linux-, -keyring, gmail)"),
		field.String("description").Optional().
			Comment("Human-readable explanation of why this is OS core"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigOsCoreRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("rule_type", "os_type", "value").Unique(),
		index.Fields("rule_type", "is_active"),
	}
}

func (ConfigOsCoreRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "os_core_rules"},
	}
}
