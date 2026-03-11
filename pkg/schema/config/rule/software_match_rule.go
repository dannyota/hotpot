package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigSoftwareMatchRule stores product matching overrides for software lifecycle
// detection. Controls how installed software maps to endoflife.date products.
// Manageable via admin UI.
//
// Rule types:
//   - extra_prefix: additional package name prefixes beyond PURL/repology
//   - exclude:      package name substrings to skip for a product
//   - name_cycle_map: maps display names to cycle versions (e.g. "2022" → "16.0")
type ConfigSoftwareMatchRule struct {
	ent.Schema
}

func (ConfigSoftwareMatchRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("product_slug").NotEmpty().
			Comment("endoflife.date product slug (e.g. docker-engine, firefox)"),
		field.String("rule_type").NotEmpty().
			Comment("Rule type: extra_prefix, exclude, name_cycle_map"),
		field.String("os_type").Default("").
			Comment("OS filter: linux, windows, macos, or empty for all"),
		field.String("value").NotEmpty().
			Comment("The prefix, exclude word, or name-cycle key"),
		field.String("extra_value").Optional().
			Comment("For name_cycle_map: the cycle value (e.g. 12.0)"),
		field.String("description").Optional().
			Comment("Human-readable note about this rule"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigSoftwareMatchRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_slug", "rule_type", "os_type", "value").Unique(),
		index.Fields("product_slug", "is_active"),
		index.Fields("rule_type", "is_active"),
	}
}

func (ConfigSoftwareMatchRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "software_match_rules"},
	}
}
