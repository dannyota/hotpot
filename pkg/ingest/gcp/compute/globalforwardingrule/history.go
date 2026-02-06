package globalforwardingrule

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for global forwarding rules.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new global forwarding rule and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, fr *bronze.GCPComputeGlobalForwardingRule, now time.Time) error {
	// Create global forwarding rule history
	frHist := toGlobalForwardingRuleHistory(fr, now)
	if err := tx.Create(&frHist).Error; err != nil {
		return err
	}

	// Create children history with global_forwarding_rule_history_id
	return h.createChildrenHistory(tx, frHist.HistoryID, fr, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeGlobalForwardingRule, diff *GlobalForwardingRuleDiff, now time.Time) error {
	// Get current global forwarding rule history
	var currentHist bronze_history.GCPComputeGlobalForwardingRule
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If forwarding rule-level fields changed, close old and create new history
	if diff.IsChanged {
		// Close old global forwarding rule history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new global forwarding rule history
		frHist := toGlobalForwardingRuleHistory(new, now)
		if err := tx.Create(&frHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, frHist.HistoryID, new, now)
	}

	// Global forwarding rule unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted global forwarding rule.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current global forwarding rule history
	var currentHist bronze_history.GCPComputeGlobalForwardingRule
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close global forwarding rule history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, globalForwardingRuleHistoryID uint, fr *bronze.GCPComputeGlobalForwardingRule, now time.Time) error {
	for _, label := range fr.Labels {
		labelHist := toLabelHistory(&label, globalForwardingRuleHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, globalForwardingRuleHistoryID uint, now time.Time) error {
	if err := tx.Table("bronze_history.gcp_compute_global_forwarding_rule_labels").
		Where("global_forwarding_rule_history_id = ? AND valid_to IS NULL", globalForwardingRuleHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}
	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, globalForwardingRuleHistoryID uint, new *bronze.GCPComputeGlobalForwardingRule, diff *GlobalForwardingRuleDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, globalForwardingRuleHistoryID, new.Labels, now); err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, globalForwardingRuleHistoryID uint, labels []bronze.GCPComputeGlobalForwardingRuleLabel, now time.Time) error {
	if err := tx.Table("bronze_history.gcp_compute_global_forwarding_rule_labels").
		Where("global_forwarding_rule_history_id = ? AND valid_to IS NULL", globalForwardingRuleHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	for _, label := range labels {
		labelHist := toLabelHistory(&label, globalForwardingRuleHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toGlobalForwardingRuleHistory(fr *bronze.GCPComputeGlobalForwardingRule, now time.Time) bronze_history.GCPComputeGlobalForwardingRule {
	return bronze_history.GCPComputeGlobalForwardingRule{
		ResourceID:                                          fr.ResourceID,
		ValidFrom:                                           now,
		ValidTo:                                             nil,
		Name:                                                fr.Name,
		Description:                                         fr.Description,
		IPAddress:                                           fr.IPAddress,
		IPProtocol:                                          fr.IPProtocol,
		AllPorts:                                            fr.AllPorts,
		AllowGlobalAccess:                                   fr.AllowGlobalAccess,
		AllowPscGlobalAccess:                                fr.AllowPscGlobalAccess,
		BackendService:                                      fr.BackendService,
		BaseForwardingRule:                                  fr.BaseForwardingRule,
		CreationTimestamp:                                   fr.CreationTimestamp,
		ExternalManagedBackendBucketMigrationState:          fr.ExternalManagedBackendBucketMigrationState,
		ExternalManagedBackendBucketMigrationTestingPercentage: fr.ExternalManagedBackendBucketMigrationTestingPercentage,
		Fingerprint:                                         fr.Fingerprint,
		IpCollection:                                        fr.IpCollection,
		IpVersion:                                           fr.IpVersion,
		IsMirroringCollector:                                fr.IsMirroringCollector,
		LabelFingerprint:                                    fr.LabelFingerprint,
		LoadBalancingScheme:                                 fr.LoadBalancingScheme,
		Network:                                             fr.Network,
		NetworkTier:                                         fr.NetworkTier,
		NoAutomateDnsZone:                                   fr.NoAutomateDnsZone,
		PortRange:                                           fr.PortRange,
		PscConnectionId:                                     fr.PscConnectionId,
		PscConnectionStatus:                                 fr.PscConnectionStatus,
		Region:                                              fr.Region,
		SelfLink:                                            fr.SelfLink,
		SelfLinkWithId:                                      fr.SelfLinkWithId,
		ServiceLabel:                                        fr.ServiceLabel,
		ServiceName:                                         fr.ServiceName,
		Subnetwork:                                          fr.Subnetwork,
		Target:                                              fr.Target,
		PortsJSON:                                           fr.PortsJSON,
		SourceIpRangesJSON:                                  fr.SourceIpRangesJSON,
		MetadataFiltersJSON:                                 fr.MetadataFiltersJSON,
		ServiceDirectoryRegistrationsJSON:                   fr.ServiceDirectoryRegistrationsJSON,
		ProjectID:                                           fr.ProjectID,
		CollectedAt:                                         fr.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeGlobalForwardingRuleLabel, globalForwardingRuleHistoryID uint, now time.Time) bronze_history.GCPComputeGlobalForwardingRuleLabel {
	return bronze_history.GCPComputeGlobalForwardingRuleLabel{
		GlobalForwardingRuleHistoryID: globalForwardingRuleHistoryID,
		ValidFrom:                     now,
		ValidTo:                       nil,
		Key:                           label.Key,
		Value:                         label.Value,
	}
}
