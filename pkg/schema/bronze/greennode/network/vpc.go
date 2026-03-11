package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeNetworkVpc represents a GreenNode VPC in the bronze layer.
type BronzeGreenNodeNetworkVpc struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkVpc) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkVpc) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("VPC ID"),
		field.String("name").
			NotEmpty(),
		field.String("cidr").
			Optional(),
		field.String("status").
			Optional(),
		field.String("route_table_id").
			Optional(),
		field.String("route_table_name").
			Optional(),
		field.String("dhcp_option_id").
			Optional(),
		field.String("dhcp_option_name").
			Optional(),
		field.String("dns_status").
			Optional(),
		field.String("dns_id").
			Optional(),
		field.String("zone_uuid").
			Optional(),
		field.String("zone_name").
			Optional(),
		field.String("created_at").
			Optional(),
		field.JSON("elastic_ips", []string{}).
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkVpc) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeNetworkVpc) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_vpcs"},
	}
}
