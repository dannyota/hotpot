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

// BronzeHistoryGCPComputeInterconnect stores historical snapshots of GCP Compute interconnects.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeInterconnect struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInterconnect) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeInterconnect) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze interconnect by resource_id"),

		// All interconnect fields (same as bronze.BronzeGCPComputeInterconnect)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("location").
			Optional(),
		field.String("interconnect_type").
			Optional(),
		field.String("link_type").
			Optional(),
		field.Bool("admin_enabled").
			Default(false),
		field.String("operational_status").
			Optional(),
		field.Int("provisioned_link_count").
			Optional(),
		field.Int("requested_link_count").
			Optional(),
		field.String("peer_ip_address").
			Optional(),
		field.String("google_ip_address").
			Optional(),
		field.String("google_reference_id").
			Optional(),
		field.String("noc_contact_email").
			Optional(),
		field.String("customer_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// JSONB fields
		field.JSON("expected_outages_json", json.RawMessage{}).
			Optional(),
		field.JSON("circuit_infos_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeInterconnect) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeInterconnect) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_interconnects_history"},
	}
}
