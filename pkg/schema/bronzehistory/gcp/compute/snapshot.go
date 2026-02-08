package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPComputeSnapshot stores historical snapshots of GCP Compute Engine snapshots.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeSnapshot struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze snapshot by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All snapshot fields (same as bronze.BronzeGCPComputeSnapshot)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("status").
			Optional(),
		field.Int64("disk_size_gb").
			Optional(),
		field.Int64("storage_bytes").
			Optional(),
		field.String("storage_bytes_status").
			Optional(),
		field.Int64("download_bytes").
			Optional(),
		field.String("snapshot_type").
			Optional(),
		field.String("architecture").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),

		// Source fields
		field.String("source_disk").
			Optional(),
		field.String("source_disk_id").
			Optional(),

		// Recovery
		field.String("source_disk_for_recovery_checkpoint").
			Optional(),

		// Flags
		field.Bool("auto_created").
			Default(false),
		field.Bool("satisfies_pzi").
			Default(false),
		field.Bool("satisfies_pzs").
			Default(false),
		field.Bool("enable_confidential_compute").
			Default(false),

		// JSONB fields
		field.JSON("snapshot_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("source_disk_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("guest_os_features_json", json.RawMessage{}).
			Optional(),
		field.JSON("storage_locations_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_snapshots_history"},
	}
}

// BronzeHistoryGCPComputeSnapshotLabel stores historical snapshots of snapshot labels.
// Links via snapshot_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeSnapshotLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeSnapshotLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("snapshot_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeSnapshot"),
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

func (BronzeHistoryGCPComputeSnapshotLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("snapshot_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeSnapshotLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_snapshot_labels_history"},
	}
}

// BronzeHistoryGCPComputeSnapshotLicense stores historical snapshots of snapshot licenses.
// Links via snapshot_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeSnapshotLicense struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeSnapshotLicense) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("snapshot_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeSnapshot"),
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

func (BronzeHistoryGCPComputeSnapshotLicense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("snapshot_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeSnapshotLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_snapshot_licenses_history"},
	}
}
