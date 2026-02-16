package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODomain represents a DigitalOcean Domain in the bronze layer.
type BronzeDODomain struct {
	ent.Schema
}

func (BronzeDODomain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODomain) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Domain name (used as ID)"),
		field.Int("ttl").
			Default(0),
		field.String("zone_file").
			Optional(),
	}
}

func (BronzeDODomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
	}
}

func (BronzeDODomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_domains"},
	}
}
