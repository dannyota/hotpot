package connector

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for VPC Access connectors.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates a history record for a new connector.
func (h *HistoryService) CreateHistory(tx *gorm.DB, c *bronze.GCPVpcAccessConnector, now time.Time) error {
	hist := toConnectorHistory(c, now)
	return tx.Create(&hist).Error
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPVpcAccessConnector, diff *ConnectorDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	if err := tx.Model(&bronze_history.GCPVpcAccessConnector{}).
		Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Create new history
	hist := toConnectorHistory(new, now)
	return tx.Create(&hist).Error
}

// CloseHistory closes history records for a deleted connector.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	return tx.Model(&bronze_history.GCPVpcAccessConnector{}).
		Where("resource_id = ? AND valid_to IS NULL", resourceID).
		Update("valid_to", now).Error
}

func toConnectorHistory(c *bronze.GCPVpcAccessConnector, now time.Time) bronze_history.GCPVpcAccessConnector {
	return bronze_history.GCPVpcAccessConnector{
		ResourceID:            c.ResourceID,
		ValidFrom:             now,
		ValidTo:               nil,
		Network:               c.Network,
		IpCidrRange:           c.IpCidrRange,
		State:                 c.State,
		MinThroughput:         c.MinThroughput,
		MaxThroughput:         c.MaxThroughput,
		MinInstances:          c.MinInstances,
		MaxInstances:          c.MaxInstances,
		MachineType:           c.MachineType,
		Region:                c.Region,
		SubnetJSON:            c.SubnetJSON,
		ConnectedProjectsJSON: c.ConnectedProjectsJSON,
		ProjectID:             c.ProjectID,
		CollectedAt:           c.CollectedAt,
	}
}
