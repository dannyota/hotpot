package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
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
		field.String("state").
			Optional(),
		field.String("account_type").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.Time("api_updated_at").
			Optional().
			Nillable(),
		field.Time("expiration").
			Optional().
			Nillable(),
		field.Bool("unlimited_expiration").
			Default(false),
		field.Int("active_agents").
			Default(0),
		field.Int("total_licenses").
			Default(0),
		field.String("usage_type").
			Optional(),
		field.String("billing_mode").
			Optional(),
		field.String("creator").
			Optional(),
		field.String("creator_id").
			Optional(),
		field.Int("number_of_sites").
			Default(0),
		field.String("external_id").
			Optional(),
		field.JSON("licenses_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeS1Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("state"),
		index.Fields("account_type"),
	}
}

func (BronzeS1Account) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_accounts"},
	}
}
