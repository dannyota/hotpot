package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeTargetTcpProxy stores historical snapshots of GCP Compute target TCP proxies.
type BronzeHistoryGCPComputeTargetTcpProxy struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeTargetTcpProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeTargetTcpProxy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze target TCP proxy by resource_id"),

		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("service").
			Optional(),
		field.Bool("proxy_bind").
			Default(false),
		field.String("proxy_header").
			Optional(),
		field.String("region").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeTargetTcpProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeTargetTcpProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_tcp_proxies_history"},
	}
}
