package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOProjectResource stores historical snapshots of DigitalOcean Project Resources.
type BronzeHistoryDOProjectResource struct {
	ent.Schema
}

func (BronzeHistoryDOProjectResource) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOProjectResource) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze ProjectResource by resource_id"),
		field.String("project_id").
			NotEmpty(),
		field.String("urn").
			NotEmpty(),
		field.String("assigned_at").
			Optional(),
		field.String("status").
			Optional(),
	}
}

func (BronzeHistoryDOProjectResource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
		index.Fields("urn"),
	}
}

func (BronzeHistoryDOProjectResource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_project_resources_history"},
	}
}
