package bronze

import "time"

type GCPIAMServiceAccount struct {
	ResourceID     string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"uniqueId"`
	Name           string `gorm:"column:name;type:text;not null" json:"name"`
	Email          string `gorm:"column:email;type:varchar(255);not null;index" json:"email"`
	DisplayName    string `gorm:"column:display_name;type:varchar(255)" json:"displayName"`
	Description    string `gorm:"column:description;type:text" json:"description"`
	Oauth2ClientId string `gorm:"column:oauth2_client_id;type:varchar(255)" json:"oauth2ClientId"`
	Disabled       bool   `gorm:"column:disabled" json:"disabled"`
	Etag           string `gorm:"column:etag;type:varchar(255)" json:"etag"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`
}

func (GCPIAMServiceAccount) TableName() string {
	return "bronze.gcp_iam_service_accounts"
}
