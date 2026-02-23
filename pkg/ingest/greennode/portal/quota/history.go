package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodeportalquota"
)

// HistoryService handles history tracking for quotas.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new quota.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *QuotaData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodePortalQuota.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetType(data.Type).
		SetLimitValue(data.LimitValue).
		SetUsedValue(data.UsedValue).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create quota history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodePortalQuota, new *QuotaData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodePortalQuota.Query().
		Where(
			bronzehistorygreennodeportalquota.ResourceID(old.ID),
			bronzehistorygreennodeportalquota.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current quota history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodePortalQuota.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close quota history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodePortalQuota.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetType(new.Type).
		SetLimitValue(new.LimitValue).
		SetUsedValue(new.UsedValue).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new quota history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted quota.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodePortalQuota.Query().
		Where(
			bronzehistorygreennodeportalquota.ResourceID(resourceID),
			bronzehistorygreennodeportalquota.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current quota history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodePortalQuota.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close quota history: %w", err)
	}
	return nil
}
