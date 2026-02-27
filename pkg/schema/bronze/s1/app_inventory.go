package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1AppInventory represents a SentinelOne application inventory entry in the bronze layer.
type BronzeS1AppInventory struct {
	ent.Schema
}

func (BronzeS1AppInventory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1AppInventory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Synthesized: name||vendor"),
		field.String("application_name").
			NotEmpty(),
		field.String("application_vendor").
			Optional(),
		field.Int("endpoints_count").
			Optional(),
		field.Int("application_versions_count").
			Optional(),
		field.Bool("estimate").
			Default(false),
	}
}

func (BronzeS1AppInventory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_name"),
		index.Fields("application_vendor"),
		index.Fields("collected_at"),
	}
}

func (BronzeS1AppInventory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_app_inventory"},
	}
}
