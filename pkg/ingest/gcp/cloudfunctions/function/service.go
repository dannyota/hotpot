package function

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcloudfunctionsfunction"
)

// Service handles Cloud Function ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Cloud Function ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for function ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of function ingestion.
type IngestResult struct {
	ProjectID      string
	FunctionCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Cloud Functions from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	funcs, err := s.client.ListFunctions(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	funcDataList := make([]*FunctionData, 0, len(funcs))
	for _, f := range funcs {
		data, err := ConvertFunction(f, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert function: %w", err)
		}
		funcDataList = append(funcDataList, data)
	}

	if err := s.saveFunctions(ctx, funcDataList); err != nil {
		return nil, fmt.Errorf("failed to save functions: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		FunctionCount:  len(funcDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveFunctions saves Cloud Functions to the database with history tracking.
func (s *Service) saveFunctions(ctx context.Context, functions []*FunctionData) error {
	if len(functions) == 0 {
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

	for _, funcData := range functions {
		// Load existing function
		existing, err := tx.BronzeGCPCloudFunctionsFunction.Query().
			Where(bronzegcpcloudfunctionsfunction.ID(funcData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing function %s: %w", funcData.ID, err)
		}

		// Compute diff
		diff := DiffFunctionData(existing, funcData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPCloudFunctionsFunction.UpdateOneID(funcData.ID).
				SetCollectedAt(funcData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for function %s: %w", funcData.ID, err)
			}
			continue
		}

		// Create or update function
		if existing == nil {
			create := tx.BronzeGCPCloudFunctionsFunction.Create().
				SetID(funcData.ID).
				SetName(funcData.Name).
				SetEnvironment(funcData.Environment).
				SetState(funcData.State).
				SetSatisfiesPzs(funcData.SatisfiesPzs).
				SetProjectID(funcData.ProjectID).
				SetCollectedAt(funcData.CollectedAt).
				SetFirstCollectedAt(funcData.CollectedAt)

			if funcData.Description != "" {
				create.SetDescription(funcData.Description)
			}
			if funcData.UpdateTime != "" {
				create.SetUpdateTime(funcData.UpdateTime)
			}
			if funcData.CreateTime != "" {
				create.SetCreateTime(funcData.CreateTime)
			}
			if funcData.KmsKeyName != "" {
				create.SetKmsKeyName(funcData.KmsKeyName)
			}
			if funcData.URL != "" {
				create.SetURL(funcData.URL)
			}
			if funcData.Location != "" {
				create.SetLocation(funcData.Location)
			}
			if funcData.BuildConfigJSON != nil {
				create.SetBuildConfigJSON(funcData.BuildConfigJSON)
			}
			if funcData.ServiceConfigJSON != nil {
				create.SetServiceConfigJSON(funcData.ServiceConfigJSON)
			}
			if funcData.EventTriggerJSON != nil {
				create.SetEventTriggerJSON(funcData.EventTriggerJSON)
			}
			if funcData.StateMessagesJSON != nil {
				create.SetStateMessagesJSON(funcData.StateMessagesJSON)
			}
			if funcData.LabelsJSON != nil {
				create.SetLabelsJSON(funcData.LabelsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create function %s: %w", funcData.ID, err)
			}
		} else {
			update := tx.BronzeGCPCloudFunctionsFunction.UpdateOneID(funcData.ID).
				SetName(funcData.Name).
				SetEnvironment(funcData.Environment).
				SetState(funcData.State).
				SetSatisfiesPzs(funcData.SatisfiesPzs).
				SetProjectID(funcData.ProjectID).
				SetCollectedAt(funcData.CollectedAt)

			if funcData.Description != "" {
				update.SetDescription(funcData.Description)
			}
			if funcData.UpdateTime != "" {
				update.SetUpdateTime(funcData.UpdateTime)
			}
			if funcData.CreateTime != "" {
				update.SetCreateTime(funcData.CreateTime)
			}
			if funcData.KmsKeyName != "" {
				update.SetKmsKeyName(funcData.KmsKeyName)
			}
			if funcData.URL != "" {
				update.SetURL(funcData.URL)
			}
			if funcData.Location != "" {
				update.SetLocation(funcData.Location)
			}
			if funcData.BuildConfigJSON != nil {
				update.SetBuildConfigJSON(funcData.BuildConfigJSON)
			}
			if funcData.ServiceConfigJSON != nil {
				update.SetServiceConfigJSON(funcData.ServiceConfigJSON)
			}
			if funcData.EventTriggerJSON != nil {
				update.SetEventTriggerJSON(funcData.EventTriggerJSON)
			}
			if funcData.StateMessagesJSON != nil {
				update.SetStateMessagesJSON(funcData.StateMessagesJSON)
			}
			if funcData.LabelsJSON != nil {
				update.SetLabelsJSON(funcData.LabelsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update function %s: %w", funcData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, funcData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for function %s: %w", funcData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, funcData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for function %s: %w", funcData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleFunctions removes functions that were not collected in the latest run.
func (s *Service) DeleteStaleFunctions(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleFunctions, err := tx.BronzeGCPCloudFunctionsFunction.Query().
		Where(
			bronzegcpcloudfunctionsfunction.ProjectID(projectID),
			bronzegcpcloudfunctionsfunction.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, fn := range staleFunctions {
		if err := s.history.CloseHistory(ctx, tx, fn.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for function %s: %w", fn.ID, err)
		}

		if err := tx.BronzeGCPCloudFunctionsFunction.DeleteOne(fn).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete function %s: %w", fn.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
