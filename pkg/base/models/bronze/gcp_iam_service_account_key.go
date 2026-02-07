package bronze

import "time"

type GCPIAMServiceAccountKey struct {
	ResourceID          string    `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"name"`
	Name                string    `gorm:"column:name;type:text;not null" json:"fullName"`
	ServiceAccountEmail string    `gorm:"column:service_account_email;type:varchar(255);not null;index" json:"serviceAccountEmail"`
	KeyOrigin           string    `gorm:"column:key_origin;type:varchar(50)" json:"keyOrigin"`
	KeyType             string    `gorm:"column:key_type;type:varchar(50)" json:"keyType"`
	KeyAlgorithm        string    `gorm:"column:key_algorithm;type:varchar(50)" json:"keyAlgorithm"`
	ValidAfterTime      time.Time `gorm:"column:valid_after_time" json:"validAfterTime"`
	ValidBeforeTime     time.Time `gorm:"column:valid_before_time" json:"validBeforeTime"`
	Disabled            bool      `gorm:"column:disabled" json:"disabled"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`
}

func (GCPIAMServiceAccountKey) TableName() string {
	return "bronze.gcp_iam_service_account_keys"
}
