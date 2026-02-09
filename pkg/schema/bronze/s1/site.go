package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1Site represents a SentinelOne site in the bronze layer.
type BronzeS1Site struct {
	ent.Schema
}

func (BronzeS1Site) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1Site) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne site ID"),
		field.String("name").
			NotEmpty(),
		field.String("account_id").
			Optional(),
		field.String("account_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("site_type").
			Optional(),
		field.String("suite").
			Optional(),
		field.String("creator").
			Optional(),
		field.String("creator_id").
			Optional(),
		field.Bool("health_status").
			Default(false),
		field.Int("active_licenses").
			Default(0),
		field.Int("total_licenses").
			Default(0),
		field.Bool("unlimited_licenses").
			Default(false),
		field.Bool("is_default").
			Default(false),
		field.String("description").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.Time("expiration").
			Optional().
			Nillable(),
	}
}

func (BronzeS1Site) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("account_id"),
		index.Fields("state"),
		index.Fields("collected_at"),
	}
}

func (BronzeS1Site) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_sites"},
	}
}
