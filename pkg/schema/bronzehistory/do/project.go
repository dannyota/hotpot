package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOProject stores historical snapshots of DigitalOcean Projects.
type BronzeHistoryDOProject struct {
	ent.Schema
}

func (BronzeHistoryDOProject) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOProject) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Project by resource_id"),
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
			Optional(),
		field.String("api_updated_at").
			Optional(),
	}
}

func (BronzeHistoryDOProject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("environment"),
		index.Fields("is_default"),
	}
}

func (BronzeHistoryDOProject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_projects_history"},
	}
}
