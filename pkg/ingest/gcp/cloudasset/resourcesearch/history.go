package resourcesearch

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcloudassetresourcesearch"
)

// HistoryService manages resource search history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new resource search result.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ResourceSearchData, now time.Time) error {
	create := tx.BronzeHistoryGCPCloudAssetResourceSearch.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetAssetType(data.AssetType).
		SetOrganizationID(data.OrganizationID)

	if data.Project != "" {
		create.SetProject(data.Project)
	}
	if data.DisplayName != "" {
		create.SetDisplayName(data.DisplayName)
	}
	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.Location != "" {
		create.SetLocation(data.Location)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.NetworkTagsJSON != nil {
		create.SetNetworkTagsJSON(data.NetworkTagsJSON)
	}
	if data.AdditionalAttributesJSON != nil {
		create.SetAdditionalAttributesJSON(data.AdditionalAttributesJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create resource search history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed resource search result.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPCloudAssetResourceSearch, new *ResourceSearchData, diff *ResourceSearchDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudAssetResourceSearch.Query().
		Where(
			bronzehistorygcpcloudassetresourcesearch.ResourceID(old.ID),
			bronzehistorygcpcloudassetresourcesearch.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current resource search history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPCloudAssetResourceSearch.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current resource search history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPCloudAssetResourceSearch.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetAssetType(new.AssetType).
			SetOrganizationID(new.OrganizationID)

		if new.Project != "" {
			create.SetProject(new.Project)
		}
		if new.DisplayName != "" {
			create.SetDisplayName(new.DisplayName)
		}
		if new.Description != "" {
			create.SetDescription(new.Description)
		}
		if new.Location != "" {
			create.SetLocation(new.Location)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.NetworkTagsJSON != nil {
			create.SetNetworkTagsJSON(new.NetworkTagsJSON)
		}
		if new.AdditionalAttributesJSON != nil {
			create.SetAdditionalAttributesJSON(new.AdditionalAttributesJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("failed to create new resource search history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted resource search result.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudAssetResourceSearch.Query().
		Where(
			bronzehistorygcpcloudassetresourcesearch.ResourceID(resourceID),
			bronzehistorygcpcloudassetresourcesearch.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current resource search history: %w", err)
	}

	err = tx.BronzeHistoryGCPCloudAssetResourceSearch.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close resource search history: %w", err)
	}

	return nil
}
