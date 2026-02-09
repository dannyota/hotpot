package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1Account stores historical snapshots of SentinelOne accounts.
type BronzeHistoryS1Account struct {
	ent.Schema
}

func (BronzeHistoryS1Account) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1Account) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze account by resource_id"),

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

func (BronzeHistoryS1Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryS1Account) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_accounts_history"},
	}
}
