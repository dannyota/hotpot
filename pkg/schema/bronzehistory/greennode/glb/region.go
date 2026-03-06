package glb

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeGLBGlobalRegion stores historical snapshots of GreenNode global LB regions.
type BronzeHistoryGreenNodeGLBGlobalRegion struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeGLBGlobalRegion) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeGLBGlobalRegion) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("vserver_endpoint").
			Optional(),
		field.String("vlb_endpoint").
			Optional(),
		field.String("ui_server_endpoint").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalRegion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalRegion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_regions_history"},
	}
}
