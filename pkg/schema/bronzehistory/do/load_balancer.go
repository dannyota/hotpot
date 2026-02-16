package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOLoadBalancer stores historical snapshots of DigitalOcean Load Balancers.
type BronzeHistoryDOLoadBalancer struct {
	ent.Schema
}

func (BronzeHistoryDOLoadBalancer) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOLoadBalancer) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze LoadBalancer by resource_id"),
		field.String("name").
			Optional(),
		field.String("ip").
			Optional(),
		field.String("ipv6").
			Optional(),
		field.String("size_slug").
			Optional(),
		field.Uint32("size_unit").
			Default(0),
		field.String("lb_type").
			Optional(),
		field.String("algorithm").
			Optional(),
		field.String("status").
			Optional(),
		field.String("region").
			Optional(),
		field.String("tag").
			Optional(),
		field.Bool("redirect_http_to_https").
			Default(false),
		field.Bool("enable_proxy_protocol").
			Default(false),
		field.Bool("enable_backend_keepalive").
			Default(false),
		field.String("vpc_uuid").
			Optional(),
		field.String("project_id").
			Optional(),
		field.Uint64("http_idle_timeout_seconds").
			Optional().
			Nillable(),
		field.Bool("disable_lets_encrypt_dns_records").
			Optional().
			Nillable(),
		field.String("network").
			Optional(),
		field.String("network_stack").
			Optional(),
		field.String("tls_cipher_policy").
			Optional(),
		field.String("api_created_at").
			Optional(),
		field.JSON("forwarding_rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("health_check_json", json.RawMessage{}).
			Optional(),
		field.JSON("sticky_sessions_json", json.RawMessage{}).
			Optional(),
		field.JSON("firewall_json", json.RawMessage{}).
			Optional(),
		field.JSON("domains_json", json.RawMessage{}).
			Optional(),
		field.JSON("glb_settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("droplet_ids_json", json.RawMessage{}).
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("target_load_balancer_ids_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryDOLoadBalancer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("region"),
		index.Fields("status"),
		index.Fields("vpc_uuid"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryDOLoadBalancer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_load_balancers_history"},
	}
}
