package portal

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodePortalZone represents a GreenNode availability zone in the bronze layer.
type BronzeGreenNodePortalZone struct {
	ent.Schema
}

func (BronzeGreenNodePortalZone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodePortalZone) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GreenNode zone UUID"),
		field.String("name").
			NotEmpty(),
		field.String("openstack_zone").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodePortalZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodePortalZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_portal_zones"},
	}
}
