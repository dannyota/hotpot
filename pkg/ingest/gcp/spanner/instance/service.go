package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpspannerinstance"
)

// Service handles Spanner instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Spanner instance ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for instance ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of instance ingestion.
type IngestResult struct {
	ProjectID      string
	InstanceCount  int
	InstanceNames  []string
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Spanner instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch instances from GCP
	instances, err := s.client.ListInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Spanner instances: %w", err)
	}

	// Convert to data structs
	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Spanner instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	// Save to database
	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save Spanner instances: %w", err)
	}

	// Collect instance names for downstream use (database listing)
	instanceNames := make([]string, 0, len(instanceDataList))
	for _, data := range instanceDataList {
		instanceNames = append(instanceNames, data.ResourceID)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		InstanceCount:  len(instanceDataList),
		InstanceNames:  instanceNames,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInstances saves Spanner instances to the database with history tracking.
func (s *Service) saveInstances(ctx context.Context, instances []*InstanceData) error {
	if len(instances) == 0 {
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

	for _, instanceData := range instances {
		// Load existing instance
		existing, err := tx.BronzeGCPSpannerInstance.Query().
			Where(bronzegcpspannerinstance.ID(instanceData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Spanner instance %s: %w", instanceData.ResourceID, err)
		}

		// Compute diff
		diff := DiffInstanceData(existing, instanceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPSpannerInstance.UpdateOneID(instanceData.ResourceID).
				SetCollectedAt(instanceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Spanner instance %s: %w", instanceData.ResourceID, err)
			}
			continue
		}

		// Create or update instance
		if existing == nil {
			create := tx.BronzeGCPSpannerInstance.Create().
				SetID(instanceData.ResourceID).
				SetName(instanceData.Name).
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt).
				SetFirstCollectedAt(instanceData.CollectedAt).
				SetNodeCount(instanceData.NodeCount).
				SetProcessingUnits(instanceData.ProcessingUnits).
				SetState(instanceData.State).
				SetEdition(instanceData.Edition).
				SetDefaultBackupScheduleType(instanceData.DefaultBackupScheduleType)

			if instanceData.Config != "" {
				create.SetConfig(instanceData.Config)
			}
			if instanceData.DisplayName != "" {
				create.SetDisplayName(instanceData.DisplayName)
			}
			if instanceData.CreateTime != "" {
				create.SetCreateTime(instanceData.CreateTime)
			}
			if instanceData.UpdateTime != "" {
				create.SetUpdateTime(instanceData.UpdateTime)
			}
			if instanceData.LabelsJSON != nil {
				create.SetLabelsJSON(instanceData.LabelsJSON)
			}
			if instanceData.EndpointUrisJSON != nil {
				create.SetEndpointUrisJSON(instanceData.EndpointUrisJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Spanner instance %s: %w", instanceData.ResourceID, err)
			}
		} else {
			update := tx.BronzeGCPSpannerInstance.UpdateOneID(instanceData.ResourceID).
				SetName(instanceData.Name).
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt).
				SetNodeCount(instanceData.NodeCount).
				SetProcessingUnits(instanceData.ProcessingUnits).
				SetState(instanceData.State).
				SetEdition(instanceData.Edition).
				SetDefaultBackupScheduleType(instanceData.DefaultBackupScheduleType)

			if instanceData.Config != "" {
				update.SetConfig(instanceData.Config)
			}
			if instanceData.DisplayName != "" {
				update.SetDisplayName(instanceData.DisplayName)
			}
			if instanceData.CreateTime != "" {
				update.SetCreateTime(instanceData.CreateTime)
			}
			if instanceData.UpdateTime != "" {
				update.SetUpdateTime(instanceData.UpdateTime)
			}
			if instanceData.LabelsJSON != nil {
				update.SetLabelsJSON(instanceData.LabelsJSON)
			}
			if instanceData.EndpointUrisJSON != nil {
				update.SetEndpointUrisJSON(instanceData.EndpointUrisJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Spanner instance %s: %w", instanceData.ResourceID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instanceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Spanner instance %s: %w", instanceData.ResourceID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instanceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Spanner instance %s: %w", instanceData.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleInstances removes instances that were not collected in the latest run.
func (s *Service) DeleteStaleInstances(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleInstances, err := tx.BronzeGCPSpannerInstance.Query().
		Where(
			bronzegcpspannerinstance.ProjectID(projectID),
			bronzegcpspannerinstance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, inst := range staleInstances {
		if err := s.history.CloseHistory(ctx, tx, inst.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Spanner instance %s: %w", inst.ID, err)
		}

		// Delete instance (CASCADE will handle databases automatically)
		if err := tx.BronzeGCPSpannerInstance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Spanner instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
