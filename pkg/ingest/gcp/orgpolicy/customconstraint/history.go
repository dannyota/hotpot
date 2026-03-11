package customconstraint

import (
	"context"
	"fmt"
	"time"

	entorgpolicy "danny.vn/hotpot/pkg/storage/ent/gcp/orgpolicy"
	"danny.vn/hotpot/pkg/storage/ent/gcp/orgpolicy/bronzehistorygcporgpolicycustomconstraint"
)

type HistoryService struct {
	entClient *entorgpolicy.Client
}

func NewHistoryService(entClient *entorgpolicy.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *entorgpolicy.Tx, data *CustomConstraintData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPOrgPolicyCustomConstraint.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetResourceTypes(data.ResourceTypes).
		SetMethodTypes(data.MethodTypes).
		SetCondition(data.Condition).
		SetActionType(data.ActionType).
		SetDisplayName(data.DisplayName).
		SetDescription(data.Description).
		SetNillableUpdateTime(data.UpdateTime).
		SetOrganizationID(data.OrganizationID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create org policy custom constraint history: %w", err)
	}

	return nil
}

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entorgpolicy.Tx, old *entorgpolicy.BronzeGCPOrgPolicyCustomConstraint, new *CustomConstraintData, diff *CustomConstraintDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPOrgPolicyCustomConstraint.Query().
		Where(
			bronzehistorygcporgpolicycustomconstraint.ResourceID(old.ID),
			bronzehistorygcporgpolicycustomconstraint.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current org policy custom constraint history: %w", err)
	}

	if diff.IsChanged {
		err = tx.BronzeHistoryGCPOrgPolicyCustomConstraint.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current org policy custom constraint history: %w", err)
		}

		_, err := tx.BronzeHistoryGCPOrgPolicyCustomConstraint.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetResourceTypes(new.ResourceTypes).
			SetMethodTypes(new.MethodTypes).
			SetCondition(new.Condition).
			SetActionType(new.ActionType).
			SetDisplayName(new.DisplayName).
			SetDescription(new.Description).
			SetNillableUpdateTime(new.UpdateTime).
			SetOrganizationID(new.OrganizationID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new org policy custom constraint history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) CloseHistory(ctx context.Context, tx *entorgpolicy.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPOrgPolicyCustomConstraint.Query().
		Where(
			bronzehistorygcporgpolicycustomconstraint.ResourceID(resourceID),
			bronzehistorygcporgpolicycustomconstraint.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entorgpolicy.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current org policy custom constraint history: %w", err)
	}

	err = tx.BronzeHistoryGCPOrgPolicyCustomConstraint.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close org policy custom constraint history: %w", err)
	}

	return nil
}
