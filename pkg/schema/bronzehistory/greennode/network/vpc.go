package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeNetworkVpc stores historical snapshots of GreenNode VPCs.
type BronzeHistoryGreenNodeNetworkVpc struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkVpc) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeNetworkVpc) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
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
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeNetworkVpc) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeNetworkVpc) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_vpcs_history"},
	}
}
