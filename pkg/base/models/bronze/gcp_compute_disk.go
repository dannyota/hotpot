package bronze

import (
	"time"
)

// GCPComputeDisk represents a GCP Compute Engine persistent disk in the bronze layer.
// Fields preserve raw API response data from compute.disks.list.
type GCPComputeDisk struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Zone              string `gorm:"column:zone;type:text" json:"zone"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Type              string `gorm:"column:type;type:text" json:"type"`
	Status            string `gorm:"column:status;type:varchar(50);index" json:"status"`
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
	ProvisionedIops       int64 `gorm:"column:provisioned_iops" json:"provisionedIops"`
	ProvisionedThroughput int64 `gorm:"column:provisioned_throughput" json:"provisionedThroughput"`
	PhysicalBlockSizeBytes int64 `gorm:"column:physical_block_size_bytes" json:"physicalBlockSizeBytes"`

	// Security
	EnableConfidentialCompute bool `gorm:"column:enable_confidential_compute" json:"enableConfidentialCompute"`

	// DiskEncryptionKeyJSON contains disk encryption configuration.
	//
	//	{
	//	  "sha256": "...",
	//	  "kmsKeyName": "projects/.../cryptoKeys/..."
	//	}
	DiskEncryptionKeyJSON string `gorm:"column:disk_encryption_key_json;type:jsonb" json:"diskEncryptionKey"`

	// UsersJSON contains list of instance URLs using this disk.
	//
	//	["projects/.../instances/vm1", "projects/.../instances/vm2"]
	UsersJSON string `gorm:"column:users_json;type:jsonb" json:"users"`

	// ReplicaZonesJSON contains zones for regional disk replication.
	//
	//	["zones/us-central1-a", "zones/us-central1-b"]
	ReplicaZonesJSON string `gorm:"column:replica_zones_json;type:jsonb" json:"replicaZones"`

	// ResourcePoliciesJSON contains attached resource policies for snapshots.
	//
	//	["projects/.../resourcePolicies/policy1"]
	ResourcePoliciesJSON string `gorm:"column:resource_policies_json;type:jsonb" json:"resourcePolicies"`

	// GuestOsFeaturesJSON contains guest OS features enabled on disk.
	//
	//	[{"type": "VIRTIO_SCSI_MULTIQUEUE"}, {"type": "UEFI_COMPATIBLE"}]
	GuestOsFeaturesJSON string `gorm:"column:guest_os_features_json;type:jsonb" json:"guestOsFeatures"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Labels   []GCPComputeDiskLabel   `gorm:"foreignKey:DiskResourceID;references:ResourceID" json:"labels,omitempty"`
	Licenses []GCPComputeDiskLicense `gorm:"foreignKey:DiskResourceID;references:ResourceID" json:"licenses,omitempty"`
}

func (GCPComputeDisk) TableName() string {
	return "bronze.gcp_compute_disks"
}

// GCPComputeDiskLabel represents a label attached to a GCP Compute disk.
type GCPComputeDiskLabel struct {
	ID             uint   `gorm:"primaryKey"`
	DiskResourceID string `gorm:"column:disk_resource_id;type:varchar(255);not null;index" json:"-"`
	Key            string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value          string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeDiskLabel) TableName() string {
	return "bronze.gcp_compute_disk_labels"
}

// GCPComputeDiskLicense represents a license attached to a GCP Compute disk.
type GCPComputeDiskLicense struct {
	ID             uint   `gorm:"primaryKey"`
	DiskResourceID string `gorm:"column:disk_resource_id;type:varchar(255);not null;index" json:"-"`
	License        string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeDiskLicense) TableName() string {
	return "bronze.gcp_compute_disk_licenses"
}
