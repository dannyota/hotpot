package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeInterconnect represents a GCP Compute Engine interconnect in the bronze layer.
// Fields preserve raw API response data from compute.interconnects.list.
type BronzeGCPComputeInterconnect struct {
	ent.Schema
}

func (BronzeGCPComputeInterconnect) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeInterconnect) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable(),
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

func (BronzeGCPComputeInterconnect) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeInterconnect) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_interconnects"},
	}
}
