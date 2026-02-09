package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1Group stores historical snapshots of SentinelOne groups.
type BronzeHistoryS1Group struct {
	ent.Schema
}

func (BronzeHistoryS1Group) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1Group) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze group by resource_id"),

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
	}
}

func (BronzeHistoryS1Group) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("site_id"),
	}
}

func (BronzeHistoryS1Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_groups_history"},
	}
}
