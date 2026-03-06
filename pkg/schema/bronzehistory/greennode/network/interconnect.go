package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeNetworkInterconnect stores historical snapshots of GreenNode network interconnects.
type BronzeHistoryGreenNodeNetworkInterconnect struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkInterconnect) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeNetworkInterconnect) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("status").
			Optional(),
		field.Bool("enable_gw2").
			Default(false),
		field.Int("circuit_id").
			Optional(),
		field.String("gw01_ip").
			Optional(),
		field.String("gw02_ip").
			Optional(),
		field.String("gw_vip").
			Optional(),
		field.String("remote_gw01_ip").
			Optional(),
		field.String("remote_gw02_ip").
			Optional(),
		field.String("package_id").
			Optional(),
		field.String("type_id").
			Optional(),
		field.String("type_name").
			Optional(),
		field.String("created_at").
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeNetworkInterconnect) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeNetworkInterconnect) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_interconnects_history"},
	}
}
