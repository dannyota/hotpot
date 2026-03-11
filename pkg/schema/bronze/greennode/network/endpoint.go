package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeNetworkEndpoint represents a GreenNode network endpoint in the bronze layer.
type BronzeGreenNodeNetworkEndpoint struct {
	ent.Schema
}

func (BronzeGreenNodeNetworkEndpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeNetworkEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Endpoint ID"),
		field.String("name").
			NotEmpty(),
		field.String("ipv4_address").
			Optional(),
		field.String("endpoint_url").
			Optional(),
		field.String("endpoint_auth_url").
			Optional(),
		field.String("endpoint_service_id").
			Optional(),
		field.String("status").
			Optional(),
		field.String("billing_status").
			Optional(),
		field.String("endpoint_type").
			Optional(),
		field.String("version").
			Optional(),
		field.String("description").
			Optional(),
		field.String("created_at").
			Optional(),
		field.String("updated_at").
			Optional(),
		field.String("vpc_id").
			Optional(),
		field.String("vpc_name").
			Optional(),
		field.String("zone_uuid").
			Optional(),
		field.Bool("enable_dns_name").
			Default(false),
		field.JSON("endpoint_domains", []string{}).
			Optional(),
		field.String("subnet_id").
			Optional(),
		field.String("category_name").
			Optional(),
		field.String("service_name").
			Optional(),
		field.String("service_endpoint_type").
			Optional(),
		field.String("package_name").
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeNetworkEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeNetworkEndpoint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_endpoints"},
	}
}
