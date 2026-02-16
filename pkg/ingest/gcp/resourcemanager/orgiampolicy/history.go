package orgiampolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcporgiampolicy"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcporgiampolicybinding"
)

// HistoryService manages organization IAM policy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new organization IAM policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *OrgIamPolicyData, now time.Time) error {
	// Create policy history
	policyHistory, err := tx.BronzeHistoryGCPOrgIamPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetResourceName(data.ResourceName).
		SetEtag(data.Etag).
		SetVersion(data.Version).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create org IAM policy history: %w", err)
	}

	// Create binding history
	for _, binding := range data.Bindings {
		create := tx.BronzeHistoryGCPOrgIamPolicyBinding.Create().
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

// UpdateHistory updates history records for a changed organization IAM policy.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPOrgIamPolicy, new *OrgIamPolicyData, diff *OrgIamPolicyDiff, now time.Time) error {
	// Get current policy history
	currentHistory, err := tx.BronzeHistoryGCPOrgIamPolicy.Query().
		Where(
			bronzehistorygcporgiampolicy.ResourceID(old.ID),
			bronzehistorygcporgiampolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current org IAM policy history: %w", err)
	}

	// Close current policy history if core fields changed
	if diff.IsChanged {
		// Close old binding history first
		_, err := tx.BronzeHistoryGCPOrgIamPolicyBinding.Update().
			Where(
				bronzehistorygcporgiampolicybinding.PolicyHistoryID(currentHistory.HistoryID),
				bronzehistorygcporgiampolicybinding.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close old binding history: %w", err)
		}

		// Close current policy history
		err = tx.BronzeHistoryGCPOrgIamPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current policy history: %w", err)
		}

		// Create new policy history
		newHistory, err := tx.BronzeHistoryGCPOrgIamPolicy.Create().
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
			create := tx.BronzeHistoryGCPOrgIamPolicyBinding.Create().
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
		_, err := tx.BronzeHistoryGCPOrgIamPolicyBinding.Update().
			Where(
				bronzehistorygcporgiampolicybinding.PolicyHistoryID(currentHistory.HistoryID),
				bronzehistorygcporgiampolicybinding.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close binding history: %w", err)
		}

		for _, binding := range new.Bindings {
			create := tx.BronzeHistoryGCPOrgIamPolicyBinding.Create().
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

// CloseHistory closes all history records for a deleted organization IAM policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current policy history
	currentHistory, err := tx.BronzeHistoryGCPOrgIamPolicy.Query().
		Where(
			bronzehistorygcporgiampolicy.ResourceID(resourceID),
			bronzehistorygcporgiampolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current org IAM policy history: %w", err)
	}

	// Close policy history
	err = tx.BronzeHistoryGCPOrgIamPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close policy history: %w", err)
	}

	// Close binding history
	_, err = tx.BronzeHistoryGCPOrgIamPolicyBinding.Update().
		Where(
			bronzehistorygcporgiampolicybinding.PolicyHistoryID(currentHistory.HistoryID),
			bronzehistorygcporgiampolicybinding.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close binding history: %w", err)
	}

	return nil
}
