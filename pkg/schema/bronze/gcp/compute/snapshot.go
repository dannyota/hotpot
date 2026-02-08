package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeSnapshot represents a GCP Compute Engine snapshot in the bronze layer.
// Fields preserve raw API response data from compute.snapshots.list.
type BronzeGCPComputeSnapshot struct {
	ent.Schema
}

func (BronzeGCPComputeSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
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

		// SnapshotEncryptionKeyJSON contains snapshot encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("snapshot_encryption_key_json", json.RawMessage{}).
			Optional(),

		// SourceDiskEncryptionKeyJSON contains source disk encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("source_disk_encryption_key_json", json.RawMessage{}).
			Optional(),

		// GuestOsFeaturesJSON contains guest OS features enabled on this snapshot.
		//
		//	[{"type": "VIRTIO_SCSI_MULTIQUEUE"}, {"type": "UEFI_COMPATIBLE"}]
		field.JSON("guest_os_features_json", json.RawMessage{}).
			Optional(),

		// StorageLocationsJSON contains regions where snapshot data is stored.
		//
		//	["us-central1", "us"]
		field.JSON("storage_locations_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeSnapshotLabel.Type),
		edge.To("licenses", BronzeGCPComputeSnapshotLicense.Type),
	}
}

func (BronzeGCPComputeSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_snapshots"},
	}
}

// BronzeGCPComputeSnapshotLabel represents a label attached to a GCP Compute snapshot.
type BronzeGCPComputeSnapshotLabel struct {
	ent.Schema
}

func (BronzeGCPComputeSnapshotLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeSnapshotLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("snapshot", BronzeGCPComputeSnapshot.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeSnapshotLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_snapshot_labels"},
	}
}

// BronzeGCPComputeSnapshotLicense represents a license attached to a GCP Compute snapshot.
type BronzeGCPComputeSnapshotLicense struct {
	ent.Schema
}

func (BronzeGCPComputeSnapshotLicense) Fields() []ent.Field {
	return []ent.Field{
		field.String("license").
			NotEmpty(),
	}
}

func (BronzeGCPComputeSnapshotLicense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("snapshot", BronzeGCPComputeSnapshot.Type).
			Ref("licenses").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeSnapshotLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_snapshot_licenses"},
	}
}
