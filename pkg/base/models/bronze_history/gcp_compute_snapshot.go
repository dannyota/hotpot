package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeSnapshot stores historical snapshots of GCP Compute Engine snapshots.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeSnapshot struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Snapshot has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All snapshot fields (same as bronze.GCPComputeSnapshot)
	Name               string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description        string `gorm:"column:description;type:text" json:"description"`
	Status             string `gorm:"column:status;type:varchar(50)" json:"status"`
	DiskSizeGb         int64  `gorm:"column:disk_size_gb" json:"diskSizeGb"`
	StorageBytes       int64  `gorm:"column:storage_bytes" json:"storageBytes"`
	StorageBytesStatus string `gorm:"column:storage_bytes_status;type:varchar(50)" json:"storageBytesStatus"`
	DownloadBytes      int64  `gorm:"column:download_bytes" json:"downloadBytes"`
	SnapshotType       string `gorm:"column:snapshot_type;type:varchar(50)" json:"snapshotType"`
	Architecture       string `gorm:"column:architecture;type:varchar(50)" json:"architecture"`
	SelfLink           string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp  string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint   string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

	// Source fields
	SourceDisk   string `gorm:"column:source_disk;type:text" json:"sourceDisk"`
	SourceDiskId string `gorm:"column:source_disk_id;type:varchar(255)" json:"sourceDiskId"`

	// Recovery
	SourceDiskForRecoveryCheckpoint string `gorm:"column:source_disk_for_recovery_checkpoint;type:text" json:"sourceDiskForRecoveryCheckpoint"`

	// Flags
	AutoCreated               bool `gorm:"column:auto_created" json:"autoCreated"`
	SatisfiesPzi              bool `gorm:"column:satisfies_pzi" json:"satisfiesPzi"`
	SatisfiesPzs              bool `gorm:"column:satisfies_pzs" json:"satisfiesPzs"`
	EnableConfidentialCompute bool `gorm:"column:enable_confidential_compute" json:"enableConfidentialCompute"`

	// JSONB fields
	SnapshotEncryptionKeyJSON   jsonb.JSON `gorm:"column:snapshot_encryption_key_json;type:jsonb" json:"snapshotEncryptionKey"`
	SourceDiskEncryptionKeyJSON jsonb.JSON `gorm:"column:source_disk_encryption_key_json;type:jsonb" json:"sourceDiskEncryptionKey"`
	GuestOsFeaturesJSON         jsonb.JSON `gorm:"column:guest_os_features_json;type:jsonb" json:"guestOsFeatures"`
	StorageLocationsJSON        jsonb.JSON `gorm:"column:storage_locations_json;type:jsonb" json:"storageLocations"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeSnapshot) TableName() string {
	return "bronze_history.gcp_compute_snapshots"
}

// GCPComputeSnapshotLabel stores historical snapshots of snapshot labels.
// Links via SnapshotHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeSnapshotLabel struct {
	HistoryID         uint `gorm:"primaryKey"`
	SnapshotHistoryID uint `gorm:"column:snapshot_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeSnapshotLabel) TableName() string {
	return "bronze_history.gcp_compute_snapshot_labels"
}

// GCPComputeSnapshotLicense stores historical snapshots of snapshot licenses.
// Links via SnapshotHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeSnapshotLicense struct {
	HistoryID         uint `gorm:"primaryKey"`
	SnapshotHistoryID uint `gorm:"column:snapshot_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// License field
	License string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeSnapshotLicense) TableName() string {
	return "bronze_history.gcp_compute_snapshot_licenses"
}
