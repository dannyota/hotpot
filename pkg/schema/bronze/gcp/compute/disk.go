package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeDisk represents a GCP Compute Engine persistent disk in the bronze layer.
// Fields preserve raw API response data from compute.disks.list.
type BronzeGCPComputeDisk struct {
	ent.Schema
}

func (BronzeGCPComputeDisk) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeDisk) Fields() []ent.Field {
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

		// DiskEncryptionKeyJSON contains disk encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("disk_encryption_key_json", json.RawMessage{}).
			Optional(),

		// UsersJSON contains list of instance URLs using this disk.
		//
		//	["projects/.../instances/vm1", "projects/.../instances/vm2"]
		field.JSON("users_json", json.RawMessage{}).
			Optional(),

		// ReplicaZonesJSON contains zones for regional disk replication.
		//
		//	["zones/us-central1-a", "zones/us-central1-b"]
		field.JSON("replica_zones_json", json.RawMessage{}).
			Optional(),

		// ResourcePoliciesJSON contains attached resource policies for snapshots.
		//
		//	["projects/.../resourcePolicies/policy1"]
		field.JSON("resource_policies_json", json.RawMessage{}).
			Optional(),

		// GuestOsFeaturesJSON contains guest OS features enabled on disk.
		//
		//	[{"type": "VIRTIO_SCSI_MULTIQUEUE"}, {"type": "UEFI_COMPATIBLE"}]
		field.JSON("guest_os_features_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeDisk) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeDiskLabel.Type),
		edge.To("licenses", BronzeGCPComputeDiskLicense.Type),
	}
}

func (BronzeGCPComputeDisk) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeDisk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_disks"},
	}
}

// BronzeGCPComputeDiskLabel represents a label attached to a GCP Compute disk.
type BronzeGCPComputeDiskLabel struct {
	ent.Schema
}

func (BronzeGCPComputeDiskLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeDiskLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("disk", BronzeGCPComputeDisk.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeDiskLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_disk_labels"},
	}
}

// BronzeGCPComputeDiskLicense represents a license attached to a GCP Compute disk.
type BronzeGCPComputeDiskLicense struct {
	ent.Schema
}

func (BronzeGCPComputeDiskLicense) Fields() []ent.Field {
	return []ent.Field{
		field.String("license").
			NotEmpty(),
	}
}

func (BronzeGCPComputeDiskLicense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("disk", BronzeGCPComputeDisk.Type).
			Ref("licenses").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeDiskLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_disk_licenses"},
	}
}
