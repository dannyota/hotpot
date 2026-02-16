package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOVolume represents a DigitalOcean Block Storage Volume in the bronze layer.
type BronzeDOVolume struct {
	ent.Schema
}

func (BronzeDOVolume) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOVolume) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Volume UUID"),
		field.String("name").
			NotEmpty(),
		field.String("region").
			Optional().
			Comment("Region slug"),
		field.Int64("size_gigabytes").
			Default(0),
		field.String("description").
			Optional(),
		field.JSON("droplet_ids_json", []int{}).
			Optional().
			Comment("Attached droplet IDs"),
		field.String("filesystem_type").
			Optional(),
		field.String("filesystem_label").
			Optional(),
		field.JSON("tags_json", []string{}).
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDOVolume) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOVolume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_volumes"},
	}
}
