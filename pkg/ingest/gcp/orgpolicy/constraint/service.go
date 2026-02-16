package constraint

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcporgpolicyconstraint"
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
	ConstraintCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawConstraints, err := s.client.ListConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list constraints: %w", err)
	}

	constraintDataList := make([]*ConstraintData, 0, len(rawConstraints))
	for _, raw := range rawConstraints {
		data := ConvertConstraint(raw.OrgName, raw.Constraint, collectedAt)
		constraintDataList = append(constraintDataList, data)
	}

	if err := s.saveConstraints(ctx, constraintDataList); err != nil {
		return nil, fmt.Errorf("failed to save constraints: %w", err)
	}

	return &IngestResult{
		ConstraintCount: len(constraintDataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveConstraints(ctx context.Context, constraints []*ConstraintData) error {
	if len(constraints) == 0 {
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

	for _, constraintData := range constraints {
		existing, err := tx.BronzeGCPOrgPolicyConstraint.Query().
			Where(bronzegcporgpolicyconstraint.ID(constraintData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing constraint %s: %w", constraintData.ID, err)
		}

		diff := DiffConstraintData(existing, constraintData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPOrgPolicyConstraint.UpdateOneID(constraintData.ID).
				SetCollectedAt(constraintData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for constraint %s: %w", constraintData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPOrgPolicyConstraint.Create().
				SetID(constraintData.ID).
				SetOrganizationID(constraintData.OrganizationID).
				SetCollectedAt(constraintData.CollectedAt).
				SetFirstCollectedAt(constraintData.CollectedAt).
				SetConstraintDefault(constraintData.ConstraintDefault).
				SetSupportsDryRun(constraintData.SupportsDryRun).
				SetSupportsSimulation(constraintData.SupportsSimulation)

			if constraintData.DisplayName != "" {
				create.SetDisplayName(constraintData.DisplayName)
			}
			if constraintData.Description != "" {
				create.SetDescription(constraintData.Description)
			}
			if constraintData.ListConstraint != nil {
				create.SetListConstraint(constraintData.ListConstraint)
			}
			if constraintData.BooleanConstraint != nil {
				create.SetBooleanConstraint(constraintData.BooleanConstraint)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create constraint %s: %w", constraintData.ID, err)
			}
		} else {
			update := tx.BronzeGCPOrgPolicyConstraint.UpdateOneID(constraintData.ID).
				SetOrganizationID(constraintData.OrganizationID).
				SetCollectedAt(constraintData.CollectedAt).
				SetConstraintDefault(constraintData.ConstraintDefault).
				SetSupportsDryRun(constraintData.SupportsDryRun).
				SetSupportsSimulation(constraintData.SupportsSimulation)

			if constraintData.DisplayName != "" {
				update.SetDisplayName(constraintData.DisplayName)
			}
			if constraintData.Description != "" {
				update.SetDescription(constraintData.Description)
			}
			if constraintData.ListConstraint != nil {
				update.SetListConstraint(constraintData.ListConstraint)
			}
			if constraintData.BooleanConstraint != nil {
				update.SetBooleanConstraint(constraintData.BooleanConstraint)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update constraint %s: %w", constraintData.ID, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, constraintData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for constraint %s: %w", constraintData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, constraintData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for constraint %s: %w", constraintData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) DeleteStaleConstraints(ctx context.Context, collectedAt time.Time) error {
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

	staleConstraints, err := tx.BronzeGCPOrgPolicyConstraint.Query().
		Where(bronzegcporgpolicyconstraint.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, con := range staleConstraints {
		if err := s.history.CloseHistory(ctx, tx, con.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for constraint %s: %w", con.ID, err)
		}

		if err := tx.BronzeGCPOrgPolicyConstraint.DeleteOne(con).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete constraint %s: %w", con.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
