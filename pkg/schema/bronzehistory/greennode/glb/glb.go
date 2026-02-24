package glb

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeGLBGlobalLoadBalancer stores historical snapshots of GreenNode global load balancers.
type BronzeHistoryGreenNodeGLBGlobalLoadBalancer struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeGLBGlobalLoadBalancer) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeGLBGlobalLoadBalancer) Fields() []ent.Field {
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
		field.String("package").
			Optional(),
		field.String("type").
			Optional(),
		field.Int("user_id").
			Optional(),
		field.JSON("vips_json", json.RawMessage{}).
			Optional(),
		field.JSON("domains_json", json.RawMessage{}).
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.String("deleted_at_api").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalLoadBalancer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalLoadBalancer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_load_balancers_history"},
	}
}

// BronzeHistoryGreenNodeGLBGlobalListener stores historical snapshots of global load balancer listeners.
type BronzeHistoryGreenNodeGLBGlobalListener struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeGLBGlobalListener) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("glb_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeGLBGlobalLoadBalancer"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("listener_id").
			NotEmpty(),
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

func (BronzeHistoryGreenNodeGLBGlobalListener) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("glb_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalListener) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_listeners_history"},
	}
}

// BronzeHistoryGreenNodeGLBGlobalPool stores historical snapshots of global load balancer pools.
type BronzeHistoryGreenNodeGLBGlobalPool struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeGLBGlobalPool) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("glb_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeGLBGlobalLoadBalancer"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("pool_id").
			NotEmpty(),
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
			Optional(),
		field.JSON("pool_members_json", json.RawMessage{}).
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

func (BronzeHistoryGreenNodeGLBGlobalPool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("glb_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_pools_history"},
	}
}
