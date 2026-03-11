package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeNetworkInterconnect represents a GreenNode network interconnect in the bronze layer.
type BronzeGreenNodeNetworkInterconnect struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkInterconnect) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkInterconnect) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Interconnect ID"),
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
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkInterconnect) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeNetworkInterconnect) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_interconnects"},
	}
}
