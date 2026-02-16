package asset

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcloudassetasset"
)

// HistoryService manages Cloud Asset asset history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Cloud Asset asset.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AssetData, now time.Time) error {
	create := tx.BronzeHistoryGCPCloudAssetAsset.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetAssetType(data.AssetType).
		SetOrganizationID(data.OrganizationID)

	if data.UpdateTime != "" {
		create.SetUpdateTime(data.UpdateTime)
	}
	if data.ResourceJSON != nil {
		create.SetResourceJSON(data.ResourceJSON)
	}
	if data.IamPolicyJSON != nil {
		create.SetIamPolicyJSON(data.IamPolicyJSON)
	}
	if data.OrgPolicyJSON != nil {
		create.SetOrgPolicyJSON(data.OrgPolicyJSON)
	}
	if data.AccessPolicyJSON != nil {
		create.SetAccessPolicyJSON(data.AccessPolicyJSON)
	}
	if data.OsInventoryJSON != nil {
		create.SetOsInventoryJSON(data.OsInventoryJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create Cloud Asset asset history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Cloud Asset asset.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPCloudAssetAsset, new *AssetData, diff *AssetDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudAssetAsset.Query().
		Where(
			bronzehistorygcpcloudassetasset.ResourceID(old.ID),
			bronzehistorygcpcloudassetasset.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Cloud Asset asset history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPCloudAssetAsset.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Cloud Asset asset history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPCloudAssetAsset.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetAssetType(new.AssetType).
			SetOrganizationID(new.OrganizationID)

		if new.UpdateTime != "" {
			create.SetUpdateTime(new.UpdateTime)
		}
		if new.ResourceJSON != nil {
			create.SetResourceJSON(new.ResourceJSON)
		}
		if new.IamPolicyJSON != nil {
			create.SetIamPolicyJSON(new.IamPolicyJSON)
		}
		if new.OrgPolicyJSON != nil {
			create.SetOrgPolicyJSON(new.OrgPolicyJSON)
		}
		if new.AccessPolicyJSON != nil {
			create.SetAccessPolicyJSON(new.AccessPolicyJSON)
		}
		if new.OsInventoryJSON != nil {
			create.SetOsInventoryJSON(new.OsInventoryJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("failed to create new Cloud Asset asset history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Cloud Asset asset.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudAssetAsset.Query().
		Where(
			bronzehistorygcpcloudassetasset.ResourceID(resourceID),
			bronzehistorygcpcloudassetasset.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Cloud Asset asset history: %w", err)
	}

	err = tx.BronzeHistoryGCPCloudAssetAsset.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Cloud Asset asset history: %w", err)
	}

	return nil
}
