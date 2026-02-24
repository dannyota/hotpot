package loadbalancer

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeLoadBalancerLB represents a GreenNode load balancer in the bronze layer.
type BronzeGreenNodeLoadBalancerLB struct {
	ent.Schema
}

func (BronzeGreenNodeLoadBalancerLB) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeLoadBalancerLB) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Load balancer UUID"),
		field.String("name").
			NotEmpty(),
		field.String("display_status").
			Optional(),
		field.String("address").
			Optional(),
		field.String("private_subnet_id").
			Optional(),
		field.String("private_subnet_cidr").
			Optional(),
		field.String("type").
			Optional().
			Comment("Load balancer type"),
		field.String("display_type").
			Optional(),
		field.String("load_balancer_schema").
			Optional(),
		field.String("package_id").
			Optional(),
		field.String("description").
			Optional(),
		field.String("location").
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.String("progress_status").
			Optional(),
		field.String("status").
			Optional(),
		field.String("backend_subnet_id").
			Optional(),
		field.Bool("internal").
			Default(false),
		field.Bool("auto_scalable").
			Default(false),
		field.String("zone_id").
			Optional(),
		field.Int("min_size").
			Optional(),
		field.Int("max_size").
			Optional(),
		field.Int("total_nodes").
			Optional(),
		field.JSON("nodes_json", json.RawMessage{}).
			Optional().
			Comment("Nodes as JSONB"),

		// Collection metadata
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeLoadBalancerLB) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("listeners", BronzeGreenNodeLoadBalancerListener.Type),
		edge.To("pools", BronzeGreenNodeLoadBalancerPool.Type),
	}
}

func (BronzeGreenNodeLoadBalancerLB) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeLoadBalancerLB) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_lbs"},
	}
}

// BronzeGreenNodeLoadBalancerListener represents a listener attached to a load balancer.
type BronzeGreenNodeLoadBalancerListener struct {
	ent.Schema
}

func (BronzeGreenNodeLoadBalancerListener) Fields() []ent.Field {
	return []ent.Field{
		field.String("listener_id").
			NotEmpty().
			Comment("SDK listener UUID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("protocol").
			Optional(),
		field.Int("protocol_port").
			Optional(),
		field.Int("connection_limit").
			Optional(),
		field.String("default_pool_id").
			Optional(),
		field.String("default_pool_name").
			Optional(),
		field.Int("timeout_client").
			Optional(),
		field.Int("timeout_member").
			Optional(),
		field.Int("timeout_connection").
			Optional(),
		field.String("allowed_cidrs").
			Optional(),
		field.JSON("certificate_authorities_json", json.RawMessage{}).
			Optional(),
		field.String("display_status").
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.String("default_certificate_authority").
			Optional().
			Nillable(),
		field.String("client_certificate_authentication").
			Optional().
			Nillable(),
		field.String("progress_status").
			Optional(),
		field.JSON("insert_headers_json", json.RawMessage{}).
			Optional(),
		field.JSON("policies_json", json.RawMessage{}).
			Optional().
			Comment("Fetched policies as JSONB"),
	}
}

func (BronzeGreenNodeLoadBalancerListener) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("lb", BronzeGreenNodeLoadBalancerLB.Type).
			Ref("listeners").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeLoadBalancerListener) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_listeners"},
	}
}

// BronzeGreenNodeLoadBalancerPool represents a pool attached to a load balancer.
type BronzeGreenNodeLoadBalancerPool struct {
	ent.Schema
}

func (BronzeGreenNodeLoadBalancerPool) Fields() []ent.Field {
	return []ent.Field{
		field.String("pool_id").
			NotEmpty().
			Comment("SDK pool UUID"),
		field.String("name").
			NotEmpty(),
		field.String("protocol").
			Optional(),
		field.String("description").
			Optional(),
		field.String("load_balance_method").
			Optional(),
		field.String("status").
			Optional(),
		field.Bool("stickiness").
			Default(false),
		field.Bool("tls_encryption").
			Default(false),
		field.JSON("members_json", json.RawMessage{}).
			Optional().
			Comment("Members as JSONB"),
		field.JSON("health_monitor_json", json.RawMessage{}).
			Optional().
			Comment("Health monitor as JSONB"),
	}
}

func (BronzeGreenNodeLoadBalancerPool) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("lb", BronzeGreenNodeLoadBalancerLB.Type).
			Ref("pools").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeLoadBalancerPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_pools"},
	}
}
