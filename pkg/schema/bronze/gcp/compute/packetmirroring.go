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

// BronzeGCPComputePacketMirroring represents a GCP Compute Engine packet mirroring in the bronze layer.
// Fields preserve raw API response data from compute.packetMirrorings.list.
type BronzeGCPComputePacketMirroring struct {
	ent.Schema
}

func (BronzeGCPComputePacketMirroring) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputePacketMirroring) Fields() []ent.Field {
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

func (BronzeGCPComputePacketMirroring) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputePacketMirroring) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_packet_mirrorings"},
	}
}
