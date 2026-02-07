package serviceaccount

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

func (h *HistoryService) CreateHistory(tx *gorm.DB, sa *bronze.GCPIAMServiceAccount, now time.Time) error {
	hist := toServiceAccountHistory(sa, now)
	return tx.Create(&hist).Error
}

func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPIAMServiceAccount, diff *ServiceAccountDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	if err := tx.Model(&bronze_history.GCPIAMServiceAccount{}).
		Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Create new history
	hist := toServiceAccountHistory(new, now)
	return tx.Create(&hist).Error
}

func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	return tx.Model(&bronze_history.GCPIAMServiceAccount{}).
		Where("resource_id = ? AND valid_to IS NULL", resourceID).
		Update("valid_to", now).Error
}

func toServiceAccountHistory(sa *bronze.GCPIAMServiceAccount, now time.Time) bronze_history.GCPIAMServiceAccount {
	return bronze_history.GCPIAMServiceAccount{
		ResourceID:     sa.ResourceID,
		ValidFrom:      now,
		ValidTo:        nil,
		Name:           sa.Name,
		Email:          sa.Email,
		DisplayName:    sa.DisplayName,
		Description:    sa.Description,
		Oauth2ClientId: sa.Oauth2ClientId,
		Disabled:       sa.Disabled,
		Etag:           sa.Etag,
		ProjectID:      sa.ProjectID,
		CollectedAt:    sa.CollectedAt,
	}
}
