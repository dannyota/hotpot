package osimage

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodecomputeosimage"
)

// HistoryService handles history tracking for OS images.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new OS image.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *OSImageData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeComputeOSImage.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetImageType(data.ImageType).
		SetImageVersion(data.ImageVersion).
		SetNillableLicence(data.Licence).
		SetNillableLicenseKey(data.LicenseKey).
		SetDescription(data.Description).
		SetZoneID(data.ZoneID).
		SetFlavorZoneIds(data.FlavorZoneIDs).
		SetDefaultTagIds(data.DefaultTagIDs).
		SetPackageLimitCPU(data.PackageLimitCpu).
		SetPackageLimitMemory(data.PackageLimitMemory).
		SetPackageLimitDiskSize(data.PackageLimitDiskSize).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create os image history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeComputeOSImage, new *OSImageData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeOSImage.Query().
		Where(
			bronzehistorygreennodecomputeosimage.ResourceID(old.ID),
			bronzehistorygreennodecomputeosimage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current os image history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeOSImage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close os image history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeComputeOSImage.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetImageType(new.ImageType).
		SetImageVersion(new.ImageVersion).
		SetNillableLicence(new.Licence).
		SetNillableLicenseKey(new.LicenseKey).
		SetDescription(new.Description).
		SetZoneID(new.ZoneID).
		SetFlavorZoneIds(new.FlavorZoneIDs).
		SetDefaultTagIds(new.DefaultTagIDs).
		SetPackageLimitCPU(new.PackageLimitCpu).
		SetPackageLimitMemory(new.PackageLimitMemory).
		SetPackageLimitDiskSize(new.PackageLimitDiskSize).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new os image history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted OS image.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeOSImage.Query().
		Where(
			bronzehistorygreennodecomputeosimage.ResourceID(resourceID),
			bronzehistorygreennodecomputeosimage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current os image history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeOSImage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close os image history: %w", err)
	}
	return nil
}
