package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeImage represents a GCP Compute Engine image in the bronze layer.
// Fields preserve raw API response data from compute.images.list.
type GCPComputeImage struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50);index" json:"status"`
	Architecture      string `gorm:"column:architecture;type:varchar(50)" json:"architecture"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`
	Family            string `gorm:"column:family;type:varchar(255)" json:"family"`

	// Source fields
	SourceDisk       string `gorm:"column:source_disk;type:text" json:"sourceDisk"`
	SourceDiskId     string `gorm:"column:source_disk_id;type:varchar(255)" json:"sourceDiskId"`
	SourceImage      string `gorm:"column:source_image;type:text" json:"sourceImage"`
	SourceImageId    string `gorm:"column:source_image_id;type:varchar(255)" json:"sourceImageId"`
	SourceSnapshot   string `gorm:"column:source_snapshot;type:text" json:"sourceSnapshot"`
	SourceSnapshotId string `gorm:"column:source_snapshot_id;type:varchar(255)" json:"sourceSnapshotId"`
	SourceType       string `gorm:"column:source_type;type:varchar(50)" json:"sourceType"`

	// Size fields
	DiskSizeGb       int64 `gorm:"column:disk_size_gb" json:"diskSizeGb"`
	ArchiveSizeBytes int64 `gorm:"column:archive_size_bytes" json:"archiveSizeBytes"`

	// Flags
	SatisfiesPzi              bool `gorm:"column:satisfies_pzi" json:"satisfiesPzi"`
	SatisfiesPzs              bool `gorm:"column:satisfies_pzs" json:"satisfiesPzs"`
	EnableConfidentialCompute bool `gorm:"column:enable_confidential_compute" json:"enableConfidentialCompute"`

	// ImageEncryptionKeyJSON contains image encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	ImageEncryptionKeyJSON jsonb.JSON `gorm:"column:image_encryption_key_json;type:jsonb" json:"imageEncryptionKey"`

	// SourceDiskEncryptionKeyJSON contains source disk encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	SourceDiskEncryptionKeyJSON jsonb.JSON `gorm:"column:source_disk_encryption_key_json;type:jsonb" json:"sourceDiskEncryptionKey"`

	// SourceImageEncryptionKeyJSON contains source image encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	SourceImageEncryptionKeyJSON jsonb.JSON `gorm:"column:source_image_encryption_key_json;type:jsonb" json:"sourceImageEncryptionKey"`

	// SourceSnapshotEncryptionKeyJSON contains source snapshot encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	SourceSnapshotEncryptionKeyJSON jsonb.JSON `gorm:"column:source_snapshot_encryption_key_json;type:jsonb" json:"sourceSnapshotEncryptionKey"`

	// DeprecatedJSON contains deprecation state information.
	//
	//	{"state": "DEPRECATED", "replacement": "..."}
	DeprecatedJSON jsonb.JSON `gorm:"column:deprecated_json;type:jsonb" json:"deprecated"`

	// GuestOsFeaturesJSON contains guest OS features enabled on this image.
	//
	//	[{"type": "UEFI_COMPATIBLE"}]
	GuestOsFeaturesJSON jsonb.JSON `gorm:"column:guest_os_features_json;type:jsonb" json:"guestOsFeatures"`

	// ShieldedInstanceInitialStateJSON contains shielded instance initial state.
	ShieldedInstanceInitialStateJSON jsonb.JSON `gorm:"column:shielded_instance_initial_state_json;type:jsonb" json:"shieldedInstanceInitialState"`

	// RawDiskJSON contains raw disk source information.
	//
	//	{"source": "gs://...", "containerType": "TAR"}
	RawDiskJSON jsonb.JSON `gorm:"column:raw_disk_json;type:jsonb" json:"rawDisk"`

	// StorageLocationsJSON contains regions where image data is stored.
	//
	//	["us-central1"]
	StorageLocationsJSON jsonb.JSON `gorm:"column:storage_locations_json;type:jsonb" json:"storageLocations"`

	// LicenseCodesJSON contains license codes for the image.
	//
	//	[1234567890]
	LicenseCodesJSON jsonb.JSON `gorm:"column:license_codes_json;type:jsonb" json:"licenseCodes"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Labels   []GCPComputeImageLabel   `gorm:"foreignKey:ImageResourceID;references:ResourceID" json:"labels,omitempty"`
	Licenses []GCPComputeImageLicense `gorm:"foreignKey:ImageResourceID;references:ResourceID" json:"licenses,omitempty"`
}

func (GCPComputeImage) TableName() string {
	return "bronze.gcp_compute_images"
}

// GCPComputeImageLabel represents a label attached to a GCP Compute image.
type GCPComputeImageLabel struct {
	ID              uint   `gorm:"primaryKey"`
	ImageResourceID string `gorm:"column:image_resource_id;type:varchar(255);not null;index" json:"-"`
	Key             string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value           string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeImageLabel) TableName() string {
	return "bronze.gcp_compute_image_labels"
}

// GCPComputeImageLicense represents a license attached to a GCP Compute image.
type GCPComputeImageLicense struct {
	ID              uint   `gorm:"primaryKey"`
	ImageResourceID string `gorm:"column:image_resource_id;type:varchar(255);not null;index" json:"-"`
	License         string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeImageLicense) TableName() string {
	return "bronze.gcp_compute_image_licenses"
}
