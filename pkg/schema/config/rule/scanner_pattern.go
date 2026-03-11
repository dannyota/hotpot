package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigScannerPattern stores keywords used to detect security scanner tools
// from HTTP user agent strings. Manageable via admin UI.
type ConfigScannerPattern struct {
	ent.Schema
}

func (ConfigScannerPattern) Fields() []ent.Field {
	return []ent.Field{
		field.String("keyword").NotEmpty().Unique().
			Comment("Substring to match in UA string, e.g. 'sqlmap', 'nikto'"),
		field.String("description").Optional().
			Comment("Human-readable note about this scanner tool"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigScannerPattern) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}

func (ConfigScannerPattern) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "scanner_patterns"},
	}
}
