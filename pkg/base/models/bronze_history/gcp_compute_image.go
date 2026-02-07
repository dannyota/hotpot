package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeImage stores historical snapshots of GCP Compute Engine images.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeImage struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Image has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All image fields (same as bronze.GCPComputeImage)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50)" json:"status"`
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

	// JSONB fields
	ImageEncryptionKeyJSON           jsonb.JSON `gorm:"column:image_encryption_key_json;type:jsonb" json:"imageEncryptionKey"`
	SourceDiskEncryptionKeyJSON      jsonb.JSON `gorm:"column:source_disk_encryption_key_json;type:jsonb" json:"sourceDiskEncryptionKey"`
	SourceImageEncryptionKeyJSON     jsonb.JSON `gorm:"column:source_image_encryption_key_json;type:jsonb" json:"sourceImageEncryptionKey"`
	SourceSnapshotEncryptionKeyJSON  jsonb.JSON `gorm:"column:source_snapshot_encryption_key_json;type:jsonb" json:"sourceSnapshotEncryptionKey"`
	DeprecatedJSON                   jsonb.JSON `gorm:"column:deprecated_json;type:jsonb" json:"deprecated"`
	GuestOsFeaturesJSON              jsonb.JSON `gorm:"column:guest_os_features_json;type:jsonb" json:"guestOsFeatures"`
	ShieldedInstanceInitialStateJSON jsonb.JSON `gorm:"column:shielded_instance_initial_state_json;type:jsonb" json:"shieldedInstanceInitialState"`
	RawDiskJSON                      jsonb.JSON `gorm:"column:raw_disk_json;type:jsonb" json:"rawDisk"`
	StorageLocationsJSON             jsonb.JSON `gorm:"column:storage_locations_json;type:jsonb" json:"storageLocations"`
	LicenseCodesJSON                 jsonb.JSON `gorm:"column:license_codes_json;type:jsonb" json:"licenseCodes"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeImage) TableName() string {
	return "bronze_history.gcp_compute_images"
}

// GCPComputeImageLabel stores historical snapshots of image labels.
// Links via ImageHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeImageLabel struct {
	HistoryID      uint `gorm:"primaryKey"`
	ImageHistoryID uint `gorm:"column:image_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeImageLabel) TableName() string {
	return "bronze_history.gcp_compute_image_labels"
}

// GCPComputeImageLicense stores historical snapshots of image licenses.
// Links via ImageHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeImageLicense struct {
	HistoryID      uint `gorm:"primaryKey"`
	ImageHistoryID uint `gorm:"column:image_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// License field
	License string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeImageLicense) TableName() string {
	return "bronze_history.gcp_compute_image_licenses"
}
