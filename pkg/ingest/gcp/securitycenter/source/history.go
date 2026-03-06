package source

import (
	"context"
	"fmt"
	"time"

	entsecuritycenter "danny.vn/hotpot/pkg/storage/ent/gcp/securitycenter"
	"danny.vn/hotpot/pkg/storage/ent/gcp/securitycenter/bronzehistorygcpsecuritycentersource"
)

// HistoryService manages SCC source history tracking.
type HistoryService struct {
	entClient *entsecuritycenter.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entsecuritycenter.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new SCC source.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entsecuritycenter.Tx, data *SourceData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPSecurityCenterSource.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetDisplayName(data.DisplayName).
		SetDescription(data.Description).
		SetCanonicalName(data.CanonicalName).
		SetOrganizationID(data.OrganizationID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create SCC source history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed SCC source.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entsecuritycenter.Tx, old *entsecuritycenter.BronzeGCPSecurityCenterSource, new *SourceData, diff *SourceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSecurityCenterSource.Query().
		Where(
			bronzehistorygcpsecuritycentersource.ResourceID(old.ID),
			bronzehistorygcpsecuritycentersource.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current SCC source history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPSecurityCenterSource.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current SCC source history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPSecurityCenterSource.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetDisplayName(new.DisplayName).
			SetDescription(new.Description).
			SetCanonicalName(new.CanonicalName).
			SetOrganizationID(new.OrganizationID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new SCC source history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted SCC source.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entsecuritycenter.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSecurityCenterSource.Query().
		Where(
			bronzehistorygcpsecuritycentersource.ResourceID(resourceID),
			bronzehistorygcpsecuritycentersource.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entsecuritycenter.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current SCC source history: %w", err)
	}

	err = tx.BronzeHistoryGCPSecurityCenterSource.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close SCC source history: %w", err)
	}

	return nil
}
