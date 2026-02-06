package bronze_history

import (
	"time"
)

// GCPComputeTargetInstance stores historical snapshots of GCP Compute target instances.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeTargetInstance struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (TargetInstance has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All target instance fields (same as bronze.GCPComputeTargetInstance)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Zone              string `gorm:"column:zone;type:text" json:"zone"`
	Instance          string `gorm:"column:instance;type:text" json:"instance"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	NatPolicy         string `gorm:"column:nat_policy;type:varchar(50)" json:"natPolicy"`
	SecurityPolicy    string `gorm:"column:security_policy;type:text" json:"securityPolicy"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeTargetInstance) TableName() string {
	return "bronze_history.gcp_compute_target_instances"
}
