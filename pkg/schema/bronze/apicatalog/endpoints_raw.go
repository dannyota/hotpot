package apicatalog

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeApicatalogEndpointsRaw holds raw API endpoint data imported from CSV.
type BronzeApicatalogEndpointsRaw struct {
	ent.Schema
}

func (BronzeApicatalogEndpointsRaw) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeApicatalogEndpointsRaw) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("UUID primary key"),
		field.String("log_source_id").
			Optional().
			Comment("Which traffic source these belong to"),
		field.String("name").
			Optional().
			Comment("Route name from CSV"),
		field.String("service_name").
			Optional().
			Comment("Service label, e.g. \"DBS jwt-auth\""),
		field.String("upstream").
			Optional().
			Comment("Upstream code, e.g. \"dbs\""),
		field.String("uri").
			NotEmpty().
			Comment("API path, e.g. \"/protected/app/api/revi-card/payment\""),
		field.String("method").
			Default("").
			Comment("HTTP method(s), raw from CSV (may be \"POST,PUT\")"),
		field.String("route_status").
			Default("").
			Comment("Active/Inactive as raw string"),
		field.String("plugin_auth").
			Optional().
			Comment("Auth plugin name, e.g. \"jwt-auth\""),
		field.String("plugin_auth_enable").
			Optional().
			Comment("Auth plugin enabled as raw string, e.g. \"true\""),
		field.String("source_file").
			Optional().
			Comment("CSV filename this was imported from"),
	}
}

func (BronzeApicatalogEndpointsRaw) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "upstream", "uri", "method", "route_status").Unique(),
		index.Fields("upstream"),
		index.Fields("log_source_id"),
	}
}

func (BronzeApicatalogEndpointsRaw) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "apicatalog_endpoints_raw"},
	}
}
