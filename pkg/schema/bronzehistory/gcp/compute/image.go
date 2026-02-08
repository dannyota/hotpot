package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPComputeImage stores historical snapshots of GCP Compute Engine images.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeImage struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeImage) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze image by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All image fields (same as bronze.BronzeGCPComputeImage)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("status").
			Optional(),
		field.String("architecture").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),
		field.String("family").
			Optional(),

		// Source fields
		field.String("source_disk").
			Optional(),
		field.String("source_disk_id").
			Optional(),
		field.String("source_image").
			Optional(),
		field.String("source_image_id").
			Optional(),
		field.String("source_snapshot").
			Optional(),
		field.String("source_snapshot_id").
			Optional(),
		field.String("source_type").
			Optional(),

		// Size fields
		field.Int64("disk_size_gb").
			Optional(),
		field.Int64("archive_size_bytes").
			Optional(),

		// Flags
		field.Bool("satisfies_pzi").
			Default(false),
		field.Bool("satisfies_pzs").
			Default(false),
		field.Bool("enable_confidential_compute").
			Default(false),

		// JSONB fields
		field.JSON("image_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("source_disk_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("source_image_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("source_snapshot_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("deprecated_json", json.RawMessage{}).
			Optional(),
		field.JSON("guest_os_features_json", json.RawMessage{}).
			Optional(),
		field.JSON("shielded_instance_initial_state_json", json.RawMessage{}).
			Optional(),
		field.JSON("raw_disk_json", json.RawMessage{}).
			Optional(),
		field.JSON("storage_locations_json", json.RawMessage{}).
			Optional(),
		field.JSON("license_codes_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeImage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_images_history"},
	}
}

// BronzeHistoryGCPComputeImageLabel stores historical snapshots of image labels.
// Links via image_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeImageLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeImageLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("image_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeImage"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPComputeImageLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("image_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeImageLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_image_labels_history"},
	}
}

// BronzeHistoryGCPComputeImageLicense stores historical snapshots of image licenses.
// Links via image_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeImageLicense struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeImageLicense) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("image_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeImage"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// License field
		field.String("license").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeImageLicense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("image_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeImageLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_image_licenses_history"},
	}
}
