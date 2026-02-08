package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeTargetSslProxy stores historical snapshots of GCP Compute target SSL proxies.
type BronzeHistoryGCPComputeTargetSslProxy struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeTargetSslProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeTargetSslProxy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").Unique().Immutable(),
		field.String("resource_id").NotEmpty().Comment("Link to bronze target SSL proxy by resource_id"),
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.String("creation_timestamp").Optional(),
		field.String("self_link").Optional(),
		field.String("service").Optional(),
		field.String("proxy_header").Optional(),
		field.String("certificate_map").Optional(),
		field.String("ssl_policy").Optional(),
		field.JSON("ssl_certificates_json", []interface{}{}).Optional(),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeTargetSslProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeTargetSslProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_ssl_proxies_history"},
	}
}
