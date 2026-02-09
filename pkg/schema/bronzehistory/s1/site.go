package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1Site stores historical snapshots of SentinelOne sites.
type BronzeHistoryS1Site struct {
	ent.Schema
}

func (BronzeHistoryS1Site) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1Site) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze site by resource_id"),

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

func (BronzeHistoryS1Site) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("account_id"),
	}
}

func (BronzeHistoryS1Site) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_sites_history"},
	}
}
