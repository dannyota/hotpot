package firewall

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute/bronzegcpcomputefirewall"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute/bronzegcpcomputefirewallallowed"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute/bronzegcpcomputefirewalldenied"
)

// Service handles GCP Compute firewall ingestion.
type Service struct {
	client    *Client
	entClient *entcompute.Client
	history   *HistoryService
}

// NewService creates a new firewall ingestion service.
func NewService(client *Client, entClient *entcompute.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestPage converts raw proto firewalls and saves them as one batch.
func (s *Service) IngestPage(ctx context.Context, firewalls []*computepb.Firewall, projectID string, collectedAt time.Time) (int, error) {
	dataList := make([]*FirewallData, 0, len(firewalls))
	for _, fw := range firewalls {
		data, err := ConvertFirewall(fw, projectID, collectedAt)
		if err != nil {
			return 0, fmt.Errorf("convert firewall: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err := s.saveFirewallBatch(ctx, dataList, time.Now()); err != nil {
		return 0, fmt.Errorf("save firewall page: %w", err)
	}

	return len(dataList), nil
}

// saveFirewallBatch saves a batch of firewalls in a single transaction.
func (s *Service) saveFirewallBatch(ctx context.Context, firewalls []*FirewallData, now time.Time) error {
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

	for _, firewallData := range firewalls {
		// Load existing firewall with allowed and denied
		existing, err := tx.BronzeGCPComputeFirewall.Query().
			Where(bronzegcpcomputefirewall.ID(firewallData.ID)).
			WithAllowed().
			WithDenied().
			First(ctx)
		if err != nil && !entcompute.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing firewall %s: %w", firewallData.Name, err)
		}

		// Compute diff
		diff := DiffFirewallData(existing, firewallData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeFirewall.UpdateOneID(firewallData.ID).
				SetCollectedAt(firewallData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for firewall %s: %w", firewallData.Name, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := deleteFirewallChildren(ctx, tx, firewallData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for firewall %s: %w", firewallData.Name, err)
			}
		}

		// Create or update firewall
		var savedFirewall *entcompute.BronzeGCPComputeFirewall
		if existing == nil {
			// Create new firewall
			create := tx.BronzeGCPComputeFirewall.Create().
				SetID(firewallData.ID).
				SetName(firewallData.Name).
				SetDescription(firewallData.Description).
				SetSelfLink(firewallData.SelfLink).
				SetCreationTimestamp(firewallData.CreationTimestamp).
				SetNetwork(firewallData.Network).
				SetPriority(firewallData.Priority).
				SetDirection(firewallData.Direction).
				SetDisabled(firewallData.Disabled).
				SetProjectID(firewallData.ProjectID).
				SetCollectedAt(firewallData.CollectedAt).
				SetFirstCollectedAt(firewallData.CollectedAt)

			if firewallData.SourceRangesJSON != nil {
				create.SetSourceRangesJSON(firewallData.SourceRangesJSON)
			}
			if firewallData.DestinationRangesJSON != nil {
				create.SetDestinationRangesJSON(firewallData.DestinationRangesJSON)
			}
			if firewallData.SourceTagsJSON != nil {
				create.SetSourceTagsJSON(firewallData.SourceTagsJSON)
			}
			if firewallData.TargetTagsJSON != nil {
				create.SetTargetTagsJSON(firewallData.TargetTagsJSON)
			}
			if firewallData.SourceServiceAccountsJSON != nil {
				create.SetSourceServiceAccountsJSON(firewallData.SourceServiceAccountsJSON)
			}
			if firewallData.TargetServiceAccountsJSON != nil {
				create.SetTargetServiceAccountsJSON(firewallData.TargetServiceAccountsJSON)
			}
			if firewallData.LogConfigJSON != nil {
				create.SetLogConfigJSON(firewallData.LogConfigJSON)
			}

			savedFirewall, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create firewall %s: %w", firewallData.Name, err)
			}
		} else {
			// Update existing firewall
			update := tx.BronzeGCPComputeFirewall.UpdateOneID(firewallData.ID).
				SetName(firewallData.Name).
				SetDescription(firewallData.Description).
				SetSelfLink(firewallData.SelfLink).
				SetCreationTimestamp(firewallData.CreationTimestamp).
				SetNetwork(firewallData.Network).
				SetPriority(firewallData.Priority).
				SetDirection(firewallData.Direction).
				SetDisabled(firewallData.Disabled).
				SetProjectID(firewallData.ProjectID).
				SetCollectedAt(firewallData.CollectedAt)

			if firewallData.SourceRangesJSON != nil {
				update.SetSourceRangesJSON(firewallData.SourceRangesJSON)
			}
			if firewallData.DestinationRangesJSON != nil {
				update.SetDestinationRangesJSON(firewallData.DestinationRangesJSON)
			}
			if firewallData.SourceTagsJSON != nil {
				update.SetSourceTagsJSON(firewallData.SourceTagsJSON)
			}
			if firewallData.TargetTagsJSON != nil {
				update.SetTargetTagsJSON(firewallData.TargetTagsJSON)
			}
			if firewallData.SourceServiceAccountsJSON != nil {
				update.SetSourceServiceAccountsJSON(firewallData.SourceServiceAccountsJSON)
			}
			if firewallData.TargetServiceAccountsJSON != nil {
				update.SetTargetServiceAccountsJSON(firewallData.TargetServiceAccountsJSON)
			}
			if firewallData.LogConfigJSON != nil {
				update.SetLogConfigJSON(firewallData.LogConfigJSON)
			}

			savedFirewall, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update firewall %s: %w", firewallData.Name, err)
			}
		}

		// Create new children
		if err := createFirewallChildren(ctx, tx, savedFirewall, firewallData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for firewall %s: %w", firewallData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, firewallData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for firewall %s: %w", firewallData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, firewallData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for firewall %s: %w", firewallData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteFirewallChildren deletes allowed and denied rules for a firewall.
func deleteFirewallChildren(ctx context.Context, tx *entcompute.Tx, firewallID string) error {
	// Delete allowed rules
	_, err := tx.BronzeGCPComputeFirewallAllowed.Delete().
		Where(bronzegcpcomputefirewallallowed.HasFirewallRefWith(bronzegcpcomputefirewall.ID(firewallID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete allowed rules: %w", err)
	}

	// Delete denied rules
	_, err = tx.BronzeGCPComputeFirewallDenied.Delete().
		Where(bronzegcpcomputefirewalldenied.HasFirewallRefWith(bronzegcpcomputefirewall.ID(firewallID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete denied rules: %w", err)
	}

	return nil
}

// createFirewallChildren creates allowed and denied rules for a firewall.
func createFirewallChildren(ctx context.Context, tx *entcompute.Tx, savedFirewall *entcompute.BronzeGCPComputeFirewall, firewallData *FirewallData) error {
	// Create allowed rules
	for _, allowed := range firewallData.Allowed {
		create := tx.BronzeGCPComputeFirewallAllowed.Create().
			SetIPProtocol(allowed.IpProtocol).
			SetFirewallRef(savedFirewall)
		if allowed.PortsJSON != nil {
			create.SetPortsJSON(allowed.PortsJSON)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create allowed rule: %w", err)
		}
	}

	// Create denied rules
	for _, denied := range firewallData.Denied {
		create := tx.BronzeGCPComputeFirewallDenied.Create().
			SetIPProtocol(denied.IpProtocol).
			SetFirewallRef(savedFirewall)
		if denied.PortsJSON != nil {
			create.SetPortsJSON(denied.PortsJSON)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create denied rule: %w", err)
		}
	}

	return nil
}

// DeleteStaleFirewalls removes firewalls that were not collected in the latest run.
// Also closes history records for deleted firewalls.
func (s *Service) DeleteStaleFirewalls(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale firewalls
	staleFirewalls, err := tx.BronzeGCPComputeFirewall.Query().
		Where(
			bronzegcpcomputefirewall.ProjectID(projectID),
			bronzegcpcomputefirewall.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale firewall
	for _, fw := range staleFirewalls {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, fw.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for firewall %s: %w", fw.ID, err)
		}

		// Delete firewall (allowed/denied will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeFirewall.DeleteOne(fw).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete firewall %s: %w", fw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
