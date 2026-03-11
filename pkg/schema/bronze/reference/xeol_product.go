package reference

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeReferenceXeolProduct represents a product from the xeol EOL database.
type BronzeReferenceXeolProduct struct {
	ent.Schema
}

func (BronzeReferenceXeolProduct) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeReferenceXeolProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite key: {ecosystem}:{name} (e.g. maven:com.h2database:h2)"),
		field.String("name").
			Comment("Product name (e.g. com.h2database:h2)"),
		field.String("permalink").
			Optional().
			Comment("External URL (e.g. https://central.sonatype.com/artifact/com.h2database/h2)"),
	}
}

func (BronzeReferenceXeolProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("collected_at"),
	}
}

func (BronzeReferenceXeolProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reference_xeol_products"},
	}
}
