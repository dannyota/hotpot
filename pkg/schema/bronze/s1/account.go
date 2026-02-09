package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1Account represents a SentinelOne account in the bronze layer.
type BronzeS1Account struct {
	ent.Schema
}

func (BronzeS1Account) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1Account) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne account ID"),
		field.String("name").
			NotEmpty(),
	}
}

func (BronzeS1Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
	}
}

func (BronzeS1Account) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_accounts"},
	}
}
