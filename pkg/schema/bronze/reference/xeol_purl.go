package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceXeolPurl represents a package URL from the xeol EOL database.
type BronzeReferenceXeolPurl struct {
	ent.Schema
}

func (BronzeReferenceXeolPurl) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceXeolPurl) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {product_id}:{purl}"),
		field.String("product_id").
			Comment("Reference to xeol product ID"),
		field.String("purl").
			Comment("Package URL (e.g. pkg:maven/com.h2database/h2)"),
	}
}

func (BronzeReferenceXeolPurl) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
		index.Fields("purl"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceXeolPurl) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_xeol_purls"},
	}
}
