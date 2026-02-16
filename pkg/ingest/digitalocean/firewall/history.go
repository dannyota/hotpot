package firewall

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydofirewall"
)

// HistoryService handles history tracking for Firewalls.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *FirewallData) *ent.BronzeHistoryDOFirewallCreate {
	return tx.BronzeHistoryDOFirewall.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetStatus(data.Status).
		SetInboundRulesJSON(data.InboundRulesJSON).
		SetOutboundRulesJSON(data.OutboundRulesJSON).
		SetDropletIdsJSON(data.DropletIdsJSON).
		SetTagsJSON(data.TagsJSON).
		SetAPICreatedAt(data.APICreatedAt).
		SetPendingChangesJSON(data.PendingChangesJSON)
}

// CreateHistory creates a history record for a new Firewall.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *FirewallData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create firewall history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Firewall.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOFirewall, new *FirewallData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOFirewall.Query().
		Where(
			bronzehistorydofirewall.ResourceID(old.ID),
			bronzehistorydofirewall.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current firewall history: %w", err)
	}

	if err := tx.BronzeHistoryDOFirewall.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close firewall history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new firewall history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Firewall.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOFirewall.Query().
		Where(
			bronzehistorydofirewall.ResourceID(resourceID),
			bronzehistorydofirewall.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current firewall history: %w", err)
	}

	if err := tx.BronzeHistoryDOFirewall.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close firewall history: %w", err)
	}

	return nil
}
