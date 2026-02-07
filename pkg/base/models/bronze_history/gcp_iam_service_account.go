package bronze_history

import "time"

type GCPIAMServiceAccount struct {
	HistoryID  uint   `gorm:"primaryKey"`
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"uniqueId"`

	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Name           string `gorm:"column:name;type:text;not null" json:"name"`
	Email          string `gorm:"column:email;type:varchar(255);not null" json:"email"`
	DisplayName    string `gorm:"column:display_name;type:varchar(255)" json:"displayName"`
	Description    string `gorm:"column:description;type:text" json:"description"`
	Oauth2ClientId string `gorm:"column:oauth2_client_id;type:varchar(255)" json:"oauth2ClientId"`
	Disabled       bool   `gorm:"column:disabled" json:"disabled"`
	Etag           string `gorm:"column:etag;type:varchar(255)" json:"etag"`

	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPIAMServiceAccount) TableName() string {
	return "bronze_history.gcp_iam_service_accounts"
}
