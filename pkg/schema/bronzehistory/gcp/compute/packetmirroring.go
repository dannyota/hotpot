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

// BronzeHistoryGCPComputePacketMirroring stores historical snapshots of GCP Compute packet mirrorings.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputePacketMirroring struct {
	ent.Schema
}

func (BronzeHistoryGCPComputePacketMirroring) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputePacketMirroring) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze packet mirroring by resource_id"),

		// All packet mirroring fields (same as bronze.BronzeGCPComputePacketMirroring)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("region").
			Optional(),
		field.String("network").
			Optional(),
		field.Int("priority").
			Optional(),
		field.String("enable").
			Optional(),

		// JSONB fields
		field.JSON("collector_ilb_json", json.RawMessage{}).
			Optional(),
		field.JSON("mirrored_resources_json", json.RawMessage{}).
			Optional(),
		field.JSON("filter_json", json.RawMessage{}).
			Optional(),

		field.String("creation_timestamp").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputePacketMirroring) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputePacketMirroring) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_packet_mirrorings_history"},
	}
}
