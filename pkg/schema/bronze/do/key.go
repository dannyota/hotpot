package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOKey represents a DigitalOcean SSH Key in the bronze layer.
type BronzeDOKey struct {
	ent.Schema
}

func (BronzeDOKey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean SSH Key ID (int converted to string)"),
		field.String("name").
			NotEmpty(),
		field.String("fingerprint").
			Optional(),
		field.String("public_key").
			Optional(),
	}
}

func (BronzeDOKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("fingerprint"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_keys"},
	}
}
