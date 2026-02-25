package network

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeNetworkEndpoint stores historical snapshots of GreenNode network endpoints.
type BronzeHistoryGreenNodeNetworkEndpoint struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeNetworkEndpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeNetworkEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
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
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeNetworkEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeNetworkEndpoint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_network_endpoints_history"},
	}
}
