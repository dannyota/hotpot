package constraint

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcporgpolicyconstraint"
)

type HistoryService struct {
	entClient *ent.Client
}

func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ConstraintData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPOrgPolicyConstraint.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetDisplayName(data.DisplayName).
		SetDescription(data.Description).
		SetConstraintDefault(data.ConstraintDefault).
		SetSupportsDryRun(data.SupportsDryRun).
		SetSupportsSimulation(data.SupportsSimulation).
		SetListConstraint(data.ListConstraint).
		SetBooleanConstraint(data.BooleanConstraint).
		SetOrganizationID(data.OrganizationID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create org policy constraint history: %w", err)
	}

	return nil
}

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPOrgPolicyConstraint, new *ConstraintData, diff *ConstraintDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPOrgPolicyConstraint.Query().
		Where(
			bronzehistorygcporgpolicyconstraint.ResourceID(old.ID),
			bronzehistorygcporgpolicyconstraint.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current org policy constraint history: %w", err)
	}

	if diff.IsChanged {
		err = tx.BronzeHistoryGCPOrgPolicyConstraint.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current org policy constraint history: %w", err)
		}

		_, err := tx.BronzeHistoryGCPOrgPolicyConstraint.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetDisplayName(new.DisplayName).
			SetDescription(new.Description).
			SetConstraintDefault(new.ConstraintDefault).
			SetSupportsDryRun(new.SupportsDryRun).
			SetSupportsSimulation(new.SupportsSimulation).
			SetListConstraint(new.ListConstraint).
			SetBooleanConstraint(new.BooleanConstraint).
			SetOrganizationID(new.OrganizationID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new org policy constraint history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPOrgPolicyConstraint.Query().
		Where(
			bronzehistorygcporgpolicyconstraint.ResourceID(resourceID),
			bronzehistorygcporgpolicyconstraint.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current org policy constraint history: %w", err)
	}

	err = tx.BronzeHistoryGCPOrgPolicyConstraint.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close org policy constraint history: %w", err)
	}

	return nil
}
