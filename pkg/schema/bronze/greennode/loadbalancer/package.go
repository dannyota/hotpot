package loadbalancer

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeLoadBalancerPackage represents a GreenNode load balancer package in the bronze layer.
type BronzeGreenNodeLoadBalancerPackage struct {
	ent.Schema
}

func (BronzeGreenNodeLoadBalancerPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeLoadBalancerPackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Load balancer package UUID"),
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

		// Collection metadata
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeLoadBalancerPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeLoadBalancerPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_packages"},
	}
}
