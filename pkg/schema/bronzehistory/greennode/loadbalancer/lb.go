package loadbalancer

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeLoadBalancerLB stores historical snapshots of GreenNode load balancers.
type BronzeHistoryGreenNodeLoadBalancerLB struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeLoadBalancerLB) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeLoadBalancerLB) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
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
			Optional(),
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
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerLB) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerLB) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_lbs_history"},
	}
}

// BronzeHistoryGreenNodeLoadBalancerListener stores historical snapshots of load balancer listeners.
type BronzeHistoryGreenNodeLoadBalancerListener struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeLoadBalancerListener) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("lb_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeLoadBalancerLB"),
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
			Optional(),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerListener) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("lb_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerListener) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_listeners_history"},
	}
}

// BronzeHistoryGreenNodeLoadBalancerPool stores historical snapshots of load balancer pools.
type BronzeHistoryGreenNodeLoadBalancerPool struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeLoadBalancerPool) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("lb_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeLoadBalancerLB"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("pool_id").
			NotEmpty(),
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
			Optional(),
		field.JSON("health_monitor_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerPool) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("lb_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_pools_history"},
	}
}
