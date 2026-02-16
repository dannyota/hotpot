package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOLoadBalancer represents a DigitalOcean Load Balancer in the bronze layer.
type BronzeDOLoadBalancer struct {
	ent.Schema
}

func (BronzeDOLoadBalancer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOLoadBalancer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Load Balancer UUID"),
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
			Optional().
			Comment("Load balancer type"),
		field.String("algorithm").
			Optional(),
		field.String("status").
			Optional(),
		field.String("region").
			Optional().
			Comment("Region slug"),
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
			Optional().
			Comment("API-reported creation timestamp"),
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

func (BronzeDOLoadBalancer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region"),
		index.Fields("status"),
		index.Fields("vpc_uuid"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOLoadBalancer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_load_balancers"},
	}
}
