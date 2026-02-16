package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpredisinstance"
)

// Service handles GCP Memorystore Redis instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Redis instance ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for Redis instance ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of Redis instance ingestion.
type IngestResult struct {
	ProjectID      string
	InstanceCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Redis instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	instances, err := s.client.ListInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Redis instances: %w", err)
	}

	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Redis instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save Redis instances: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		InstanceCount:  len(instanceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

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
		existing, err := tx.BronzeGCPRedisInstance.Query().
			Where(bronzegcpredisinstance.ID(instData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Redis instance %s: %w", instData.ID, err)
		}

		diff := DiffInstanceData(existing, instData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPRedisInstance.UpdateOneID(instData.ID).
				SetCollectedAt(instData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Redis instance %s: %w", instData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPRedisInstance.Create().
				SetID(instData.ID).
				SetName(instData.Name).
				SetDisplayName(instData.DisplayName).
				SetLocationID(instData.LocationID).
				SetAlternativeLocationID(instData.AlternativeLocationID).
				SetRedisVersion(instData.RedisVersion).
				SetReservedIPRange(instData.ReservedIPRange).
				SetSecondaryIPRange(instData.SecondaryIPRange).
				SetHost(instData.Host).
				SetPort(instData.Port).
				SetCurrentLocationID(instData.CurrentLocationID).
				SetCreateTime(instData.CreateTime).
				SetState(instData.State).
				SetStatusMessage(instData.StatusMessage).
				SetTier(instData.Tier).
				SetMemorySizeGB(instData.MemorySizeGB).
				SetAuthorizedNetwork(instData.AuthorizedNetwork).
				SetPersistenceIamIdentity(instData.PersistenceIamIdentity).
				SetConnectMode(instData.ConnectMode).
				SetAuthEnabled(instData.AuthEnabled).
				SetTransitEncryptionMode(instData.TransitEncryptionMode).
				SetReplicaCount(instData.ReplicaCount).
				SetReadEndpoint(instData.ReadEndpoint).
				SetReadEndpointPort(instData.ReadEndpointPort).
				SetReadReplicasMode(instData.ReadReplicasMode).
				SetCustomerManagedKey(instData.CustomerManagedKey).
				SetMaintenanceVersion(instData.MaintenanceVersion).
				SetProjectID(instData.ProjectID).
				SetCollectedAt(instData.CollectedAt).
				SetFirstCollectedAt(instData.CollectedAt)

			if instData.LabelsJSON != nil {
				create.SetLabelsJSON(instData.LabelsJSON)
			}
			if instData.RedisConfigsJSON != nil {
				create.SetRedisConfigsJSON(instData.RedisConfigsJSON)
			}
			if instData.ServerCaCertsJSON != nil {
				create.SetServerCaCertsJSON(instData.ServerCaCertsJSON)
			}
			if instData.MaintenancePolicyJSON != nil {
				create.SetMaintenancePolicyJSON(instData.MaintenancePolicyJSON)
			}
			if instData.MaintenanceScheduleJSON != nil {
				create.SetMaintenanceScheduleJSON(instData.MaintenanceScheduleJSON)
			}
			if instData.NodesJSON != nil {
				create.SetNodesJSON(instData.NodesJSON)
			}
			if instData.PersistenceConfigJSON != nil {
				create.SetPersistenceConfigJSON(instData.PersistenceConfigJSON)
			}
			if instData.SuspensionReasonsJSON != nil {
				create.SetSuspensionReasonsJSON(instData.SuspensionReasonsJSON)
			}
			if instData.AvailableMaintenanceVersionsJSON != nil {
				create.SetAvailableMaintenanceVersionsJSON(instData.AvailableMaintenanceVersionsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Redis instance %s: %w", instData.ID, err)
			}
		} else {
			update := tx.BronzeGCPRedisInstance.UpdateOneID(instData.ID).
				SetName(instData.Name).
				SetDisplayName(instData.DisplayName).
				SetLocationID(instData.LocationID).
				SetAlternativeLocationID(instData.AlternativeLocationID).
				SetRedisVersion(instData.RedisVersion).
				SetReservedIPRange(instData.ReservedIPRange).
				SetSecondaryIPRange(instData.SecondaryIPRange).
				SetHost(instData.Host).
				SetPort(instData.Port).
				SetCurrentLocationID(instData.CurrentLocationID).
				SetCreateTime(instData.CreateTime).
				SetState(instData.State).
				SetStatusMessage(instData.StatusMessage).
				SetTier(instData.Tier).
				SetMemorySizeGB(instData.MemorySizeGB).
				SetAuthorizedNetwork(instData.AuthorizedNetwork).
				SetPersistenceIamIdentity(instData.PersistenceIamIdentity).
				SetConnectMode(instData.ConnectMode).
				SetAuthEnabled(instData.AuthEnabled).
				SetTransitEncryptionMode(instData.TransitEncryptionMode).
				SetReplicaCount(instData.ReplicaCount).
				SetReadEndpoint(instData.ReadEndpoint).
				SetReadEndpointPort(instData.ReadEndpointPort).
				SetReadReplicasMode(instData.ReadReplicasMode).
				SetCustomerManagedKey(instData.CustomerManagedKey).
				SetMaintenanceVersion(instData.MaintenanceVersion).
				SetProjectID(instData.ProjectID).
				SetCollectedAt(instData.CollectedAt)

			if instData.LabelsJSON != nil {
				update.SetLabelsJSON(instData.LabelsJSON)
			}
			if instData.RedisConfigsJSON != nil {
				update.SetRedisConfigsJSON(instData.RedisConfigsJSON)
			}
			if instData.ServerCaCertsJSON != nil {
				update.SetServerCaCertsJSON(instData.ServerCaCertsJSON)
			}
			if instData.MaintenancePolicyJSON != nil {
				update.SetMaintenancePolicyJSON(instData.MaintenancePolicyJSON)
			}
			if instData.MaintenanceScheduleJSON != nil {
				update.SetMaintenanceScheduleJSON(instData.MaintenanceScheduleJSON)
			}
			if instData.NodesJSON != nil {
				update.SetNodesJSON(instData.NodesJSON)
			}
			if instData.PersistenceConfigJSON != nil {
				update.SetPersistenceConfigJSON(instData.PersistenceConfigJSON)
			}
			if instData.SuspensionReasonsJSON != nil {
				update.SetSuspensionReasonsJSON(instData.SuspensionReasonsJSON)
			}
			if instData.AvailableMaintenanceVersionsJSON != nil {
				update.SetAvailableMaintenanceVersionsJSON(instData.AvailableMaintenanceVersionsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Redis instance %s: %w", instData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Redis instance %s: %w", instData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Redis instance %s: %w", instData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleInstances removes Redis instances that were not collected in the latest run.
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

	staleInstances, err := tx.BronzeGCPRedisInstance.Query().
		Where(
			bronzegcpredisinstance.ProjectID(projectID),
			bronzegcpredisinstance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, inst := range staleInstances {
		if err := s.history.CloseHistory(ctx, tx, inst.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Redis instance %s: %w", inst.ID, err)
		}

		if err := tx.BronzeGCPRedisInstance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Redis instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
