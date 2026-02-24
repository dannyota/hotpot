package loadbalancer

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeLoadBalancerPackage stores historical snapshots of GreenNode LB packages.
type BronzeHistoryGreenNodeLoadBalancerPackage struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeLoadBalancerPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeLoadBalancerPackage) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("type").
			Optional(),
		field.Int("connection_number").
			Optional(),
		field.Int("data_transfer").
			Optional(),
		field.String("mode").
			Optional(),
		field.String("lb_type").
			Optional(),
		field.String("display_lb_type").
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_packages_history"},
	}
}
