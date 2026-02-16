package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigtablecluster"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigtableinstance"
)

// Service handles Bigtable instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Bigtable instance ingestion service.
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
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Bigtable instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch instances from GCP
	instances, err := s.client.ListInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	// Convert to data structs
	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	// Save to database
	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save instances: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		InstanceCount:  len(instanceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInstances saves Bigtable instances to the database with history tracking.
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
		existing, err := tx.BronzeGCPBigtableInstance.Query().
			Where(bronzegcpbigtableinstance.ID(instanceData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing instance %s: %w", instanceData.ID, err)
		}

		// Compute diff
		diff := DiffInstanceData(existing, instanceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPBigtableInstance.UpdateOneID(instanceData.ID).
				SetCollectedAt(instanceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for instance %s: %w", instanceData.ID, err)
			}
			continue
		}

		// Create or update instance
		if existing == nil {
			create := tx.BronzeGCPBigtableInstance.Create().
				SetID(instanceData.ID).
				SetDisplayName(instanceData.DisplayName).
				SetState(instanceData.State).
				SetInstanceType(instanceData.InstanceType).
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt).
				SetFirstCollectedAt(instanceData.CollectedAt)

			if instanceData.CreateTime != "" {
				create.SetCreateTime(instanceData.CreateTime)
			}
			if instanceData.SatisfiesPzs != nil {
				create.SetSatisfiesPzs(*instanceData.SatisfiesPzs)
			}
			if instanceData.LabelsJSON != nil {
				create.SetLabelsJSON(instanceData.LabelsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create instance %s: %w", instanceData.ID, err)
			}
		} else {
			update := tx.BronzeGCPBigtableInstance.UpdateOneID(instanceData.ID).
				SetDisplayName(instanceData.DisplayName).
				SetState(instanceData.State).
				SetInstanceType(instanceData.InstanceType).
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt)

			if instanceData.CreateTime != "" {
				update.SetCreateTime(instanceData.CreateTime)
			}
			if instanceData.SatisfiesPzs != nil {
				update.SetSatisfiesPzs(*instanceData.SatisfiesPzs)
			}
			if instanceData.LabelsJSON != nil {
				update.SetLabelsJSON(instanceData.LabelsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update instance %s: %w", instanceData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instanceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for instance %s: %w", instanceData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instanceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for instance %s: %w", instanceData.ID, err)
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

	staleInstances, err := tx.BronzeGCPBigtableInstance.Query().
		Where(
			bronzegcpbigtableinstance.ProjectID(projectID),
			bronzegcpbigtableinstance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, inst := range staleInstances {
		if err := s.history.CloseHistory(ctx, tx, inst.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for instance %s: %w", inst.ID, err)
		}

		// Delete child clusters first
		_, err := tx.BronzeGCPBigtableCluster.Delete().
			Where(bronzegcpbigtablecluster.HasInstanceWith(bronzegcpbigtableinstance.ID(inst.ID))).
			Exec(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete clusters for instance %s: %w", inst.ID, err)
		}

		if err := tx.BronzeGCPBigtableInstance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
