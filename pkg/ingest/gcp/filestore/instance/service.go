package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpfilestoreinstance"
)

// Service handles Filestore instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Filestore instance ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for Filestore instance ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of Filestore instance ingestion.
type IngestResult struct {
	ProjectID      string
	InstanceCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Filestore instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	instances, err := s.client.ListInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Filestore instances: %w", err)
	}

	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Filestore instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save Filestore instances: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		InstanceCount:  len(instanceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInstances saves Filestore instances to the database with history tracking.
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

	for _, instData := range instances {
		existing, err := tx.BronzeGCPFilestoreInstance.Query().
			Where(bronzegcpfilestoreinstance.ID(instData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Filestore instance %s: %w", instData.ID, err)
		}

		diff := DiffInstanceData(existing, instData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPFilestoreInstance.UpdateOneID(instData.ID).
				SetCollectedAt(instData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Filestore instance %s: %w", instData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPFilestoreInstance.Create().
				SetID(instData.ID).
				SetName(instData.Name).
				SetDescription(instData.Description).
				SetState(instData.State).
				SetStatusMessage(instData.StatusMessage).
				SetCreateTime(instData.CreateTime).
				SetTier(instData.Tier).
				SetEtag(instData.Etag).
				SetSatisfiesPzs(instData.SatisfiesPzs).
				SetSatisfiesPzi(instData.SatisfiesPzi).
				SetKmsKeyName(instData.KmsKeyName).
				SetMaxCapacityGB(instData.MaxCapacityGB).
				SetProtocol(instData.Protocol).
				SetProjectID(instData.ProjectID).
				SetLocation(instData.Location).
				SetCollectedAt(instData.CollectedAt).
				SetFirstCollectedAt(instData.CollectedAt)

			if instData.LabelsJSON != nil {
				create.SetLabelsJSON(instData.LabelsJSON)
			}
			if instData.FileSharesJSON != nil {
				create.SetFileSharesJSON(instData.FileSharesJSON)
			}
			if instData.NetworksJSON != nil {
				create.SetNetworksJSON(instData.NetworksJSON)
			}
			if instData.SuspensionReasonsJSON != nil {
				create.SetSuspensionReasonsJSON(instData.SuspensionReasonsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Filestore instance %s: %w", instData.ID, err)
			}
		} else {
			update := tx.BronzeGCPFilestoreInstance.UpdateOneID(instData.ID).
				SetName(instData.Name).
				SetDescription(instData.Description).
				SetState(instData.State).
				SetStatusMessage(instData.StatusMessage).
				SetCreateTime(instData.CreateTime).
				SetTier(instData.Tier).
				SetEtag(instData.Etag).
				SetSatisfiesPzs(instData.SatisfiesPzs).
				SetSatisfiesPzi(instData.SatisfiesPzi).
				SetKmsKeyName(instData.KmsKeyName).
				SetMaxCapacityGB(instData.MaxCapacityGB).
				SetProtocol(instData.Protocol).
				SetProjectID(instData.ProjectID).
				SetLocation(instData.Location).
				SetCollectedAt(instData.CollectedAt)

			if instData.LabelsJSON != nil {
				update.SetLabelsJSON(instData.LabelsJSON)
			}
			if instData.FileSharesJSON != nil {
				update.SetFileSharesJSON(instData.FileSharesJSON)
			}
			if instData.NetworksJSON != nil {
				update.SetNetworksJSON(instData.NetworksJSON)
			}
			if instData.SuspensionReasonsJSON != nil {
				update.SetSuspensionReasonsJSON(instData.SuspensionReasonsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Filestore instance %s: %w", instData.ID, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Filestore instance %s: %w", instData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Filestore instance %s: %w", instData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleInstances removes Filestore instances that were not collected in the latest run.
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

	staleInstances, err := tx.BronzeGCPFilestoreInstance.Query().
		Where(
			bronzegcpfilestoreinstance.ProjectID(projectID),
			bronzegcpfilestoreinstance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, inst := range staleInstances {
		if err := s.history.CloseHistory(ctx, tx, inst.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Filestore instance %s: %w", inst.ID, err)
		}

		if err := tx.BronzeGCPFilestoreInstance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Filestore instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
