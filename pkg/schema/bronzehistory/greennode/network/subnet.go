package network

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeNetworkSubnet stores historical snapshots of GreenNode network subnets.
type BronzeHistoryGreenNodeNetworkSubnet struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkSubnet) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeNetworkSubnet) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("network_id").
			NotEmpty(),
		field.String("cidr").
			Optional(),
		field.String("status").
			Optional(),
		field.String("route_table_id").
			Optional(),
		field.String("interface_acl_policy_id").
			Optional(),
		field.String("interface_acl_policy_name").
			Optional(),
		field.String("zone_id").
			Optional(),
		field.JSON("secondary_subnets", json.RawMessage{}).
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeNetworkSubnet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeNetworkSubnet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_subnets_history"},
	}
}
