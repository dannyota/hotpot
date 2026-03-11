package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceEOLProduct represents a product from the endoflife.date database.
type BronzeReferenceEOLProduct struct {
	ent.Schema
}

func (BronzeReferenceEOLProduct) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceEOLProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Product slug (e.g. rhel, ubuntu, windows-server)"),
		field.String("name").
			Comment("Display name (e.g. Red Hat Enterprise Linux)"),
		field.String("category").
			Comment("Product category (os, db, framework, lang, library, server-app, service, standard, device, app)"),
		field.Strings("tags").
			Optional().
			Comment("Tags from YAML frontmatter (e.g. erlang-runtime, linux-foundation)"),
	}
}

func (BronzeReferenceEOLProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceEOLProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_eol_products"},
	}
}
