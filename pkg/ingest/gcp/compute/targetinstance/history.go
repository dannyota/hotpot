package targetinstance

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for target instances.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates a history record for a new target instance.
func (h *HistoryService) CreateHistory(tx *gorm.DB, ti *bronze.GCPComputeTargetInstance, now time.Time) error {
	hist := toTargetInstanceHistory(ti, now)
	return tx.Create(&hist).Error
}

// UpdateHistory closes old history and creates new history if changed.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeTargetInstance, diff *TargetInstanceDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	var currentHist bronze_history.GCPComputeTargetInstance
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Create new history
	hist := toTargetInstanceHistory(new, now)
	return tx.Create(&hist).Error
}

// CloseHistory closes history records for a deleted target instance.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	var currentHist bronze_history.GCPComputeTargetInstance
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	return tx.Model(&currentHist).Update("valid_to", now).Error
}

// toTargetInstanceHistory converts a bronze target instance to a history record.
func toTargetInstanceHistory(ti *bronze.GCPComputeTargetInstance, now time.Time) bronze_history.GCPComputeTargetInstance {
	return bronze_history.GCPComputeTargetInstance{
		ResourceID:        ti.ResourceID,
		ValidFrom:         now,
		ValidTo:           nil,
		Name:              ti.Name,
		Description:       ti.Description,
		Zone:              ti.Zone,
		Instance:          ti.Instance,
		Network:           ti.Network,
		NatPolicy:         ti.NatPolicy,
		SecurityPolicy:    ti.SecurityPolicy,
		SelfLink:          ti.SelfLink,
		CreationTimestamp: ti.CreationTimestamp,
		ProjectID:         ti.ProjectID,
		CollectedAt:       ti.CollectedAt,
	}
}
