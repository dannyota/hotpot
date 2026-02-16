package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeRouter stores historical snapshots of GCP Compute routers.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeRouter struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeRouter) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeRouter) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze router by resource_id"),

		// All router fields (same as bronze.BronzeGCPComputeRouter)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Router configuration
		field.String("network").
			Optional(),
		field.String("region").
			Optional(),

		// BGP configuration
		field.Int("bgp_asn").
			Default(0),
		field.String("bgp_advertise_mode").
			Optional(),
		field.JSON("bgp_advertised_groups_json", json.RawMessage{}).
			Optional(),
		field.JSON("bgp_advertised_ip_ranges_json", json.RawMessage{}).
			Optional(),
		field.Int("bgp_keepalive_interval").
			Default(0),

		// JSONB fields
		field.JSON("bgp_peers_json", json.RawMessage{}).
			Optional(),
		field.JSON("interfaces_json", json.RawMessage{}).
			Optional(),
		field.JSON("nats_json", json.RawMessage{}).
			Optional(),

		field.Bool("encrypted_interconnect_router").
			Default(false),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeRouter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeRouter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_routers_history"},
	}
}
