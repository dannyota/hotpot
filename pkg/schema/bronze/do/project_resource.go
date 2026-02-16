package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOProjectResource represents a DigitalOcean Project Resource in the bronze layer.
type BronzeDOProjectResource struct {
	ent.Schema
}

func (BronzeDOProjectResource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOProjectResource) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {projectID}:{URN}"),
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

func (BronzeDOProjectResource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("urn"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOProjectResource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_project_resources"},
	}
}
