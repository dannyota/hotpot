package iampolicysearch

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcloudassetiampolicysearch"
)

// HistoryService manages IAM policy search history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new IAM policy search result.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *IAMPolicySearchData, now time.Time) error {
	create := tx.BronzeHistoryGCPCloudAssetIAMPolicySearch.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetOrganizationID(data.OrganizationID)

	if data.AssetType != "" {
		create.SetAssetType(data.AssetType)
	}
	if data.Project != "" {
		create.SetProject(data.Project)
	}
	if data.Organization != "" {
		create.SetOrganization(data.Organization)
	}
	if data.FoldersJSON != nil {
		create.SetFoldersJSON(data.FoldersJSON)
	}
	if data.PolicyJSON != nil {
		create.SetPolicyJSON(data.PolicyJSON)
	}
	if data.ExplanationJSON != nil {
		create.SetExplanationJSON(data.ExplanationJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create IAM policy search history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed IAM policy search result.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPCloudAssetIAMPolicySearch, new *IAMPolicySearchData, diff *IAMPolicySearchDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudAssetIAMPolicySearch.Query().
		Where(
			bronzehistorygcpcloudassetiampolicysearch.ResourceID(old.ID),
			bronzehistorygcpcloudassetiampolicysearch.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current IAM policy search history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPCloudAssetIAMPolicySearch.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current IAM policy search history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPCloudAssetIAMPolicySearch.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetOrganizationID(new.OrganizationID)

		if new.AssetType != "" {
			create.SetAssetType(new.AssetType)
		}
		if new.Project != "" {
			create.SetProject(new.Project)
		}
		if new.Organization != "" {
			create.SetOrganization(new.Organization)
		}
		if new.FoldersJSON != nil {
			create.SetFoldersJSON(new.FoldersJSON)
		}
		if new.PolicyJSON != nil {
			create.SetPolicyJSON(new.PolicyJSON)
		}
		if new.ExplanationJSON != nil {
			create.SetExplanationJSON(new.ExplanationJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("failed to create new IAM policy search history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted IAM policy search result.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudAssetIAMPolicySearch.Query().
		Where(
			bronzehistorygcpcloudassetiampolicysearch.ResourceID(resourceID),
			bronzehistorygcpcloudassetiampolicysearch.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current IAM policy search history: %w", err)
	}

	err = tx.BronzeHistoryGCPCloudAssetIAMPolicySearch.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close IAM policy search history: %w", err)
	}

	return nil
}
