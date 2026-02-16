package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpredisinstance"
)

// HistoryService manages Redis instance history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Redis instance.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InstanceData, now time.Time) error {
	create := tx.BronzeHistoryGCPRedisInstance.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDisplayName(data.DisplayName).
		SetLocationID(data.LocationID).
		SetAlternativeLocationID(data.AlternativeLocationID).
		SetRedisVersion(data.RedisVersion).
		SetReservedIPRange(data.ReservedIPRange).
		SetSecondaryIPRange(data.SecondaryIPRange).
		SetHost(data.Host).
		SetPort(data.Port).
		SetCurrentLocationID(data.CurrentLocationID).
		SetCreateTime(data.CreateTime).
		SetState(data.State).
		SetStatusMessage(data.StatusMessage).
		SetTier(data.Tier).
		SetMemorySizeGB(data.MemorySizeGB).
		SetAuthorizedNetwork(data.AuthorizedNetwork).
		SetPersistenceIamIdentity(data.PersistenceIamIdentity).
		SetConnectMode(data.ConnectMode).
		SetAuthEnabled(data.AuthEnabled).
		SetTransitEncryptionMode(data.TransitEncryptionMode).
		SetReplicaCount(data.ReplicaCount).
		SetReadEndpoint(data.ReadEndpoint).
		SetReadEndpointPort(data.ReadEndpointPort).
		SetReadReplicasMode(data.ReadReplicasMode).
		SetCustomerManagedKey(data.CustomerManagedKey).
		SetMaintenanceVersion(data.MaintenanceVersion).
		SetProjectID(data.ProjectID)

	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.RedisConfigsJSON != nil {
		create.SetRedisConfigsJSON(data.RedisConfigsJSON)
	}
	if data.ServerCaCertsJSON != nil {
		create.SetServerCaCertsJSON(data.ServerCaCertsJSON)
	}
	if data.MaintenancePolicyJSON != nil {
		create.SetMaintenancePolicyJSON(data.MaintenancePolicyJSON)
	}
	if data.MaintenanceScheduleJSON != nil {
		create.SetMaintenanceScheduleJSON(data.MaintenanceScheduleJSON)
	}
	if data.NodesJSON != nil {
		create.SetNodesJSON(data.NodesJSON)
	}
	if data.PersistenceConfigJSON != nil {
		create.SetPersistenceConfigJSON(data.PersistenceConfigJSON)
	}
	if data.SuspensionReasonsJSON != nil {
		create.SetSuspensionReasonsJSON(data.SuspensionReasonsJSON)
	}
	if data.AvailableMaintenanceVersionsJSON != nil {
		create.SetAvailableMaintenanceVersionsJSON(data.AvailableMaintenanceVersionsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Redis instance history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Redis instance.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPRedisInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPRedisInstance.Query().
		Where(
			bronzehistorygcpredisinstance.ResourceID(old.ID),
			bronzehistorygcpredisinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Redis instance history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPRedisInstance.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Redis instance history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPRedisInstance.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDisplayName(new.DisplayName).
			SetLocationID(new.LocationID).
			SetAlternativeLocationID(new.AlternativeLocationID).
			SetRedisVersion(new.RedisVersion).
			SetReservedIPRange(new.ReservedIPRange).
			SetSecondaryIPRange(new.SecondaryIPRange).
			SetHost(new.Host).
			SetPort(new.Port).
			SetCurrentLocationID(new.CurrentLocationID).
			SetCreateTime(new.CreateTime).
			SetState(new.State).
			SetStatusMessage(new.StatusMessage).
			SetTier(new.Tier).
			SetMemorySizeGB(new.MemorySizeGB).
			SetAuthorizedNetwork(new.AuthorizedNetwork).
			SetPersistenceIamIdentity(new.PersistenceIamIdentity).
			SetConnectMode(new.ConnectMode).
			SetAuthEnabled(new.AuthEnabled).
			SetTransitEncryptionMode(new.TransitEncryptionMode).
			SetReplicaCount(new.ReplicaCount).
			SetReadEndpoint(new.ReadEndpoint).
			SetReadEndpointPort(new.ReadEndpointPort).
			SetReadReplicasMode(new.ReadReplicasMode).
			SetCustomerManagedKey(new.CustomerManagedKey).
			SetMaintenanceVersion(new.MaintenanceVersion).
			SetProjectID(new.ProjectID)

		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.RedisConfigsJSON != nil {
			create.SetRedisConfigsJSON(new.RedisConfigsJSON)
		}
		if new.ServerCaCertsJSON != nil {
			create.SetServerCaCertsJSON(new.ServerCaCertsJSON)
		}
		if new.MaintenancePolicyJSON != nil {
			create.SetMaintenancePolicyJSON(new.MaintenancePolicyJSON)
		}
		if new.MaintenanceScheduleJSON != nil {
			create.SetMaintenanceScheduleJSON(new.MaintenanceScheduleJSON)
		}
		if new.NodesJSON != nil {
			create.SetNodesJSON(new.NodesJSON)
		}
		if new.PersistenceConfigJSON != nil {
			create.SetPersistenceConfigJSON(new.PersistenceConfigJSON)
		}
		if new.SuspensionReasonsJSON != nil {
			create.SetSuspensionReasonsJSON(new.SuspensionReasonsJSON)
		}
		if new.AvailableMaintenanceVersionsJSON != nil {
			create.SetAvailableMaintenanceVersionsJSON(new.AvailableMaintenanceVersionsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Redis instance history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Redis instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPRedisInstance.Query().
		Where(
			bronzehistorygcpredisinstance.ResourceID(resourceID),
			bronzehistorygcpredisinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Redis instance history: %w", err)
	}

	err = tx.BronzeHistoryGCPRedisInstance.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Redis instance history: %w", err)
	}

	return nil
}
