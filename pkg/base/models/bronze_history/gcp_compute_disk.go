package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeDisk stores historical snapshots of GCP Compute persistent disks.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeDisk struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Disk has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All disk fields (same as bronze.GCPComputeDisk)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Zone              string `gorm:"column:zone;type:text" json:"zone"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Type              string `gorm:"column:type;type:text" json:"type"`
	Status            string `gorm:"column:status;type:varchar(50)" json:"status"`
	SizeGb            int64  `gorm:"column:size_gb" json:"sizeGb"`
	Architecture      string `gorm:"column:architecture;type:varchar(50)" json:"architecture"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`

	// Attachment timestamps
	LastAttachTimestamp string `gorm:"column:last_attach_timestamp;type:varchar(50)" json:"lastAttachTimestamp"`
	LastDetachTimestamp string `gorm:"column:last_detach_timestamp;type:varchar(50)" json:"lastDetachTimestamp"`

	// Source fields
	SourceImage      string `gorm:"column:source_image;type:text" json:"sourceImage"`
	SourceImageId    string `gorm:"column:source_image_id;type:varchar(255)" json:"sourceImageId"`
	SourceSnapshot   string `gorm:"column:source_snapshot;type:text" json:"sourceSnapshot"`
	SourceSnapshotId string `gorm:"column:source_snapshot_id;type:varchar(255)" json:"sourceSnapshotId"`
	SourceDisk       string `gorm:"column:source_disk;type:text" json:"sourceDisk"`
	SourceDiskId     string `gorm:"column:source_disk_id;type:varchar(255)" json:"sourceDiskId"`

	// Performance settings
	ProvisionedIops        int64 `gorm:"column:provisioned_iops" json:"provisionedIops"`
	ProvisionedThroughput  int64 `gorm:"column:provisioned_throughput" json:"provisionedThroughput"`
	PhysicalBlockSizeBytes int64 `gorm:"column:physical_block_size_bytes" json:"physicalBlockSizeBytes"`

	// Security
	EnableConfidentialCompute bool       `gorm:"column:enable_confidential_compute" json:"enableConfidentialCompute"`
	DiskEncryptionKeyJSON     jsonb.JSON `gorm:"column:disk_encryption_key_json;type:jsonb" json:"diskEncryptionKey"`

	// JSON arrays
	UsersJSON            jsonb.JSON `gorm:"column:users_json;type:jsonb" json:"users"`
	ReplicaZonesJSON     jsonb.JSON `gorm:"column:replica_zones_json;type:jsonb" json:"replicaZones"`
	ResourcePoliciesJSON jsonb.JSON `gorm:"column:resource_policies_json;type:jsonb" json:"resourcePolicies"`
	GuestOsFeaturesJSON  jsonb.JSON `gorm:"column:guest_os_features_json;type:jsonb" json:"guestOsFeatures"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeDisk) TableName() string {
	return "bronze_history.gcp_compute_disks"
}

// GCPComputeDiskLabel stores historical snapshots of disk labels.
// Links via DiskHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeDiskLabel struct {
	HistoryID     uint `gorm:"primaryKey"`
	DiskHistoryID uint `gorm:"column:disk_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeDiskLabel) TableName() string {
	return "bronze_history.gcp_compute_disk_labels"
}

// GCPComputeDiskLicense stores historical snapshots of disk licenses.
// Links via DiskHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeDiskLicense struct {
	HistoryID     uint `gorm:"primaryKey"`
	DiskHistoryID uint `gorm:"column:disk_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// License field
	License string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeDiskLicense) TableName() string {
	return "bronze_history.gcp_compute_disk_licenses"
}
