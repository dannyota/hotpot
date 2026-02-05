package bronze

import (
	"time"
)

// GCPProject represents a GCP project in the bronze layer.
// Fields preserve raw API response data from cloudresourcemanager.projects.search.
type GCPProject struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ProjectID is the user-assigned project ID (e.g., "my-project-123").
	ProjectID     string `gorm:"primaryKey;column:project_id;type:varchar(255)" json:"projectId"`
	ProjectNumber string `gorm:"column:project_number;type:varchar(255);not null;uniqueIndex" json:"name"`
	DisplayName   string `gorm:"column:display_name;type:varchar(255)" json:"displayName"`
	State         string `gorm:"column:state;type:varchar(50);index" json:"state"`
	Parent        string `gorm:"column:parent;type:varchar(255);index" json:"parent"`
	CreateTime    string `gorm:"column:create_time;type:varchar(50)" json:"createTime"`
	UpdateTime    string `gorm:"column:update_time;type:varchar(50)" json:"updateTime"`
	DeleteTime    string `gorm:"column:delete_time;type:varchar(50)" json:"deleteTime"`
	Etag          string `gorm:"column:etag;type:varchar(255)" json:"etag"`

	// Collection metadata
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships
	Labels []GCPProjectLabel `gorm:"foreignKey:ProjectID;references:ProjectID" json:"labels,omitempty"`
}

func (GCPProject) TableName() string {
	return "bronze.gcp_projects"
}

// GCPProjectLabel represents a label attached to a GCP project.
type GCPProjectLabel struct {
	ID        uint   `gorm:"primaryKey"`
	ProjectID string `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	Key       string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value     string `gorm:"column:value;type:text" json:"value"`
}

func (GCPProjectLabel) TableName() string {
	return "bronze.gcp_project_labels"
}
