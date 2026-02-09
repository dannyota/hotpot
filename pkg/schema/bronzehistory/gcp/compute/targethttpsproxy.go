package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeTargetHttpsProxy stores historical snapshots of GCP Compute target HTTPS proxies.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeTargetHttpsProxy struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeTargetHttpsProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeTargetHttpsProxy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze target HTTPS proxy by resource_id"),

		// All target HTTPS proxy fields
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
			Optional(),
		field.String("quic_override").
			Optional(),
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
			Optional(),
		field.JSON("ssl_certificates_json", []interface{}{}).
			Optional(),
		field.String("region").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeTargetHttpsProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeTargetHttpsProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_https_proxies_history"},
	}
}
