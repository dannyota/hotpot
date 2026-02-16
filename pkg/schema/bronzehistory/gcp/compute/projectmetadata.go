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

// BronzeHistoryGCPComputeProjectMetadata stores historical snapshots of GCP Compute project metadata.
// Uses resource_id for lookup (project ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeProjectMetadata struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeProjectMetadata) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeProjectMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze project metadata by resource_id"),

		// All project metadata fields (same as bronze.BronzeGCPComputeProjectMetadata)
		field.String("name").
			NotEmpty(),
		field.String("default_service_account").
			Optional(),
		field.String("default_network_tier").
			Optional(),
		field.String("xpn_project_status").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// JSONB fields
		field.JSON("usage_export_location_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeProjectMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeProjectMetadata) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_project_metadata_history"},
	}
}

// BronzeHistoryGCPComputeProjectMetadataItem stores historical snapshots of project metadata items.
// Links via metadata_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeProjectMetadataItem struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeProjectMetadataItem) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("metadata_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeProjectMetadata"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Metadata item fields
		field.String("key").
			NotEmpty(),
		field.Text("value").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeProjectMetadataItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("metadata_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeProjectMetadataItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_project_metadata_items_history"},
	}
}
