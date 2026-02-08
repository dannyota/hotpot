package globalforwardingrule

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputeglobalforwardingrule"
	"hotpot/pkg/storage/ent/bronzegcpcomputeglobalforwardingrulelabel"
)

// Service handles GCP Compute global forwarding rule ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new global forwarding rule ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for global forwarding rule ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of global forwarding rule ingestion.
type IngestResult struct {
	ProjectID                string
	GlobalForwardingRuleCount int
	CollectedAt              time.Time
	DurationMillis           int64
}

// Ingest fetches global forwarding rules from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch global forwarding rules from GCP
	forwardingRules, err := s.client.ListGlobalForwardingRules(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list global forwarding rules: %w", err)
	}

	// Convert to data structs
	ruleDataList := make([]*GlobalForwardingRuleData, 0, len(forwardingRules))
	for _, fr := range forwardingRules {
		ruleData, err := ConvertGlobalForwardingRule(fr, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert global forwarding rule: %w", err)
		}
		ruleDataList = append(ruleDataList, ruleData)
	}

	// Save to database
	if err := s.saveGlobalForwardingRules(ctx, ruleDataList); err != nil {
		return nil, fmt.Errorf("failed to save global forwarding rules: %w", err)
	}

	return &IngestResult{
		ProjectID:                params.ProjectID,
		GlobalForwardingRuleCount: len(ruleDataList),
		CollectedAt:              collectedAt,
		DurationMillis:           time.Since(startTime).Milliseconds(),
	}, nil
}

// saveGlobalForwardingRules saves global forwarding rules to the database with history tracking.
func (s *Service) saveGlobalForwardingRules(ctx context.Context, forwardingRules []*GlobalForwardingRuleData) error {
	if len(forwardingRules) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, ruleData := range forwardingRules {
		// Load existing global forwarding rule with labels
		existing, err := tx.BronzeGCPComputeGlobalForwardingRule.Query().
			Where(bronzegcpcomputeglobalforwardingrule.ID(ruleData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing global forwarding rule %s: %w", ruleData.ID, err)
		}

		// Compute diff
		diff := DiffGlobalForwardingRuleData(existing, ruleData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeGlobalForwardingRule.UpdateOneID(ruleData.ID).
				SetCollectedAt(ruleData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for global forwarding rule %s: %w", ruleData.ID, err)
			}
			continue
		}

		// Delete old labels if updating
		if existing != nil {
			_, err := tx.BronzeGCPComputeGlobalForwardingRuleLabel.Delete().
				Where(bronzegcpcomputeglobalforwardingrulelabel.HasGlobalForwardingRuleWith(bronzegcpcomputeglobalforwardingrule.ID(ruleData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old labels for global forwarding rule %s: %w", ruleData.ID, err)
			}
		}

		// Create or update global forwarding rule
		var savedRule *ent.BronzeGCPComputeGlobalForwardingRule
		if existing == nil {
			// Create new global forwarding rule
			create := tx.BronzeGCPComputeGlobalForwardingRule.Create().
				SetID(ruleData.ID).
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
				SetCollectedAt(ruleData.CollectedAt)

			if ruleData.PortsJSON != nil {
				create.SetPortsJSON(ruleData.PortsJSON)
			}
			if ruleData.SourceIpRangesJSON != nil {
				create.SetSourceIPRangesJSON(ruleData.SourceIpRangesJSON)
			}
			if ruleData.MetadataFiltersJSON != nil {
				create.SetMetadataFiltersJSON(ruleData.MetadataFiltersJSON)
			}
			if ruleData.ServiceDirectoryRegistrationsJSON != nil {
				create.SetServiceDirectoryRegistrationsJSON(ruleData.ServiceDirectoryRegistrationsJSON)
			}

			savedRule, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create global forwarding rule %s: %w", ruleData.ID, err)
			}

			// Create labels for new global forwarding rule
			for _, label := range ruleData.Labels {
				_, err := tx.BronzeGCPComputeGlobalForwardingRuleLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetGlobalForwardingRule(savedRule).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for global forwarding rule %s: %w", ruleData.ID, err)
				}
			}
		} else {
			// Update existing global forwarding rule
			update := tx.BronzeGCPComputeGlobalForwardingRule.UpdateOneID(ruleData.ID).
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
				SetCollectedAt(ruleData.CollectedAt)

			if ruleData.PortsJSON != nil {
				update.SetPortsJSON(ruleData.PortsJSON)
			}
			if ruleData.SourceIpRangesJSON != nil {
				update.SetSourceIPRangesJSON(ruleData.SourceIpRangesJSON)
			}
			if ruleData.MetadataFiltersJSON != nil {
				update.SetMetadataFiltersJSON(ruleData.MetadataFiltersJSON)
			}
			if ruleData.ServiceDirectoryRegistrationsJSON != nil {
				update.SetServiceDirectoryRegistrationsJSON(ruleData.ServiceDirectoryRegistrationsJSON)
			}

			savedRule, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update global forwarding rule %s: %w", ruleData.ID, err)
			}

			// Create new labels
			for _, label := range ruleData.Labels {
				_, err := tx.BronzeGCPComputeGlobalForwardingRuleLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetGlobalForwardingRule(savedRule).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for global forwarding rule %s: %w", ruleData.ID, err)
				}
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, ruleData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for global forwarding rule %s: %w", ruleData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, ruleData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for global forwarding rule %s: %w", ruleData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleGlobalForwardingRules removes global forwarding rules that were not collected in the latest run.
// Also closes history records for deleted global forwarding rules.
func (s *Service) DeleteStaleGlobalForwardingRules(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale global forwarding rules
	staleRules, err := tx.BronzeGCPComputeGlobalForwardingRule.Query().
		Where(
			bronzegcpcomputeglobalforwardingrule.ProjectID(projectID),
			bronzegcpcomputeglobalforwardingrule.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale global forwarding rule
	for _, rule := range staleRules {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, rule.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for global forwarding rule %s: %w", rule.ID, err)
		}

		// Delete global forwarding rule (labels will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeGlobalForwardingRule.DeleteOne(rule).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete global forwarding rule %s: %w", rule.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
