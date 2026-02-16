package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpbigtablecluster"
)

// HistoryService manages Bigtable cluster history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Bigtable cluster.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ClusterData, now time.Time) error {
	create := tx.BronzeHistoryGCPBigtableCluster.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetLocation(data.Location).
		SetState(data.State).
		SetServeNodes(data.ServeNodes).
		SetDefaultStorageType(data.DefaultStorageType).
		SetInstanceName(data.InstanceName).
		SetProjectID(data.ProjectID)

	if data.EncryptionConfigJSON != nil {
		create.SetEncryptionConfigJSON(data.EncryptionConfigJSON)
	}
	if data.ClusterConfigJSON != nil {
		create.SetClusterConfigJSON(data.ClusterConfigJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Bigtable cluster history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Bigtable cluster.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPBigtableCluster, new *ClusterData, diff *ClusterDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigtableCluster.Query().
		Where(
			bronzehistorygcpbigtablecluster.ResourceID(old.ID),
			bronzehistorygcpbigtablecluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Bigtable cluster history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPBigtableCluster.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Bigtable cluster history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPBigtableCluster.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetLocation(new.Location).
			SetState(new.State).
			SetServeNodes(new.ServeNodes).
			SetDefaultStorageType(new.DefaultStorageType).
			SetInstanceName(new.InstanceName).
			SetProjectID(new.ProjectID)

		if new.EncryptionConfigJSON != nil {
			create.SetEncryptionConfigJSON(new.EncryptionConfigJSON)
		}
		if new.ClusterConfigJSON != nil {
			create.SetClusterConfigJSON(new.ClusterConfigJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Bigtable cluster history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Bigtable cluster.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigtableCluster.Query().
		Where(
			bronzehistorygcpbigtablecluster.ResourceID(resourceID),
			bronzehistorygcpbigtablecluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Bigtable cluster history: %w", err)
	}

	err = tx.BronzeHistoryGCPBigtableCluster.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Bigtable cluster history: %w", err)
	}

	return nil
}
