package bronze_history

import "time"

type GCPIAMServiceAccountKey struct {
	HistoryID  uint   `gorm:"primaryKey"`
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"name"`

	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Name                string    `gorm:"column:name;type:text;not null" json:"fullName"`
	ServiceAccountEmail string    `gorm:"column:service_account_email;type:varchar(255);not null" json:"serviceAccountEmail"`
	KeyOrigin           string    `gorm:"column:key_origin;type:varchar(50)" json:"keyOrigin"`
	KeyType             string    `gorm:"column:key_type;type:varchar(50)" json:"keyType"`
	KeyAlgorithm        string    `gorm:"column:key_algorithm;type:varchar(50)" json:"keyAlgorithm"`
	ValidAfterTime      time.Time `gorm:"column:valid_after_time" json:"validAfterTime"`
	ValidBeforeTime     time.Time `gorm:"column:valid_before_time" json:"validBeforeTime"`
	Disabled            bool      `gorm:"column:disabled" json:"disabled"`

	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPIAMServiceAccountKey) TableName() string {
	return "bronze_history.gcp_iam_service_account_keys"
}
