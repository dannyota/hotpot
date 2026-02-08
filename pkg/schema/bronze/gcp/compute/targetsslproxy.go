package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeTargetSslProxy represents a GCP Compute Engine target SSL proxy in the bronze layer.
type BronzeGCPComputeTargetSslProxy struct {
	ent.Schema
}

func (BronzeGCPComputeTargetSslProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.Timestamp{}}
}

func (BronzeGCPComputeTargetSslProxy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("resource_id").Unique().Immutable().Comment("GCP API ID"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("service").Optional().Comment("URL to the BackendService resource"),
		field.String("proxy_header").Optional(),
		field.String("certificate_map").Optional().Comment("URL of a certificate map"),
		field.String("ssl_policy").Optional().Comment("URL of SslPolicy resource"),
		field.JSON("ssl_certificates_json", []interface{}{}).Optional().Comment("URLs to SslCertificate resources"),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeGCPComputeTargetSslProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeTargetSslProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_ssl_proxies"},
	}
}
