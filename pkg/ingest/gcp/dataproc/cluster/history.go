package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpdataproccluster"
)

// HistoryService manages Dataproc cluster history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Dataproc cluster.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ClusterData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPDataprocCluster.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetClusterName(data.ClusterName).
		SetClusterUUID(data.ClusterUUID).
		SetConfigJSON(data.ConfigJSON).
		SetStatusJSON(data.StatusJSON).
		SetStatusHistoryJSON(data.StatusHistoryJSON).
		SetLabelsJSON(data.LabelsJSON).
		SetMetricsJSON(data.MetricsJSON).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Dataproc cluster history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Dataproc cluster.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPDataprocCluster, new *ClusterData, diff *ClusterDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPDataprocCluster.Query().
		Where(
			bronzehistorygcpdataproccluster.ResourceID(old.ID),
			bronzehistorygcpdataproccluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Dataproc cluster history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPDataprocCluster.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Dataproc cluster history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPDataprocCluster.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetClusterName(new.ClusterName).
			SetClusterUUID(new.ClusterUUID).
			SetConfigJSON(new.ConfigJSON).
			SetStatusJSON(new.StatusJSON).
			SetStatusHistoryJSON(new.StatusHistoryJSON).
			SetLabelsJSON(new.LabelsJSON).
			SetMetricsJSON(new.MetricsJSON).
			SetProjectID(new.ProjectID).
			SetLocation(new.Location).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Dataproc cluster history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Dataproc cluster.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPDataprocCluster.Query().
		Where(
			bronzehistorygcpdataproccluster.ResourceID(resourceID),
			bronzehistorygcpdataproccluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Dataproc cluster history: %w", err)
	}

	err = tx.BronzeHistoryGCPDataprocCluster.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Dataproc cluster history: %w", err)
	}

	return nil
}
