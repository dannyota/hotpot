package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeDisk stores historical snapshots of GCP Compute persistent disks.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeDisk struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeDisk) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeDisk) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze disk by resource_id"),

		// All disk fields (same as bronze.BronzeGCPComputeDisk)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("zone").
			Optional(),
		field.String("region").
			Optional(),
		field.String("type").
			Optional(),
		field.String("status").
			Optional(),
		field.Int64("size_gb").
			Optional(),
		field.String("architecture").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Attachment timestamps
		field.String("last_attach_timestamp").
			Optional(),
		field.String("last_detach_timestamp").
			Optional(),

		// Source fields
		field.String("source_image").
			Optional(),
		field.String("source_image_id").
			Optional(),
		field.String("source_snapshot").
			Optional(),
		field.String("source_snapshot_id").
			Optional(),
		field.String("source_disk").
			Optional(),
		field.String("source_disk_id").
			Optional(),

		// Performance settings
		field.Int64("provisioned_iops").
			Optional(),
		field.Int64("provisioned_throughput").
			Optional(),
		field.Int64("physical_block_size_bytes").
			Optional(),

		// Security
		field.Bool("enable_confidential_compute").
			Default(false),

		// JSONB fields
		field.JSON("disk_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("users_json", json.RawMessage{}).
			Optional(),
		field.JSON("replica_zones_json", json.RawMessage{}).
			Optional(),
		field.JSON("resource_policies_json", json.RawMessage{}).
			Optional(),
		field.JSON("guest_os_features_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeDisk) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeDisk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_disks_history"},
	}
}

// BronzeHistoryGCPComputeDiskLabel stores historical snapshots of disk labels.
// Links via disk_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeDiskLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeDiskLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("disk_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeDisk"),
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

func (BronzeHistoryGCPComputeDiskLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("disk_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeDiskLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_disk_labels_history"},
	}
}

// BronzeHistoryGCPComputeDiskLicense stores historical snapshots of disk licenses.
// Links via disk_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeDiskLicense struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeDiskLicense) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("disk_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeDisk"),
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

func (BronzeHistoryGCPComputeDiskLicense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("disk_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeDiskLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_disk_licenses_history"},
	}
}
