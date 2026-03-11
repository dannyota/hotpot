package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigAuthEndpointPattern stores URI patterns used to identify authentication,
// admin, and sensitive endpoints. Users can add custom patterns via admin UI.
// Seed data provides keyword-based defaults that work across common frameworks.
type ConfigAuthEndpointPattern struct {
	ent.Schema
}

func (ConfigAuthEndpointPattern) Fields() []ent.Field {
	return []ent.Field{
		field.String("pattern_type").NotEmpty().
			Comment("Endpoint category: login, otp, password_reset, register, admin, token"),
		field.String("match_mode").NotEmpty().
			Comment("How to match: keyword (ILIKE), substring (ILIKE), regex (PostgreSQL ~*)"),
		field.String("pattern").NotEmpty().
			Comment("The match string — keyword, path substring, or PostgreSQL-compatible regex"),
		field.String("source").NotEmpty().Default("custom").
			Comment("Origin: seed (shipped default) or custom (user-added)"),
		field.String("description").Optional().
			Comment("Human-readable note about this pattern"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigAuthEndpointPattern) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pattern_type", "pattern").Unique(),
		index.Fields("is_active"),
		index.Fields("pattern_type"),
	}
}

func (ConfigAuthEndpointPattern) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "auth_endpoint_patterns"},
	}
}
