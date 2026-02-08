package forwardingrule

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputeforwardingrule"
	"hotpot/pkg/storage/ent/bronzegcpcomputeforwardingrulelabel"
)

// Service handles GCP Compute forwarding rule ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new forwarding rule ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for forwarding rule ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of forwarding rule ingestion.
type IngestResult struct {
	ProjectID          string
	ForwardingRuleCount int
	CollectedAt        time.Time
	DurationMillis     int64
}

// Ingest fetches forwarding rules from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch forwarding rules from GCP
	forwardingRules, err := s.client.ListForwardingRules(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list forwarding rules: %w", err)
	}

	// Convert to data structs
	ruleDataList := make([]*ForwardingRuleData, 0, len(forwardingRules))
	for _, fr := range forwardingRules {
		ruleData, err := ConvertForwardingRule(fr, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert forwarding rule: %w", err)
		}
		ruleDataList = append(ruleDataList, ruleData)
	}

	// Save to database
	if err := s.saveForwardingRules(ctx, ruleDataList); err != nil {
		return nil, fmt.Errorf("failed to save forwarding rules: %w", err)
	}

	return &IngestResult{
		ProjectID:          params.ProjectID,
		ForwardingRuleCount: len(ruleDataList),
		CollectedAt:        collectedAt,
		DurationMillis:     time.Since(startTime).Milliseconds(),
	}, nil
}

// saveForwardingRules saves forwarding rules to the database with history tracking.
func (s *Service) saveForwardingRules(ctx context.Context, forwardingRules []*ForwardingRuleData) error {
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
		// Load existing forwarding rule with labels
		existing, err := tx.BronzeGCPComputeForwardingRule.Query().
			Where(bronzegcpcomputeforwardingrule.ID(ruleData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing forwarding rule %s: %w", ruleData.ID, err)
		}

		// Compute diff
		diff := DiffForwardingRuleData(existing, ruleData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeForwardingRule.UpdateOneID(ruleData.ID).
				SetCollectedAt(ruleData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for forwarding rule %s: %w", ruleData.ID, err)
			}
			continue
		}

		// Delete old labels if updating
		if existing != nil {
			_, err := tx.BronzeGCPComputeForwardingRuleLabel.Delete().
				Where(bronzegcpcomputeforwardingrulelabel.HasForwardingRuleWith(bronzegcpcomputeforwardingrule.ID(ruleData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old labels for forwarding rule %s: %w", ruleData.ID, err)
			}
		}

		// Create or update forwarding rule
		var savedRule *ent.BronzeGCPComputeForwardingRule
		if existing == nil {
			// Create new forwarding rule
			create := tx.BronzeGCPComputeForwardingRule.Create().
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
				SetCollectedAt(ruleData.CollectedAt).
				SetFirstCollectedAt(ruleData.CollectedAt)

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
				return fmt.Errorf("failed to create forwarding rule %s: %w", ruleData.ID, err)
			}

			// Create labels for new forwarding rule
			for _, label := range ruleData.Labels {
				_, err := tx.BronzeGCPComputeForwardingRuleLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetForwardingRule(savedRule).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for forwarding rule %s: %w", ruleData.ID, err)
				}
			}
		} else {
			// Update existing forwarding rule
			update := tx.BronzeGCPComputeForwardingRule.UpdateOneID(ruleData.ID).
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
				return fmt.Errorf("failed to update forwarding rule %s: %w", ruleData.ID, err)
			}

			// Create new labels
			for _, label := range ruleData.Labels {
				_, err := tx.BronzeGCPComputeForwardingRuleLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetForwardingRule(savedRule).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for forwarding rule %s: %w", ruleData.ID, err)
				}
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, ruleData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for forwarding rule %s: %w", ruleData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, ruleData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for forwarding rule %s: %w", ruleData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleForwardingRules removes forwarding rules that were not collected in the latest run.
// Also closes history records for deleted forwarding rules.
func (s *Service) DeleteStaleForwardingRules(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale forwarding rules
	staleRules, err := tx.BronzeGCPComputeForwardingRule.Query().
		Where(
			bronzegcpcomputeforwardingrule.ProjectID(projectID),
			bronzegcpcomputeforwardingrule.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale forwarding rule
	for _, rule := range staleRules {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, rule.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for forwarding rule %s: %w", rule.ID, err)
		}

		// Delete forwarding rule (labels will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeForwardingRule.DeleteOne(rule).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete forwarding rule %s: %w", rule.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
