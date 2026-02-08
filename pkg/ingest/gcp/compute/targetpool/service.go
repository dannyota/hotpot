package targetpool

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputetargetpool"
)

// Service handles GCP Compute target pool ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new target pool ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for target pool ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target pool ingestion.
type IngestResult struct {
	ProjectID       string
	TargetPoolCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches target pools from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch target pools from GCP
	pools, err := s.client.ListTargetPools(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target pools: %w", err)
	}

	// Convert to data structs
	dataList := make([]*TargetPoolData, 0, len(pools))
	for _, tp := range pools {
		data, err := ConvertTargetPool(tp, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert target pool: %w", err)
		}
		dataList = append(dataList, data)
	}

	// Save to database
	if err := s.saveTargetPools(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save target pools: %w", err)
	}

	return &IngestResult{
		ProjectID:       params.ProjectID,
		TargetPoolCount: len(dataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTargetPools saves target pools to the database with history tracking.
func (s *Service) saveTargetPools(ctx context.Context, pools []*TargetPoolData) error {
	if len(pools) == 0 {
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

	for _, poolData := range pools {
		// Load existing target pool
		existing, err := tx.BronzeGCPComputeTargetPool.Query().
			Where(bronzegcpcomputetargetpool.ID(poolData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing target pool %s: %w", poolData.Name, err)
		}

		// Compute diff
		diff := DiffTargetPoolData(existing, poolData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeTargetPool.UpdateOneID(poolData.ID).
				SetCollectedAt(poolData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for target pool %s: %w", poolData.Name, err)
			}
			continue
		}

		// Create or update target pool
		if existing == nil {
			// Create new target pool
			create := tx.BronzeGCPComputeTargetPool.Create().
				SetID(poolData.ID).
				SetName(poolData.Name).
				SetProjectID(poolData.ProjectID).
				SetCollectedAt(poolData.CollectedAt).
				SetFirstCollectedAt(poolData.CollectedAt)

			if poolData.Description != "" {
				create.SetDescription(poolData.Description)
			}
			if poolData.CreationTimestamp != "" {
				create.SetCreationTimestamp(poolData.CreationTimestamp)
			}
			if poolData.SelfLink != "" {
				create.SetSelfLink(poolData.SelfLink)
			}
			if poolData.SessionAffinity != "" {
				create.SetSessionAffinity(poolData.SessionAffinity)
			}
			if poolData.BackupPool != "" {
				create.SetBackupPool(poolData.BackupPool)
			}
			if poolData.FailoverRatio != 0 {
				create.SetFailoverRatio(poolData.FailoverRatio)
			}
			if poolData.SecurityPolicy != "" {
				create.SetSecurityPolicy(poolData.SecurityPolicy)
			}
			if poolData.Region != "" {
				create.SetRegion(poolData.Region)
			}
			if poolData.HealthChecksJSON != nil {
				create.SetHealthChecksJSON(poolData.HealthChecksJSON)
			}
			if poolData.InstancesJSON != nil {
				create.SetInstancesJSON(poolData.InstancesJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create target pool %s: %w", poolData.Name, err)
			}
		} else {
			// Update existing target pool
			update := tx.BronzeGCPComputeTargetPool.UpdateOneID(poolData.ID).
				SetName(poolData.Name).
				SetProjectID(poolData.ProjectID).
				SetCollectedAt(poolData.CollectedAt)

			if poolData.Description != "" {
				update.SetDescription(poolData.Description)
			}
			if poolData.CreationTimestamp != "" {
				update.SetCreationTimestamp(poolData.CreationTimestamp)
			}
			if poolData.SelfLink != "" {
				update.SetSelfLink(poolData.SelfLink)
			}
			if poolData.SessionAffinity != "" {
				update.SetSessionAffinity(poolData.SessionAffinity)
			}
			if poolData.BackupPool != "" {
				update.SetBackupPool(poolData.BackupPool)
			}
			if poolData.FailoverRatio != 0 {
				update.SetFailoverRatio(poolData.FailoverRatio)
			}
			if poolData.SecurityPolicy != "" {
				update.SetSecurityPolicy(poolData.SecurityPolicy)
			}
			if poolData.Region != "" {
				update.SetRegion(poolData.Region)
			}
			if poolData.HealthChecksJSON != nil {
				update.SetHealthChecksJSON(poolData.HealthChecksJSON)
			}
			if poolData.InstancesJSON != nil {
				update.SetInstancesJSON(poolData.InstancesJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update target pool %s: %w", poolData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, poolData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for target pool %s: %w", poolData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, poolData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for target pool %s: %w", poolData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTargetPools removes target pools not collected in the latest run.
// Also closes history records for deleted target pools.
func (s *Service) DeleteStaleTargetPools(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale target pools
	stale, err := tx.BronzeGCPComputeTargetPool.Query().
		Where(
			bronzegcpcomputetargetpool.ProjectID(projectID),
			bronzegcpcomputetargetpool.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale target pool
	for _, pool := range stale {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, pool.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for target pool %s: %w", pool.ID, err)
		}

		// Delete target pool
		if err := tx.BronzeGCPComputeTargetPool.DeleteOne(pool).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete target pool %s: %w", pool.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
