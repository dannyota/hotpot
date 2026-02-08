package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeTargetHttpProxy represents a GCP Compute Engine target HTTP proxy in the bronze layer.
// Fields preserve raw API response data from compute.targetHttpProxies.aggregatedList.
type BronzeGCPComputeTargetHttpProxy struct {
	ent.Schema
}

func (BronzeGCPComputeTargetHttpProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeTargetHttpProxy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("fingerprint").
			Optional(),
		field.String("url_map").
			Optional().
			Comment("URL to the UrlMap resource that defines the mapping from URL to the BackendService"),
		field.Bool("proxy_bind").
			Default(false),
		field.Int32("http_keep_alive_timeout_sec").
			Optional().
			Comment("Specifies how long to keep a connection open for HTTP keepalive"),
		field.String("region").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeTargetHttpProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeTargetHttpProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_http_proxies"},
	}
}
