package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigURIAttackPattern stores URI patterns used to detect injection and
// traversal attacks in access log URIs. Patterns are derived from OWASP
// ModSecurity CRS and tagged with the source rule ID for traceability.
type ConfigURIAttackPattern struct {
	ent.Schema
}

func (ConfigURIAttackPattern) Fields() []ent.Field {
	return []ent.Field{
		field.String("pattern_type").NotEmpty().
			Comment("Attack category: lfi, sqli, rce, xss, ssrf"),
		field.String("match_mode").NotEmpty().
			Comment("How to match: substring (ILIKE), regex (PostgreSQL ~*)"),
		field.String("pattern").NotEmpty().
			Comment("The match string — substring or PostgreSQL-compatible regex"),
		field.String("match_target").Optional().
			Comment("Future: uri, query — empty means full URI"),
		field.String("crs_rule_id").Optional().
			Comment("OWASP CRS rule ID for traceability, e.g. '930100'"),
		field.String("description").Optional().
			Comment("Human-readable note about what this pattern detects"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigURIAttackPattern) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pattern_type", "pattern").Unique(),
		index.Fields("is_active"),
		index.Fields("pattern_type"),
	}
}

func (ConfigURIAttackPattern) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "uri_attack_patterns"},
	}
}
