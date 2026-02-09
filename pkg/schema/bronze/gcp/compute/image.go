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

// BronzeGCPComputeImage represents a GCP Compute Engine image in the bronze layer.
// Fields preserve raw API response data from compute.images.list.
type BronzeGCPComputeImage struct {
	ent.Schema
}

func (BronzeGCPComputeImage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeImage) Fields() []ent.Field {
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

		// ImageEncryptionKeyJSON contains image encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("image_encryption_key_json", json.RawMessage{}).
			Optional(),

		// SourceDiskEncryptionKeyJSON contains source disk encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("source_disk_encryption_key_json", json.RawMessage{}).
			Optional(),

		// SourceImageEncryptionKeyJSON contains source image encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("source_image_encryption_key_json", json.RawMessage{}).
			Optional(),

		// SourceSnapshotEncryptionKeyJSON contains source snapshot encryption configuration.
		//
		//	{
		//	  "sha256": "...",
		//	  "kmsKeyName": "projects/.../cryptoKeys/..."
		//	}
		field.JSON("source_snapshot_encryption_key_json", json.RawMessage{}).
			Optional(),

		// DeprecatedJSON contains deprecation state information.
		//
		//	{"state": "DEPRECATED", "replacement": "..."}
		field.JSON("deprecated_json", json.RawMessage{}).
			Optional(),

		// GuestOsFeaturesJSON contains guest OS features enabled on this image.
		//
		//	[{"type": "UEFI_COMPATIBLE"}]
		field.JSON("guest_os_features_json", json.RawMessage{}).
			Optional(),

		// ShieldedInstanceInitialStateJSON contains shielded instance initial state.
		field.JSON("shielded_instance_initial_state_json", json.RawMessage{}).
			Optional(),

		// RawDiskJSON contains raw disk source information.
		//
		//	{"source": "gs://...", "containerType": "TAR"}
		field.JSON("raw_disk_json", json.RawMessage{}).
			Optional(),

		// StorageLocationsJSON contains regions where image data is stored.
		//
		//	["us-central1"]
		field.JSON("storage_locations_json", json.RawMessage{}).
			Optional(),

		// LicenseCodesJSON contains license codes for the image.
		//
		//	[1234567890]
		field.JSON("license_codes_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeImage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPComputeImageLabel.Type),
		edge.To("licenses", BronzeGCPComputeImageLicense.Type),
	}
}

func (BronzeGCPComputeImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeImage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_images"},
	}
}

// BronzeGCPComputeImageLabel represents a label attached to a GCP Compute image.
type BronzeGCPComputeImageLabel struct {
	ent.Schema
}

func (BronzeGCPComputeImageLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPComputeImageLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("image", BronzeGCPComputeImage.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeImageLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_image_labels"},
	}
}

// BronzeGCPComputeImageLicense represents a license attached to a GCP Compute image.
type BronzeGCPComputeImageLicense struct {
	ent.Schema
}

func (BronzeGCPComputeImageLicense) Fields() []ent.Field {
	return []ent.Field{
		field.String("license").
			NotEmpty(),
	}
}

func (BronzeGCPComputeImageLicense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("image", BronzeGCPComputeImage.Type).
			Ref("licenses").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeImageLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_image_licenses"},
	}
}
