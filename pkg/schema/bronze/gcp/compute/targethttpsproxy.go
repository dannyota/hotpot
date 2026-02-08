package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeTargetHttpsProxy represents a GCP Compute Engine target HTTPS proxy in the bronze layer.
// Fields preserve raw API response data from compute.targetHttpsProxies.aggregatedList.
type BronzeGCPComputeTargetHttpsProxy struct {
	ent.Schema
}

func (BronzeGCPComputeTargetHttpsProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeTargetHttpsProxy) Fields() []ent.Field {
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
		field.String("quic_override").
			Optional().
			Comment("QUIC negotiation override"),
		field.String("server_tls_policy").
			Optional(),
		field.String("authorization_policy").
			Optional(),
		field.String("certificate_map").
			Optional(),
		field.String("ssl_policy").
			Optional(),
		field.String("tls_early_data").
			Optional(),
		field.Bool("proxy_bind").
			Default(false),
		field.Int32("http_keep_alive_timeout_sec").
			Optional().
			Comment("Specifies how long to keep a connection open for HTTP keepalive"),
		field.JSON("ssl_certificates_json", []interface{}{}).
			Optional().
			Comment("URLs to SslCertificate resources"),
		field.String("region").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeTargetHttpsProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeTargetHttpsProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_https_proxies"},
	}
}
