package serviceaccountkey

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

type HistoryService struct {
	db *gorm.DB
}

func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

func (h *HistoryService) CreateHistory(tx *gorm.DB, key *bronze.GCPIAMServiceAccountKey, now time.Time) error {
	hist := toKeyHistory(key, now)
	return tx.Create(&hist).Error
}

func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPIAMServiceAccountKey, diff *ServiceAccountKeyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	if err := tx.Model(&bronze_history.GCPIAMServiceAccountKey{}).
		Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	hist := toKeyHistory(new, now)
	return tx.Create(&hist).Error
}

func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	return tx.Model(&bronze_history.GCPIAMServiceAccountKey{}).
		Where("resource_id = ? AND valid_to IS NULL", resourceID).
		Update("valid_to", now).Error
}

func toKeyHistory(key *bronze.GCPIAMServiceAccountKey, now time.Time) bronze_history.GCPIAMServiceAccountKey {
	return bronze_history.GCPIAMServiceAccountKey{
		ResourceID:          key.ResourceID,
		ValidFrom:           now,
		ValidTo:             nil,
		Name:                key.Name,
		ServiceAccountEmail: key.ServiceAccountEmail,
		KeyOrigin:           key.KeyOrigin,
		KeyType:             key.KeyType,
		KeyAlgorithm:        key.KeyAlgorithm,
		ValidAfterTime:      key.ValidAfterTime,
		ValidBeforeTime:     key.ValidBeforeTime,
		Disabled:            key.Disabled,
		ProjectID:           key.ProjectID,
		CollectedAt:         key.CollectedAt,
	}
}
