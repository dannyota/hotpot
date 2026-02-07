package healthcheck

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for health checks.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates a history record for a new health check.
func (h *HistoryService) CreateHistory(tx *gorm.DB, check *bronze.GCPComputeHealthCheck, now time.Time) error {
	hist := toHealthCheckHistory(check, now)
	return tx.Create(&hist).Error
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeHealthCheck, diff *HealthCheckDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil // nothing to update
	}

	// Get current health check history
	var currentHist bronze_history.GCPComputeHealthCheck
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// Close old history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Create new history
	hist := toHealthCheckHistory(new, now)
	return tx.Create(&hist).Error
}

// CloseHistory closes history records for a deleted health check.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	var currentHist bronze_history.GCPComputeHealthCheck
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	return tx.Model(&currentHist).Update("valid_to", now).Error
}

// toHealthCheckHistory converts a bronze health check to a history record.
func toHealthCheckHistory(check *bronze.GCPComputeHealthCheck, now time.Time) bronze_history.GCPComputeHealthCheck {
	return bronze_history.GCPComputeHealthCheck{
		ResourceID:         check.ResourceID,
		ValidFrom:          now,
		ValidTo:            nil,
		Name:               check.Name,
		Description:        check.Description,
		CreationTimestamp:  check.CreationTimestamp,
		SelfLink:           check.SelfLink,
		Type:               check.Type,
		Region:             check.Region,
		CheckIntervalSec:  check.CheckIntervalSec,
		TimeoutSec:         check.TimeoutSec,
		HealthyThreshold:   check.HealthyThreshold,
		UnhealthyThreshold: check.UnhealthyThreshold,
		TcpHealthCheckJSON:   check.TcpHealthCheckJSON,
		HttpHealthCheckJSON:  check.HttpHealthCheckJSON,
		HttpsHealthCheckJSON: check.HttpsHealthCheckJSON,
		Http2HealthCheckJSON: check.Http2HealthCheckJSON,
		SslHealthCheckJSON:   check.SslHealthCheckJSON,
		GrpcHealthCheckJSON:  check.GrpcHealthCheckJSON,
		LogConfigJSON:        check.LogConfigJSON,
		ProjectID:          check.ProjectID,
		CollectedAt:        check.CollectedAt,
	}
}
