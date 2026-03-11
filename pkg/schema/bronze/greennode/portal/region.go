package portal

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodePortalRegion represents a GreenNode region in the bronze layer.
type BronzeGreenNodePortalRegion struct {
	ent.Schema
}

func (BronzeGreenNodePortalRegion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodePortalRegion) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GreenNode region ID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodePortalRegion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodePortalRegion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_portal_regions"},
	}
}
