package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpalloydbcluster"
)

// HistoryService manages AlloyDB cluster history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new AlloyDB cluster.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ClusterData, now time.Time) error {
	create := tx.BronzeHistoryGCPAlloyDBCluster.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetState(data.State).
		SetClusterType(data.ClusterType).
		SetDatabaseVersion(data.DatabaseVersion).
		SetReconciling(data.Reconciling).
		SetSatisfiesPzs(data.SatisfiesPzs).
		SetSubscriptionType(data.SubscriptionType).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location)

	if data.DisplayName != "" {
		create.SetDisplayName(data.DisplayName)
	}
	if data.UID != "" {
		create.SetUID(data.UID)
	}
	if data.CreateTime != "" {
		create.SetCreateTime(data.CreateTime)
	}
	if data.UpdateTime != "" {
		create.SetUpdateTime(data.UpdateTime)
	}
	if data.DeleteTime != "" {
		create.SetDeleteTime(data.DeleteTime)
	}
	if data.Network != "" {
		create.SetNetwork(data.Network)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.NetworkConfigJSON != nil {
		create.SetNetworkConfigJSON(data.NetworkConfigJSON)
	}
	if data.AnnotationsJSON != nil {
		create.SetAnnotationsJSON(data.AnnotationsJSON)
	}
	if data.InitialUserJSON != nil {
		create.SetInitialUserJSON(data.InitialUserJSON)
	}
	if data.AutomatedBackupPolicyJSON != nil {
		create.SetAutomatedBackupPolicyJSON(data.AutomatedBackupPolicyJSON)
	}
	if data.SslConfigJSON != nil {
		create.SetSslConfigJSON(data.SslConfigJSON)
	}
	if data.EncryptionConfigJSON != nil {
		create.SetEncryptionConfigJSON(data.EncryptionConfigJSON)
	}
	if data.EncryptionInfoJSON != nil {
		create.SetEncryptionInfoJSON(data.EncryptionInfoJSON)
	}
	if data.ContinuousBackupConfigJSON != nil {
		create.SetContinuousBackupConfigJSON(data.ContinuousBackupConfigJSON)
	}
	if data.ContinuousBackupInfoJSON != nil {
		create.SetContinuousBackupInfoJSON(data.ContinuousBackupInfoJSON)
	}
	if data.SecondaryConfigJSON != nil {
		create.SetSecondaryConfigJSON(data.SecondaryConfigJSON)
	}
	if data.PrimaryConfigJSON != nil {
		create.SetPrimaryConfigJSON(data.PrimaryConfigJSON)
	}
	if data.PscConfigJSON != nil {
		create.SetPscConfigJSON(data.PscConfigJSON)
	}
	if data.MaintenanceUpdatePolicyJSON != nil {
		create.SetMaintenanceUpdatePolicyJSON(data.MaintenanceUpdatePolicyJSON)
	}
	if data.MaintenanceScheduleJSON != nil {
		create.SetMaintenanceScheduleJSON(data.MaintenanceScheduleJSON)
	}
	if data.TrialMetadataJSON != nil {
		create.SetTrialMetadataJSON(data.TrialMetadataJSON)
	}
	if data.TagsJSON != nil {
		create.SetTagsJSON(data.TagsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create AlloyDB cluster history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed AlloyDB cluster.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPAlloyDBCluster, new *ClusterData, diff *ClusterDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAlloyDBCluster.Query().
		Where(
			bronzehistorygcpalloydbcluster.ResourceID(old.ID),
			bronzehistorygcpalloydbcluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current AlloyDB cluster history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPAlloyDBCluster.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current AlloyDB cluster history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPAlloyDBCluster.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetState(new.State).
			SetClusterType(new.ClusterType).
			SetDatabaseVersion(new.DatabaseVersion).
			SetReconciling(new.Reconciling).
			SetSatisfiesPzs(new.SatisfiesPzs).
			SetSubscriptionType(new.SubscriptionType).
			SetProjectID(new.ProjectID).
			SetLocation(new.Location)

		if new.DisplayName != "" {
			create.SetDisplayName(new.DisplayName)
		}
		if new.UID != "" {
			create.SetUID(new.UID)
		}
		if new.CreateTime != "" {
			create.SetCreateTime(new.CreateTime)
		}
		if new.UpdateTime != "" {
			create.SetUpdateTime(new.UpdateTime)
		}
		if new.DeleteTime != "" {
			create.SetDeleteTime(new.DeleteTime)
		}
		if new.Network != "" {
			create.SetNetwork(new.Network)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.NetworkConfigJSON != nil {
			create.SetNetworkConfigJSON(new.NetworkConfigJSON)
		}
		if new.AnnotationsJSON != nil {
			create.SetAnnotationsJSON(new.AnnotationsJSON)
		}
		if new.InitialUserJSON != nil {
			create.SetInitialUserJSON(new.InitialUserJSON)
		}
		if new.AutomatedBackupPolicyJSON != nil {
			create.SetAutomatedBackupPolicyJSON(new.AutomatedBackupPolicyJSON)
		}
		if new.SslConfigJSON != nil {
			create.SetSslConfigJSON(new.SslConfigJSON)
		}
		if new.EncryptionConfigJSON != nil {
			create.SetEncryptionConfigJSON(new.EncryptionConfigJSON)
		}
		if new.EncryptionInfoJSON != nil {
			create.SetEncryptionInfoJSON(new.EncryptionInfoJSON)
		}
		if new.ContinuousBackupConfigJSON != nil {
			create.SetContinuousBackupConfigJSON(new.ContinuousBackupConfigJSON)
		}
		if new.ContinuousBackupInfoJSON != nil {
			create.SetContinuousBackupInfoJSON(new.ContinuousBackupInfoJSON)
		}
		if new.SecondaryConfigJSON != nil {
			create.SetSecondaryConfigJSON(new.SecondaryConfigJSON)
		}
		if new.PrimaryConfigJSON != nil {
			create.SetPrimaryConfigJSON(new.PrimaryConfigJSON)
		}
		if new.PscConfigJSON != nil {
			create.SetPscConfigJSON(new.PscConfigJSON)
		}
		if new.MaintenanceUpdatePolicyJSON != nil {
			create.SetMaintenanceUpdatePolicyJSON(new.MaintenanceUpdatePolicyJSON)
		}
		if new.MaintenanceScheduleJSON != nil {
			create.SetMaintenanceScheduleJSON(new.MaintenanceScheduleJSON)
		}
		if new.TrialMetadataJSON != nil {
			create.SetTrialMetadataJSON(new.TrialMetadataJSON)
		}
		if new.TagsJSON != nil {
			create.SetTagsJSON(new.TagsJSON)
		}

		_, err = create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new AlloyDB cluster history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted AlloyDB cluster.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAlloyDBCluster.Query().
		Where(
			bronzehistorygcpalloydbcluster.ResourceID(resourceID),
			bronzehistorygcpalloydbcluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current AlloyDB cluster history: %w", err)
	}

	err = tx.BronzeHistoryGCPAlloyDBCluster.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close AlloyDB cluster history: %w", err)
	}

	return nil
}
