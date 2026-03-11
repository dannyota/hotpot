package glbpackage

import (
	"context"
	"fmt"
	"time"

	entglb "danny.vn/hotpot/pkg/storage/ent/greennode/glb"
	"danny.vn/hotpot/pkg/storage/ent/greennode/glb/bronzehistorygreennodeglbglobalpackage"
)

// HistoryService handles history tracking for global packages.
type HistoryService struct {
	entClient *entglb.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entglb.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new global package.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entglb.Tx, data *GLBPackageData, now time.Time) error {
	create := tx.BronzeHistoryGreenNodeGLBGlobalPackage.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetDescriptionEn(data.DescriptionEn).
		SetEnabled(data.Enabled).
		SetBaseSku(data.BaseSku).
		SetBaseConnectionRate(data.BaseConnectionRate).
		SetBaseDomesticTrafficTotal(data.BaseDomesticTrafficTotal).
		SetBaseNonDomesticTrafficTotal(data.BaseNonDomesticTrafficTotal).
		SetConnectionSku(data.ConnectionSku).
		SetDomesticTrafficSku(data.DomesticTrafficSku).
		SetNonDomesticTrafficSku(data.NonDomesticTrafficSku).
		SetCreatedAtAPI(data.CreatedAtAPI).
		SetUpdatedAtAPI(data.UpdatedAtAPI).
		SetProjectID(data.ProjectID)

	if data.DetailJSON != nil {
		create.SetDetailJSON(data.DetailJSON)
	}
	if data.VlbPackagesJSON != nil {
		create.SetVlbPackagesJSON(data.VlbPackagesJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create package history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entglb.Tx, old *entglb.BronzeGreenNodeGLBGlobalPackage, new *GLBPackageData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeGLBGlobalPackage.Query().
		Where(
			bronzehistorygreennodeglbglobalpackage.ResourceID(old.ID),
			bronzehistorygreennodeglbglobalpackage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current package history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeGLBGlobalPackage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close package history: %w", err)
	}

	create := tx.BronzeHistoryGreenNodeGLBGlobalPackage.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetDescriptionEn(new.DescriptionEn).
		SetEnabled(new.Enabled).
		SetBaseSku(new.BaseSku).
		SetBaseConnectionRate(new.BaseConnectionRate).
		SetBaseDomesticTrafficTotal(new.BaseDomesticTrafficTotal).
		SetBaseNonDomesticTrafficTotal(new.BaseNonDomesticTrafficTotal).
		SetConnectionSku(new.ConnectionSku).
		SetDomesticTrafficSku(new.DomesticTrafficSku).
		SetNonDomesticTrafficSku(new.NonDomesticTrafficSku).
		SetCreatedAtAPI(new.CreatedAtAPI).
		SetUpdatedAtAPI(new.UpdatedAtAPI).
		SetProjectID(new.ProjectID)

	if new.DetailJSON != nil {
		create.SetDetailJSON(new.DetailJSON)
	}
	if new.VlbPackagesJSON != nil {
		create.SetVlbPackagesJSON(new.VlbPackagesJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new package history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted package.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entglb.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeGLBGlobalPackage.Query().
		Where(
			bronzehistorygreennodeglbglobalpackage.ResourceID(resourceID),
			bronzehistorygreennodeglbglobalpackage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entglb.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current package history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeGLBGlobalPackage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close package history: %w", err)
	}
	return nil
}
