package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigSanctionedCountry stores country codes that should trigger critical
// alerts when traffic originates from them (OFAC/sanctioned countries).
type ConfigSanctionedCountry struct {
	ent.Schema
}

func (ConfigSanctionedCountry) Fields() []ent.Field {
	return []ent.Field{
		field.String("country_code").NotEmpty().Unique().
			Comment("ISO 3166-1 alpha-2 country code, e.g. 'KP', 'IR'"),
		field.String("description").Optional().
			Comment("Human-readable note, e.g. 'North Korea (OFAC)'"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigSanctionedCountry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}

func (ConfigSanctionedCountry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sanctioned_countries"},
	}
}
