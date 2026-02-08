package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeTargetHttpProxy stores historical snapshots of GCP Compute target HTTP proxies.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeTargetHttpProxy struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeTargetHttpProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeTargetHttpProxy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze target HTTP proxy by resource_id"),

		// All target HTTP proxy fields
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
		field.Bool("proxy_bind").
			Default(false),
		field.Int32("http_keep_alive_timeout_sec").
			Optional(),
		field.String("region").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeTargetHttpProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeTargetHttpProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_http_proxies_history"},
	}
}
