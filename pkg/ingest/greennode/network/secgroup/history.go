package secgroup

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodenetworksecgroup"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodenetworksecgrouprule"
)

// HistoryService handles history tracking for security groups.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new security group and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SecgroupData, now time.Time) error {
	sgHist, err := h.createSecgroupHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	return h.createRulesHistory(ctx, tx, sgHist.ID, data.Rules, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeNetworkSecgroup, new *SecgroupData, diff *SecgroupDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkSecgroup.Query().
		Where(
			bronzehistorygreennodenetworksecgroup.ResourceID(old.ID),
			bronzehistorygreennodenetworksecgroup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current secgroup history: %w", err)
	}

	if diff.IsChanged {
		// Close old history
		if err := tx.BronzeHistoryGreenNodeNetworkSecgroup.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close secgroup history: %w", err)
		}

		// Create new history
		sgHist, err := h.createSecgroupHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeRulesHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createRulesHistory(ctx, tx, sgHist.ID, new.Rules, now)
	}

	// Secgroup unchanged, check children
	if diff.RulesDiff.Changed {
		if err := h.closeRulesHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createRulesHistory(ctx, tx, currentHist.ID, new.Rules, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted security group.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkSecgroup.Query().
		Where(
			bronzehistorygreennodenetworksecgroup.ResourceID(resourceID),
			bronzehistorygreennodenetworksecgroup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current secgroup history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkSecgroup.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close secgroup history: %w", err)
	}

	return h.closeRulesHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createSecgroupHistory(ctx context.Context, tx *ent.Tx, data *SecgroupData, now time.Time, firstCollectedAt time.Time) (*ent.BronzeHistoryGreenNodeNetworkSecgroup, error) {
	hist, err := tx.BronzeHistoryGreenNodeNetworkSecgroup.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetStatus(data.Status).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create secgroup history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createRulesHistory(ctx context.Context, tx *ent.Tx, secgroupHistoryID uint, rules []SecgroupRuleData, now time.Time) error {
	for _, r := range rules {
		_, err := tx.BronzeHistoryGreenNodeNetworkSecgroupRule.Create().
			SetSecgroupHistoryID(secgroupHistoryID).
			SetValidFrom(now).
			SetRuleID(r.RuleID).
			SetDirection(r.Direction).
			SetEtherType(r.EtherType).
			SetProtocol(r.Protocol).
			SetDescription(r.Description).
			SetRemoteIPPrefix(r.RemoteIPPrefix).
			SetPortRangeMax(r.PortRangeMax).
			SetPortRangeMin(r.PortRangeMin).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create rule history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeRulesHistory(ctx context.Context, tx *ent.Tx, secgroupHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkSecgroupRule.Update().
		Where(
			bronzehistorygreennodenetworksecgrouprule.SecgroupHistoryID(secgroupHistoryID),
			bronzehistorygreennodenetworksecgrouprule.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close rules history: %w", err)
	}
	return nil
}
