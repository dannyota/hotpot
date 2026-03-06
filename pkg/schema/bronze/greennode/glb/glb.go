package glb

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeGLBGlobalLoadBalancer represents a GreenNode global load balancer in the bronze layer.
type BronzeGreenNodeGLBGlobalLoadBalancer struct {
	ent.Schema
}

func (BronzeGreenNodeGLBGlobalLoadBalancer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeGLBGlobalLoadBalancer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Global load balancer ID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("status").
			Optional(),
		field.String("package").
			Optional(),
		field.String("type").
			Optional(),
		field.Int("user_id").
			Optional(),
		field.JSON("vips_json", json.RawMessage{}).
			Optional().
			Comment("VIPs as JSONB"),
		field.JSON("domains_json", json.RawMessage{}).
			Optional().
			Comment("Domains as JSONB"),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.String("deleted_at_api").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeGLBGlobalLoadBalancer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("listeners", BronzeGreenNodeGLBGlobalListener.Type),
		edge.To("pools", BronzeGreenNodeGLBGlobalPool.Type),
	}
}

func (BronzeGreenNodeGLBGlobalLoadBalancer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeGLBGlobalLoadBalancer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_load_balancers"},
	}
}

// BronzeGreenNodeGLBGlobalListener represents a listener attached to a global load balancer.
type BronzeGreenNodeGLBGlobalListener struct {
	ent.Schema
}

func (BronzeGreenNodeGLBGlobalListener) Fields() []ent.Field {
	return []ent.Field{
		field.String("listener_id").
			NotEmpty().
			Comment("SDK listener ID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("protocol").
			Optional(),
		field.Int("port").
			Optional(),
		field.String("global_pool_id").
			Optional(),
		field.Int("timeout_client").
			Optional(),
		field.Int("timeout_member").
			Optional(),
		field.Int("timeout_connection").
			Optional(),
		field.String("allowed_cidrs").
			Optional(),
		field.String("headers").
			Optional().
			Nillable(),
		field.String("status").
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.String("deleted_at_api").
			Optional().
			Nillable(),
	}
}

func (BronzeGreenNodeGLBGlobalListener) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("glb", BronzeGreenNodeGLBGlobalLoadBalancer.Type).
			Ref("listeners").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeGLBGlobalListener) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_listeners"},
	}
}

// BronzeGreenNodeGLBGlobalPool represents a pool attached to a global load balancer.
type BronzeGreenNodeGLBGlobalPool struct {
	ent.Schema
}

func (BronzeGreenNodeGLBGlobalPool) Fields() []ent.Field {
	return []ent.Field{
		field.String("pool_id").
			NotEmpty().
			Comment("SDK pool ID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("algorithm").
			Optional(),
		field.String("sticky_session").
			Optional().
			Nillable(),
		field.String("tls_enabled").
			Optional().
			Nillable(),
		field.String("protocol").
			Optional(),
		field.String("status").
			Optional(),
		field.JSON("health_json", json.RawMessage{}).
			Optional().
			Comment("Health monitor as JSONB"),
		field.JSON("pool_members_json", json.RawMessage{}).
			Optional().
			Comment("Fetched pool members as JSONB"),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.String("deleted_at_api").
			Optional().
			Nillable(),
	}
}

func (BronzeGreenNodeGLBGlobalPool) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("glb", BronzeGreenNodeGLBGlobalLoadBalancer.Type).
			Ref("pools").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeGLBGlobalPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_pools"},
	}
}
