package bronze

import (
	"time"
)

// GCPComputeTargetInstance represents a GCP Compute Engine target instance in the bronze layer.
// Fields preserve raw API response data from compute.targetInstances.aggregatedList.
type GCPComputeTargetInstance struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
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
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`
}

func (GCPComputeTargetInstance) TableName() string {
	return "bronze.gcp_compute_target_instances"
}
