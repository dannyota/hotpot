package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeNetworkSecgroup stores historical snapshots of GreenNode security groups.
type BronzeHistoryGreenNodeNetworkSecgroup struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkSecgroup) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeNetworkSecgroup) Fields() []ent.Field {
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
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeNetworkSecgroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeNetworkSecgroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_secgroups_history"},
	}
}

// BronzeHistoryGreenNodeNetworkSecgroupRule stores historical snapshots of security group rules.
type BronzeHistoryGreenNodeNetworkSecgroupRule struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkSecgroupRule) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("secgroup_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeNetworkSecgroup"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("rule_id").
			NotEmpty(),
		field.String("direction").
			Optional(),
		field.String("ether_type").
			Optional(),
		field.String("protocol").
			Optional(),
		field.String("description").
			Optional(),
		field.String("remote_ip_prefix").
			Optional(),
		field.Int("port_range_max").
			Optional(),
		field.Int("port_range_min").
			Optional(),
	}
}

func (BronzeHistoryGreenNodeNetworkSecgroupRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("secgroup_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeNetworkSecgroupRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_secgroup_rules_history"},
	}
}
