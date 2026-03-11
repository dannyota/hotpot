package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceEOLIdentifier represents a product identifier from the endoflife.date database.
type BronzeReferenceEOLIdentifier struct {
	ent.Schema
}

func (BronzeReferenceEOLIdentifier) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceEOLIdentifier) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {product}:{type}:{value}"),
		field.String("product").
			Comment("Product slug (e.g. nginx, mysql)"),
		field.String("identifier_type").
			Comment("Identifier type: purl, repology, cpe"),
		field.String("value").
			Comment("Full identifier string (e.g. pkg:deb/ubuntu/nginx)"),
	}
}

func (BronzeReferenceEOLIdentifier) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product"),
		index.Fields("identifier_type"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceEOLIdentifier) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_eol_identifiers"},
	}
}
