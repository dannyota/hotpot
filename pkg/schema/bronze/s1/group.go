package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1Group represents a SentinelOne group in the bronze layer.
type BronzeS1Group struct {
	ent.Schema
}

func (BronzeS1Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne group ID"),
		field.String("name").
			NotEmpty(),
		field.String("site_id").
			Optional(),
		field.String("type").
			Optional(),
		field.Bool("is_default").
			Default(false),
		field.Bool("inherits").
			Default(false),
		field.Int("rank").
			Optional().
			Nillable(),
		field.Int("total_agents").
			Default(0),
		field.String("creator").
			Optional(),
		field.String("creator_id").
			Optional(),
		field.String("filter_name").
			Optional(),
		field.String("filter_id").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.Time("api_updated_at").
			Optional().
			Nillable(),
		field.String("registration_token").
			Optional(),
	}
}

func (BronzeS1Group) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("site_id"),
		index.Fields("type"),
		index.Fields("collected_at"),
	}
}

func (BronzeS1Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_groups"},
	}
}
