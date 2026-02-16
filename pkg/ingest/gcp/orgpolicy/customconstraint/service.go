package customconstraint

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcporgpolicycustomconstraint"
)

type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

type IngestResult struct {
	CustomConstraintCount int
	CollectedAt           time.Time
	DurationMillis        int64
}

func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawCustomConstraints, err := s.client.ListCustomConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom constraints: %w", err)
	}

	customConstraintDataList := make([]*CustomConstraintData, 0, len(rawCustomConstraints))
	for _, raw := range rawCustomConstraints {
		data := ConvertCustomConstraint(raw.OrgName, raw.CustomConstraint, collectedAt)
		customConstraintDataList = append(customConstraintDataList, data)
	}

	if err := s.saveCustomConstraints(ctx, customConstraintDataList); err != nil {
		return nil, fmt.Errorf("failed to save custom constraints: %w", err)
	}

	return &IngestResult{
		CustomConstraintCount: len(customConstraintDataList),
		CollectedAt:           collectedAt,
		DurationMillis:        time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveCustomConstraints(ctx context.Context, customConstraints []*CustomConstraintData) error {
	if len(customConstraints) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, ccData := range customConstraints {
		existing, err := tx.BronzeGCPOrgPolicyCustomConstraint.Query().
			Where(bronzegcporgpolicycustomconstraint.ID(ccData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing custom constraint %s: %w", ccData.ID, err)
		}

		diff := DiffCustomConstraintData(existing, ccData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPOrgPolicyCustomConstraint.UpdateOneID(ccData.ID).
				SetCollectedAt(ccData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for custom constraint %s: %w", ccData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPOrgPolicyCustomConstraint.Create().
				SetID(ccData.ID).
				SetOrganizationID(ccData.OrganizationID).
				SetCollectedAt(ccData.CollectedAt).
				SetFirstCollectedAt(ccData.CollectedAt).
				SetActionType(ccData.ActionType)

			if ccData.DisplayName != "" {
				create.SetDisplayName(ccData.DisplayName)
			}
			if ccData.Description != "" {
				create.SetDescription(ccData.Description)
			}
			if ccData.Condition != "" {
				create.SetCondition(ccData.Condition)
			}
			if ccData.ResourceTypes != nil {
				create.SetResourceTypes(ccData.ResourceTypes)
			}
			if ccData.MethodTypes != nil {
				create.SetMethodTypes(ccData.MethodTypes)
			}
			if ccData.UpdateTime != nil {
				create.SetUpdateTime(*ccData.UpdateTime)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create custom constraint %s: %w", ccData.ID, err)
			}
		} else {
			update := tx.BronzeGCPOrgPolicyCustomConstraint.UpdateOneID(ccData.ID).
				SetOrganizationID(ccData.OrganizationID).
				SetCollectedAt(ccData.CollectedAt).
				SetActionType(ccData.ActionType)

			if ccData.DisplayName != "" {
				update.SetDisplayName(ccData.DisplayName)
			}
			if ccData.Description != "" {
				update.SetDescription(ccData.Description)
			}
			if ccData.Condition != "" {
				update.SetCondition(ccData.Condition)
			}
			if ccData.ResourceTypes != nil {
				update.SetResourceTypes(ccData.ResourceTypes)
			}
			if ccData.MethodTypes != nil {
				update.SetMethodTypes(ccData.MethodTypes)
			}
			if ccData.UpdateTime != nil {
				update.SetUpdateTime(*ccData.UpdateTime)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update custom constraint %s: %w", ccData.ID, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, ccData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for custom constraint %s: %w", ccData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, ccData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for custom constraint %s: %w", ccData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) DeleteStaleCustomConstraints(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleCustomConstraints, err := tx.BronzeGCPOrgPolicyCustomConstraint.Query().
		Where(bronzegcporgpolicycustomconstraint.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, cc := range staleCustomConstraints {
		if err := s.history.CloseHistory(ctx, tx, cc.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for custom constraint %s: %w", cc.ID, err)
		}

		if err := tx.BronzeGCPOrgPolicyCustomConstraint.DeleteOne(cc).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete custom constraint %s: %w", cc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
