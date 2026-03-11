package rule

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ConfigLibraryUa stores UA family names considered automated/library clients.
// Traffic from these families on protected endpoints triggers anomaly detection.
// Manageable via admin UI.
type ConfigLibraryUa struct {
	ent.Schema
}

func (ConfigLibraryUa) Fields() []ent.Field {
	return []ent.Field{
		field.String("family").NotEmpty().Unique().
			Comment("Normalized UA family name, e.g. 'curl', 'python-requests'"),
		field.String("description").Optional().
			Comment("Human-readable note about this UA family"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}

func (ConfigLibraryUa) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}

func (ConfigLibraryUa) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "library_uas"},
	}
}
