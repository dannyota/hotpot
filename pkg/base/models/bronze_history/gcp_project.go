package bronze_history

import (
	"time"
)

// GCPProject stores historical snapshots of GCP projects.
// Uses ProjectID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPProject struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Project has unique project ID)
	ProjectID string `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All project fields (same as bronze.GCPProject)
	ProjectNumber string    `gorm:"column:project_number;type:varchar(255);not null" json:"name"`
	DisplayName   string    `gorm:"column:display_name;type:varchar(255)" json:"displayName"`
	State         string    `gorm:"column:state;type:varchar(50)" json:"state"`
	Parent        string    `gorm:"column:parent;type:varchar(255)" json:"parent"`
	CreateTime    string    `gorm:"column:create_time;type:varchar(50)" json:"createTime"`
	UpdateTime    string    `gorm:"column:update_time;type:varchar(50)" json:"updateTime"`
	DeleteTime    string    `gorm:"column:delete_time;type:varchar(50)" json:"deleteTime"`
	Etag          string    `gorm:"column:etag;type:varchar(255)" json:"etag"`
	CollectedAt   time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPProject) TableName() string {
	return "bronze_history.gcp_projects"
}

// GCPProjectLabel stores historical snapshots of project labels.
// Links via ProjectHistoryID, has own valid_from/valid_to for granular tracking.
type GCPProjectLabel struct {
	HistoryID        uint `gorm:"primaryKey"`
	ProjectHistoryID uint `gorm:"column:project_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Key   string `gorm:"column:key;type:varchar(255)" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPProjectLabel) TableName() string {
	return "bronze_history.gcp_project_labels"
}
