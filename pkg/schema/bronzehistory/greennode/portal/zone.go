package portal

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodePortalZone stores historical snapshots of GreenNode zones.
type BronzeHistoryGreenNodePortalZone struct {
	ent.Schema
}

func (BronzeHistoryGreenNodePortalZone) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodePortalZone) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("openstack_zone").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodePortalZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodePortalZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_portal_zones_history"},
	}
}
