package glb

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeGLBGlobalRegion represents a GreenNode global load balancer region in the bronze layer.
type BronzeGreenNodeGLBGlobalRegion struct {
	ent.Schema
}

func (BronzeGreenNodeGLBGlobalRegion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeGLBGlobalRegion) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Global region ID"),
		field.String("name").
			NotEmpty(),
		field.String("vserver_endpoint").
			Optional(),
		field.String("vlb_endpoint").
			Optional(),
		field.String("ui_server_endpoint").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeGLBGlobalRegion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeGLBGlobalRegion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_regions"},
	}
}
