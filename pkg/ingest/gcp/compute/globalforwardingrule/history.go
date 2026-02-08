package globalforwardingrule

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeglobalforwardingrule"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeglobalforwardingrulelabel"
)

// HistoryService manages global forwarding rule history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new global forwarding rule.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, ruleData *GlobalForwardingRuleData, now time.Time) error {
	// Create global forwarding rule history
	ruleHistory, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.Create().
		SetResourceID(ruleData.ID).
		SetValidFrom(now).
		SetCollectedAt(ruleData.CollectedAt).
		SetName(ruleData.Name).
		SetDescription(ruleData.Description).
		SetIPAddress(ruleData.IPAddress).
		SetIPProtocol(ruleData.IPProtocol).
		SetAllPorts(ruleData.AllPorts).
		SetAllowGlobalAccess(ruleData.AllowGlobalAccess).
		SetAllowPscGlobalAccess(ruleData.AllowPscGlobalAccess).
		SetBackendService(ruleData.BackendService).
		SetBaseForwardingRule(ruleData.BaseForwardingRule).
		SetCreationTimestamp(ruleData.CreationTimestamp).
		SetExternalManagedBackendBucketMigrationState(ruleData.ExternalManagedBackendBucketMigrationState).
		SetExternalManagedBackendBucketMigrationTestingPercentage(ruleData.ExternalManagedBackendBucketMigrationTestingPercentage).
		SetFingerprint(ruleData.Fingerprint).
		SetIPCollection(ruleData.IpCollection).
		SetIPVersion(ruleData.IpVersion).
		SetIsMirroringCollector(ruleData.IsMirroringCollector).
		SetLabelFingerprint(ruleData.LabelFingerprint).
		SetLoadBalancingScheme(ruleData.LoadBalancingScheme).
		SetNetwork(ruleData.Network).
		SetNetworkTier(ruleData.NetworkTier).
		SetNoAutomateDNSZone(ruleData.NoAutomateDnsZone).
		SetPortRange(ruleData.PortRange).
		SetPscConnectionID(ruleData.PscConnectionId).
		SetPscConnectionStatus(ruleData.PscConnectionStatus).
		SetRegion(ruleData.Region).
		SetSelfLink(ruleData.SelfLink).
		SetSelfLinkWithID(ruleData.SelfLinkWithId).
		SetServiceLabel(ruleData.ServiceLabel).
		SetServiceName(ruleData.ServiceName).
		SetSubnetwork(ruleData.Subnetwork).
		SetTarget(ruleData.Target).
		SetProjectID(ruleData.ProjectID).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create global forwarding rule history: %w", err)
	}

	// Set JSON fields if present
	if ruleData.PortsJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(ruleHistory).
			SetPortsJSON(ruleData.PortsJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set ports json: %w", err)
		}
	}
	if ruleData.SourceIpRangesJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(ruleHistory).
			SetSourceIPRangesJSON(ruleData.SourceIpRangesJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set source ip ranges json: %w", err)
		}
	}
	if ruleData.MetadataFiltersJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(ruleHistory).
			SetMetadataFiltersJSON(ruleData.MetadataFiltersJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set metadata filters json: %w", err)
		}
	}
	if ruleData.ServiceDirectoryRegistrationsJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(ruleHistory).
			SetServiceDirectoryRegistrationsJSON(ruleData.ServiceDirectoryRegistrationsJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set service directory registrations json: %w", err)
		}
	}

	// Create label history
	for _, label := range ruleData.Labels {
		_, err := tx.BronzeHistoryGCPComputeGlobalForwardingRuleLabel.Create().
			SetGlobalForwardingRuleHistoryID(ruleHistory.HistoryID).
			SetValidFrom(now).
			SetKey(label.Key).
			SetValue(label.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}

// UpdateHistory updates history records for a changed global forwarding rule.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeGlobalForwardingRule, new *GlobalForwardingRuleData, diff *GlobalForwardingRuleDiff, now time.Time) error {
	// Get current global forwarding rule history
	currentHistory, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.Query().
		Where(
			bronzehistorygcpcomputeglobalforwardingrule.ResourceID(old.ID),
			bronzehistorygcpcomputeglobalforwardingrule.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current global forwarding rule history: %w", err)
	}

	// Close current global forwarding rule history if core fields changed
	if diff.IsChanged {
		// Close old label history first
		_, err := tx.BronzeHistoryGCPComputeGlobalForwardingRuleLabel.Update().
			Where(
				bronzehistorygcpcomputeglobalforwardingrulelabel.GlobalForwardingRuleHistoryID(currentHistory.HistoryID),
				bronzehistorygcpcomputeglobalforwardingrulelabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close old label history: %w", err)
		}

		// Close current global forwarding rule history
		err = tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current global forwarding rule history: %w", err)
		}

		// Create new global forwarding rule history
		newHistory, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetIPAddress(new.IPAddress).
			SetIPProtocol(new.IPProtocol).
			SetAllPorts(new.AllPorts).
			SetAllowGlobalAccess(new.AllowGlobalAccess).
			SetAllowPscGlobalAccess(new.AllowPscGlobalAccess).
			SetBackendService(new.BackendService).
			SetBaseForwardingRule(new.BaseForwardingRule).
			SetCreationTimestamp(new.CreationTimestamp).
			SetExternalManagedBackendBucketMigrationState(new.ExternalManagedBackendBucketMigrationState).
			SetExternalManagedBackendBucketMigrationTestingPercentage(new.ExternalManagedBackendBucketMigrationTestingPercentage).
			SetFingerprint(new.Fingerprint).
			SetIPCollection(new.IpCollection).
			SetIPVersion(new.IpVersion).
			SetIsMirroringCollector(new.IsMirroringCollector).
			SetLabelFingerprint(new.LabelFingerprint).
			SetLoadBalancingScheme(new.LoadBalancingScheme).
			SetNetwork(new.Network).
			SetNetworkTier(new.NetworkTier).
			SetNoAutomateDNSZone(new.NoAutomateDnsZone).
			SetPortRange(new.PortRange).
			SetPscConnectionID(new.PscConnectionId).
			SetPscConnectionStatus(new.PscConnectionStatus).
			SetRegion(new.Region).
			SetSelfLink(new.SelfLink).
			SetSelfLinkWithID(new.SelfLinkWithId).
			SetServiceLabel(new.ServiceLabel).
			SetServiceName(new.ServiceName).
			SetSubnetwork(new.Subnetwork).
			SetTarget(new.Target).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new global forwarding rule history: %w", err)
		}

		// Set JSON fields if present
		if new.PortsJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(newHistory).
				SetPortsJSON(new.PortsJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set ports json: %w", err)
			}
		}
		if new.SourceIpRangesJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(newHistory).
				SetSourceIPRangesJSON(new.SourceIpRangesJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set source ip ranges json: %w", err)
			}
		}
		if new.MetadataFiltersJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(newHistory).
				SetMetadataFiltersJSON(new.MetadataFiltersJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set metadata filters json: %w", err)
			}
		}
		if new.ServiceDirectoryRegistrationsJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(newHistory).
				SetServiceDirectoryRegistrationsJSON(new.ServiceDirectoryRegistrationsJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set service directory registrations json: %w", err)
			}
		}

		// Create new label history linked to new global forwarding rule history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPComputeGlobalForwardingRuleLabel.Create().
				SetGlobalForwardingRuleHistoryID(newHistory.HistoryID).
				SetValidFrom(now).
				SetKey(label.Key).
				SetValue(label.Value).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create label history: %w", err)
			}
		}
	} else if diff.LabelsDiff.HasChanges {
		// Only labels changed - close old label history and create new ones
		_, err := tx.BronzeHistoryGCPComputeGlobalForwardingRuleLabel.Update().
			Where(
				bronzehistorygcpcomputeglobalforwardingrulelabel.GlobalForwardingRuleHistoryID(currentHistory.HistoryID),
				bronzehistorygcpcomputeglobalforwardingrulelabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close label history: %w", err)
		}

		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPComputeGlobalForwardingRuleLabel.Create().
				SetGlobalForwardingRuleHistoryID(currentHistory.HistoryID).
				SetValidFrom(now).
				SetKey(label.Key).
				SetValue(label.Value).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create label history: %w", err)
			}
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted global forwarding rule.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current global forwarding rule history
	currentHistory, err := tx.BronzeHistoryGCPComputeGlobalForwardingRule.Query().
		Where(
			bronzehistorygcpcomputeglobalforwardingrule.ResourceID(resourceID),
			bronzehistorygcpcomputeglobalforwardingrule.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current global forwarding rule history: %w", err)
	}

	// Close global forwarding rule history
	err = tx.BronzeHistoryGCPComputeGlobalForwardingRule.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close global forwarding rule history: %w", err)
	}

	// Close label history
	_, err = tx.BronzeHistoryGCPComputeGlobalForwardingRuleLabel.Update().
		Where(
			bronzehistorygcpcomputeglobalforwardingrulelabel.GlobalForwardingRuleHistoryID(currentHistory.HistoryID),
			bronzehistorygcpcomputeglobalforwardingrulelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	return nil
}
