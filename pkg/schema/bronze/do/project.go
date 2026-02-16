package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOProject represents a DigitalOcean Project in the bronze layer.
type BronzeDOProject struct {
	ent.Schema
}

func (BronzeDOProject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOProject) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Project UUID"),
		field.String("owner_uuid").
			Optional(),
		field.Uint64("owner_id").
			Default(0),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("purpose").
			Optional(),
		field.String("environment").
			Optional(),
		field.Bool("is_default").
			Default(false),
		field.String("api_created_at").
			Optional().
			Comment("API-reported creation timestamp"),
		field.String("api_updated_at").
			Optional().
			Comment("API-reported update timestamp"),
	}
}

func (BronzeDOProject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("environment"),
		index.Fields("is_default"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOProject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_projects"},
	}
}
