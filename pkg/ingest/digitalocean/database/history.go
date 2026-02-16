package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabase"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabasebackup"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabaseconfig"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabasefirewallrule"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabasepool"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabasereplica"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodatabaseuser"
)

// DatabaseHistoryService handles history tracking for Database clusters.
type DatabaseHistoryService struct {
	entClient *ent.Client
}

func NewDatabaseHistoryService(entClient *ent.Client) *DatabaseHistoryService {
	return &DatabaseHistoryService{entClient: entClient}
}

func (h *DatabaseHistoryService) buildCreate(tx *ent.Tx, data *DatabaseData) *ent.BronzeHistoryDODatabaseCreate {
	return tx.BronzeHistoryDODatabase.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetEngineSlug(data.EngineSlug).
		SetVersionSlug(data.VersionSlug).
		SetNumNodes(data.NumNodes).
		SetSizeSlug(data.SizeSlug).
		SetRegionSlug(data.RegionSlug).
		SetStatus(data.Status).
		SetProjectID(data.ProjectID).
		SetStorageSizeMib(data.StorageSizeMib).
		SetPrivateNetworkUUID(data.PrivateNetworkUUID).
		SetTagsJSON(data.TagsJSON).
		SetMaintenanceWindowJSON(data.MaintenanceWindowJSON).
		SetNillableAPICreatedAt(data.APICreatedAt)
}

func (h *DatabaseHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *DatabaseData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Database history: %w", err)
	}
	return nil
}

func (h *DatabaseHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabase, new *DatabaseData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabase.Query().
		Where(
			bronzehistorydodatabase.ResourceID(old.ID),
			bronzehistorydodatabase.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Database history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabase.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Database history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Database history: %w", err)
	}

	return nil
}

func (h *DatabaseHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabase.Query().
		Where(
			bronzehistorydodatabase.ResourceID(resourceID),
			bronzehistorydodatabase.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Database history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabase.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Database history: %w", err)
	}

	return nil
}

// FirewallRuleHistoryService handles history tracking for Database Firewall Rules.
type FirewallRuleHistoryService struct {
	entClient *ent.Client
}

func NewFirewallRuleHistoryService(entClient *ent.Client) *FirewallRuleHistoryService {
	return &FirewallRuleHistoryService{entClient: entClient}
}

func (h *FirewallRuleHistoryService) buildCreate(tx *ent.Tx, data *FirewallRuleData) *ent.BronzeHistoryDODatabaseFirewallRuleCreate {
	return tx.BronzeHistoryDODatabaseFirewallRule.Create().
		SetResourceID(data.ResourceID).
		SetClusterID(data.ClusterID).
		SetUUID(data.UUID).
		SetType(data.Type).
		SetValue(data.Value).
		SetNillableAPICreatedAt(data.APICreatedAt)
}

func (h *FirewallRuleHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *FirewallRuleData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create FirewallRule history: %w", err)
	}
	return nil
}

func (h *FirewallRuleHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabaseFirewallRule, new *FirewallRuleData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseFirewallRule.Query().
		Where(
			bronzehistorydodatabasefirewallrule.ResourceID(old.ID),
			bronzehistorydodatabasefirewallrule.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current FirewallRule history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseFirewallRule.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close FirewallRule history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new FirewallRule history: %w", err)
	}

	return nil
}

func (h *FirewallRuleHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseFirewallRule.Query().
		Where(
			bronzehistorydodatabasefirewallrule.ResourceID(resourceID),
			bronzehistorydodatabasefirewallrule.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current FirewallRule history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseFirewallRule.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close FirewallRule history: %w", err)
	}

	return nil
}

// UserHistoryService handles history tracking for Database Users.
type UserHistoryService struct {
	entClient *ent.Client
}

func NewUserHistoryService(entClient *ent.Client) *UserHistoryService {
	return &UserHistoryService{entClient: entClient}
}

func (h *UserHistoryService) buildCreate(tx *ent.Tx, data *UserData) *ent.BronzeHistoryDODatabaseUserCreate {
	return tx.BronzeHistoryDODatabaseUser.Create().
		SetResourceID(data.ResourceID).
		SetClusterID(data.ClusterID).
		SetName(data.Name).
		SetRole(data.Role).
		SetMysqlSettingsJSON(data.MySQLSettingsJSON).
		SetSettingsJSON(data.SettingsJSON)
}

func (h *UserHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *UserData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create User history: %w", err)
	}
	return nil
}

func (h *UserHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabaseUser, new *UserData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseUser.Query().
		Where(
			bronzehistorydodatabaseuser.ResourceID(old.ID),
			bronzehistorydodatabaseuser.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current User history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseUser.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close User history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new User history: %w", err)
	}

	return nil
}

func (h *UserHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseUser.Query().
		Where(
			bronzehistorydodatabaseuser.ResourceID(resourceID),
			bronzehistorydodatabaseuser.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current User history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseUser.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close User history: %w", err)
	}

	return nil
}

// ReplicaHistoryService handles history tracking for Database Replicas.
type ReplicaHistoryService struct {
	entClient *ent.Client
}

func NewReplicaHistoryService(entClient *ent.Client) *ReplicaHistoryService {
	return &ReplicaHistoryService{entClient: entClient}
}

func (h *ReplicaHistoryService) buildCreate(tx *ent.Tx, data *ReplicaData) *ent.BronzeHistoryDODatabaseReplicaCreate {
	return tx.BronzeHistoryDODatabaseReplica.Create().
		SetResourceID(data.ResourceID).
		SetClusterID(data.ClusterID).
		SetName(data.Name).
		SetRegion(data.Region).
		SetStatus(data.Status).
		SetSize(data.Size).
		SetStorageSizeMib(data.StorageSizeMib).
		SetPrivateNetworkUUID(data.PrivateNetworkUUID).
		SetTagsJSON(data.TagsJSON).
		SetNillableAPICreatedAt(data.APICreatedAt)
}

func (h *ReplicaHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ReplicaData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Replica history: %w", err)
	}
	return nil
}

func (h *ReplicaHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabaseReplica, new *ReplicaData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseReplica.Query().
		Where(
			bronzehistorydodatabasereplica.ResourceID(old.ID),
			bronzehistorydodatabasereplica.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Replica history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseReplica.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Replica history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Replica history: %w", err)
	}

	return nil
}

func (h *ReplicaHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseReplica.Query().
		Where(
			bronzehistorydodatabasereplica.ResourceID(resourceID),
			bronzehistorydodatabasereplica.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Replica history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseReplica.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Replica history: %w", err)
	}

	return nil
}

// BackupHistoryService handles history tracking for Database Backups.
type BackupHistoryService struct {
	entClient *ent.Client
}

func NewBackupHistoryService(entClient *ent.Client) *BackupHistoryService {
	return &BackupHistoryService{entClient: entClient}
}

func (h *BackupHistoryService) buildCreate(tx *ent.Tx, data *BackupData) *ent.BronzeHistoryDODatabaseBackupCreate {
	return tx.BronzeHistoryDODatabaseBackup.Create().
		SetResourceID(data.ResourceID).
		SetClusterID(data.ClusterID).
		SetSizeGigabytes(data.SizeGigabytes).
		SetNillableAPICreatedAt(data.APICreatedAt)
}

func (h *BackupHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *BackupData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Backup history: %w", err)
	}
	return nil
}

func (h *BackupHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabaseBackup, new *BackupData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseBackup.Query().
		Where(
			bronzehistorydodatabasebackup.ResourceID(old.ID),
			bronzehistorydodatabasebackup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Backup history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseBackup.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Backup history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Backup history: %w", err)
	}

	return nil
}

func (h *BackupHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseBackup.Query().
		Where(
			bronzehistorydodatabasebackup.ResourceID(resourceID),
			bronzehistorydodatabasebackup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Backup history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseBackup.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Backup history: %w", err)
	}

	return nil
}

// ConfigHistoryService handles history tracking for Database Configs.
type ConfigHistoryService struct {
	entClient *ent.Client
}

func NewConfigHistoryService(entClient *ent.Client) *ConfigHistoryService {
	return &ConfigHistoryService{entClient: entClient}
}

func (h *ConfigHistoryService) buildCreate(tx *ent.Tx, data *ConfigData) *ent.BronzeHistoryDODatabaseConfigCreate {
	return tx.BronzeHistoryDODatabaseConfig.Create().
		SetResourceID(data.ResourceID).
		SetClusterID(data.ClusterID).
		SetEngineSlug(data.EngineSlug).
		SetConfigJSON(data.ConfigJSON)
}

func (h *ConfigHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ConfigData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Config history: %w", err)
	}
	return nil
}

func (h *ConfigHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabaseConfig, new *ConfigData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseConfig.Query().
		Where(
			bronzehistorydodatabaseconfig.ResourceID(old.ID),
			bronzehistorydodatabaseconfig.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Config history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseConfig.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Config history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Config history: %w", err)
	}

	return nil
}

func (h *ConfigHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabaseConfig.Query().
		Where(
			bronzehistorydodatabaseconfig.ResourceID(resourceID),
			bronzehistorydodatabaseconfig.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Config history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabaseConfig.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Config history: %w", err)
	}

	return nil
}

// PoolHistoryService handles history tracking for Database Connection Pools.
type PoolHistoryService struct {
	entClient *ent.Client
}

func NewPoolHistoryService(entClient *ent.Client) *PoolHistoryService {
	return &PoolHistoryService{entClient: entClient}
}

func (h *PoolHistoryService) buildCreate(tx *ent.Tx, data *PoolData) *ent.BronzeHistoryDODatabasePoolCreate {
	return tx.BronzeHistoryDODatabasePool.Create().
		SetResourceID(data.ResourceID).
		SetClusterID(data.ClusterID).
		SetName(data.Name).
		SetUser(data.User).
		SetSize(data.Size).
		SetDatabase(data.Database).
		SetMode(data.Mode)
}

func (h *PoolHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *PoolData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Pool history: %w", err)
	}
	return nil
}

func (h *PoolHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODatabasePool, new *PoolData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabasePool.Query().
		Where(
			bronzehistorydodatabasepool.ResourceID(old.ID),
			bronzehistorydodatabasepool.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Pool history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabasePool.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Pool history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Pool history: %w", err)
	}

	return nil
}

func (h *PoolHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODatabasePool.Query().
		Where(
			bronzehistorydodatabasepool.ResourceID(resourceID),
			bronzehistorydodatabasepool.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Pool history: %w", err)
	}

	if err := tx.BronzeHistoryDODatabasePool.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Pool history: %w", err)
	}

	return nil
}
