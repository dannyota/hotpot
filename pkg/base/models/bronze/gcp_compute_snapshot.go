package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeSnapshot represents a GCP Compute Engine snapshot in the bronze layer.
// Fields preserve raw API response data from compute.snapshots.list.
type GCPComputeSnapshot struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50);index" json:"status"`
	DiskSizeGb        int64  `gorm:"column:disk_size_gb" json:"diskSizeGb"`
	StorageBytes      int64  `gorm:"column:storage_bytes" json:"storageBytes"`
	StorageBytesStatus string `gorm:"column:storage_bytes_status;type:varchar(50)" json:"storageBytesStatus"`
	DownloadBytes     int64  `gorm:"column:download_bytes" json:"downloadBytes"`
	SnapshotType      string `gorm:"column:snapshot_type;type:varchar(50)" json:"snapshotType"`
	Architecture      string `gorm:"column:architecture;type:varchar(50)" json:"architecture"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

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

	// SnapshotEncryptionKeyJSON contains snapshot encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	SnapshotEncryptionKeyJSON jsonb.JSON `gorm:"column:snapshot_encryption_key_json;type:jsonb" json:"snapshotEncryptionKey"`

	// SourceDiskEncryptionKeyJSON contains source disk encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	SourceDiskEncryptionKeyJSON jsonb.JSON `gorm:"column:source_disk_encryption_key_json;type:jsonb" json:"sourceDiskEncryptionKey"`

	// GuestOsFeaturesJSON contains guest OS features enabled on this snapshot.
	//
	//	[{"type": "VIRTIO_SCSI_MULTIQUEUE"}, {"type": "UEFI_COMPATIBLE"}]
	GuestOsFeaturesJSON jsonb.JSON `gorm:"column:guest_os_features_json;type:jsonb" json:"guestOsFeatures"`

	// StorageLocationsJSON contains regions where snapshot data is stored.
	//
	//	["us-central1", "us"]
	StorageLocationsJSON jsonb.JSON `gorm:"column:storage_locations_json;type:jsonb" json:"storageLocations"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Labels   []GCPComputeSnapshotLabel   `gorm:"foreignKey:SnapshotResourceID;references:ResourceID" json:"labels,omitempty"`
	Licenses []GCPComputeSnapshotLicense `gorm:"foreignKey:SnapshotResourceID;references:ResourceID" json:"licenses,omitempty"`
}

func (GCPComputeSnapshot) TableName() string {
	return "bronze.gcp_compute_snapshots"
}

// GCPComputeSnapshotLabel represents a label attached to a GCP Compute snapshot.
type GCPComputeSnapshotLabel struct {
	ID                 uint   `gorm:"primaryKey"`
	SnapshotResourceID string `gorm:"column:snapshot_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value              string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeSnapshotLabel) TableName() string {
	return "bronze.gcp_compute_snapshot_labels"
}

// GCPComputeSnapshotLicense represents a license attached to a GCP Compute snapshot.
type GCPComputeSnapshotLicense struct {
	ID                 uint   `gorm:"primaryKey"`
	SnapshotResourceID string `gorm:"column:snapshot_resource_id;type:varchar(255);not null;index" json:"-"`
	License            string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeSnapshotLicense) TableName() string {
	return "bronze.gcp_compute_snapshot_licenses"
}
