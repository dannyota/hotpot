package firewall

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedofirewall"
)

// Service handles DigitalOcean Firewall ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Firewall ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of Firewall ingestion.
type IngestResult struct {
	FirewallCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all Firewalls from DigitalOcean and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiFirewalls, err := s.client.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list firewalls: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allFirewalls []*FirewallData
	for _, v := range apiFirewalls {
		allFirewalls = append(allFirewalls, ConvertFirewall(v, collectedAt))
	}

	if err := s.saveFirewalls(ctx, allFirewalls); err != nil {
		return nil, fmt.Errorf("save firewalls: %w", err)
	}

	return &IngestResult{
		FirewallCount:  len(allFirewalls),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveFirewalls(ctx context.Context, firewalls []*FirewallData) error {
	if len(firewalls) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range firewalls {
		existing, err := tx.BronzeDOFirewall.Query().
			Where(bronzedofirewall.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing firewall %s: %w", data.ResourceID, err)
		}

		diff := DiffFirewallData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOFirewall.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for firewall %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOFirewall.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetInboundRulesJSON(data.InboundRulesJSON).
				SetOutboundRulesJSON(data.OutboundRulesJSON).
				SetDropletIdsJSON(data.DropletIdsJSON).
				SetTagsJSON(data.TagsJSON).
				SetAPICreatedAt(data.APICreatedAt).
				SetPendingChangesJSON(data.PendingChangesJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create firewall %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for firewall %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOFirewall.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetInboundRulesJSON(data.InboundRulesJSON).
				SetOutboundRulesJSON(data.OutboundRulesJSON).
				SetDropletIdsJSON(data.DropletIdsJSON).
				SetTagsJSON(data.TagsJSON).
				SetAPICreatedAt(data.APICreatedAt).
				SetPendingChangesJSON(data.PendingChangesJSON).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update firewall %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for firewall %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes Firewalls that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeDOFirewall.Query().
		Where(bronzedofirewall.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doFirewall := range stale {
		if err := s.history.CloseHistory(ctx, tx, doFirewall.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for firewall %s: %w", doFirewall.ID, err)
		}

		if err := tx.BronzeDOFirewall.DeleteOne(doFirewall).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete firewall %s: %w", doFirewall.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
