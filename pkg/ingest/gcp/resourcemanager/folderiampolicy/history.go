package folderiampolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpfolderiampolicy"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpfolderiampolicybinding"
)

// HistoryService manages folder IAM policy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new folder IAM policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *FolderIamPolicyData, now time.Time) error {
	// Create policy history
	policyHistory, err := tx.BronzeHistoryGCPFolderIamPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetResourceName(data.ResourceName).
		SetEtag(data.Etag).
		SetVersion(data.Version).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create folder IAM policy history: %w", err)
	}

	// Create binding history
	for _, binding := range data.Bindings {
		create := tx.BronzeHistoryGCPFolderIamPolicyBinding.Create().
			SetPolicyHistoryID(policyHistory.HistoryID).
			SetValidFrom(now).
			SetRole(binding.Role)

		if binding.MembersJSON != nil {
			create.SetMembersJSON(binding.MembersJSON)
		}
		if binding.ConditionJSON != nil {
			create.SetConditionJSON(binding.ConditionJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("failed to create binding history: %w", err)
		}
	}

	return nil
}

// UpdateHistory updates history records for a changed folder IAM policy.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPFolderIamPolicy, new *FolderIamPolicyData, diff *FolderIamPolicyDiff, now time.Time) error {
	// Get current policy history
	currentHistory, err := tx.BronzeHistoryGCPFolderIamPolicy.Query().
		Where(
			bronzehistorygcpfolderiampolicy.ResourceID(old.ID),
			bronzehistorygcpfolderiampolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current folder IAM policy history: %w", err)
	}

	// Close current policy history if core fields changed
	if diff.IsChanged {
		// Close old binding history first
		_, err := tx.BronzeHistoryGCPFolderIamPolicyBinding.Update().
			Where(
				bronzehistorygcpfolderiampolicybinding.PolicyHistoryID(currentHistory.HistoryID),
				bronzehistorygcpfolderiampolicybinding.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close old binding history: %w", err)
		}

		// Close current policy history
		err = tx.BronzeHistoryGCPFolderIamPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current policy history: %w", err)
		}

		// Create new policy history
		newHistory, err := tx.BronzeHistoryGCPFolderIamPolicy.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetResourceName(new.ResourceName).
			SetEtag(new.Etag).
			SetVersion(new.Version).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new policy history: %w", err)
		}

		// Create new binding history linked to new policy history
		for _, binding := range new.Bindings {
			create := tx.BronzeHistoryGCPFolderIamPolicyBinding.Create().
				SetPolicyHistoryID(newHistory.HistoryID).
				SetValidFrom(now).
				SetRole(binding.Role)

			if binding.MembersJSON != nil {
				create.SetMembersJSON(binding.MembersJSON)
			}
			if binding.ConditionJSON != nil {
				create.SetConditionJSON(binding.ConditionJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				return fmt.Errorf("failed to create binding history: %w", err)
			}
		}
	} else if diff.BindingsDiff.HasChanges {
		// Only bindings changed - close old binding history and create new ones
		_, err := tx.BronzeHistoryGCPFolderIamPolicyBinding.Update().
			Where(
				bronzehistorygcpfolderiampolicybinding.PolicyHistoryID(currentHistory.HistoryID),
				bronzehistorygcpfolderiampolicybinding.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close binding history: %w", err)
		}

		for _, binding := range new.Bindings {
			create := tx.BronzeHistoryGCPFolderIamPolicyBinding.Create().
				SetPolicyHistoryID(currentHistory.HistoryID).
				SetValidFrom(now).
				SetRole(binding.Role)

			if binding.MembersJSON != nil {
				create.SetMembersJSON(binding.MembersJSON)
			}
			if binding.ConditionJSON != nil {
				create.SetConditionJSON(binding.ConditionJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				return fmt.Errorf("failed to create binding history: %w", err)
			}
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted folder IAM policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current policy history
	currentHistory, err := tx.BronzeHistoryGCPFolderIamPolicy.Query().
		Where(
			bronzehistorygcpfolderiampolicy.ResourceID(resourceID),
			bronzehistorygcpfolderiampolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current folder IAM policy history: %w", err)
	}

	// Close policy history
	err = tx.BronzeHistoryGCPFolderIamPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close policy history: %w", err)
	}

	// Close binding history
	_, err = tx.BronzeHistoryGCPFolderIamPolicyBinding.Update().
		Where(
			bronzehistorygcpfolderiampolicybinding.PolicyHistoryID(currentHistory.HistoryID),
			bronzehistorygcpfolderiampolicybinding.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close binding history: %w", err)
	}

	return nil
}
