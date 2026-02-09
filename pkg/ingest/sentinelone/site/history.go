package site

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorys1site"
)

// HistoryService handles history tracking for sites.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *SiteData) *ent.BronzeHistoryS1SiteCreate {
	create := tx.BronzeHistoryS1Site.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetAccountID(data.AccountID).
		SetAccountName(data.AccountName).
		SetState(data.State).
		SetSiteType(data.SiteType).
		SetSuite(data.Suite).
		SetCreator(data.Creator).
		SetCreatorID(data.CreatorID).
		SetHealthStatus(data.HealthStatus).
		SetActiveLicenses(data.ActiveLicenses).
		SetTotalLicenses(data.TotalLicenses).
		SetUnlimitedLicenses(data.UnlimitedLicenses).
		SetIsDefault(data.IsDefault).
		SetDescription(data.Description)

	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}
	if data.Expiration != nil {
		create.SetExpiration(*data.Expiration)
	}

	return create
}

// CreateHistory creates a history record for a new site.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SiteData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create site history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed site.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1Site, new *SiteData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Site.Query().
		Where(
			bronzehistorys1site.ResourceID(old.ID),
			bronzehistorys1site.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current site history: %w", err)
	}

	if err := tx.BronzeHistoryS1Site.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close site history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new site history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted site.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Site.Query().
		Where(
			bronzehistorys1site.ResourceID(resourceID),
			bronzehistorys1site.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current site history: %w", err)
	}

	if err := tx.BronzeHistoryS1Site.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close site history: %w", err)
	}

	return nil
}
