package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabase"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabasebackup"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabaseconfig"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabasefirewallrule"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabasepool"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabasereplica"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodatabaseuser"
)

// Service handles DigitalOcean Database ingestion.
type Service struct {
	client          *Client
	entClient       *ent.Client
	dbHistory       *DatabaseHistoryService
	fwHistory       *FirewallRuleHistoryService
	userHistory     *UserHistoryService
	replicaHistory  *ReplicaHistoryService
	backupHistory   *BackupHistoryService
	configHistory   *ConfigHistoryService
	poolHistory     *PoolHistoryService
}

// NewService creates a new Database ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:         client,
		entClient:      entClient,
		dbHistory:      NewDatabaseHistoryService(entClient),
		fwHistory:      NewFirewallRuleHistoryService(entClient),
		userHistory:    NewUserHistoryService(entClient),
		replicaHistory: NewReplicaHistoryService(entClient),
		backupHistory:  NewBackupHistoryService(entClient),
		configHistory:  NewConfigHistoryService(entClient),
		poolHistory:    NewPoolHistoryService(entClient),
	}
}

// IngestDatabasesResult contains the result of Database cluster ingestion.
type IngestDatabasesResult struct {
	ClusterCount   int
	CollectedAt    time.Time
	DurationMillis int64
	ClusterIDs     []string
	EngineMap      map[string]string // clusterID -> engineSlug
}

// IngestDatabases fetches all Database clusters from DigitalOcean and saves them.
func (s *Service) IngestDatabases(ctx context.Context, heartbeat func()) (*IngestDatabasesResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiDatabases, err := s.client.ListAllDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allDatabases []*DatabaseData
	var clusterIDs []string
	engineMap := make(map[string]string)
	for _, v := range apiDatabases {
		allDatabases = append(allDatabases, ConvertDatabase(v, collectedAt))
		clusterIDs = append(clusterIDs, v.ID)
		engineMap[v.ID] = v.EngineSlug
	}

	if err := s.saveDatabases(ctx, allDatabases); err != nil {
		return nil, fmt.Errorf("save databases: %w", err)
	}

	return &IngestDatabasesResult{
		ClusterCount:   len(allDatabases),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
		ClusterIDs:     clusterIDs,
		EngineMap:      engineMap,
	}, nil
}

func (s *Service) saveDatabases(ctx context.Context, databases []*DatabaseData) error {
	if len(databases) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range databases {
		existing, err := tx.BronzeDODatabase.Query().
			Where(bronzedodatabase.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Database %s: %w", data.ResourceID, err)
		}

		diff := DiffDatabaseData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabase.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Database %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabase.Create().
				SetID(data.ResourceID).
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
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Database %s: %w", data.ResourceID, err)
			}

			if err := s.dbHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Database %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabase.UpdateOneID(data.ResourceID).
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
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Database %s: %w", data.ResourceID, err)
			}

			if err := s.dbHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Database %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleDatabases removes Database clusters that were not collected in the latest run.
func (s *Service) DeleteStaleDatabases(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeDODatabase.Query().
		Where(bronzedodatabase.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, d := range stale {
		if err := s.dbHistory.CloseHistory(ctx, tx, d.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for Database %s: %w", d.ID, err)
		}

		if err := tx.BronzeDODatabase.DeleteOne(d).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete Database %s: %w", d.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// IngestChildrenResult contains the result of all child resource ingestion.
type IngestChildrenResult struct {
	FirewallRuleCount int
	UserCount         int
	ReplicaCount      int
	BackupCount       int
	ConfigCount       int
	PoolCount         int
	CollectedAt       time.Time
	DurationMillis    int64
}

// IngestChildren fetches all child resources for given clusters and saves them.
func (s *Service) IngestChildren(ctx context.Context, clusterIDs []string, engineMap map[string]string, heartbeat func()) (*IngestChildrenResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allFirewallRules []*FirewallRuleData
	var allUsers []*UserData
	var allReplicas []*ReplicaData
	var allBackups []*BackupData
	var allConfigs []*ConfigData
	var allPools []*PoolData

	for _, clusterID := range clusterIDs {
		// Firewall rules
		apiRules, err := s.client.GetFirewallRules(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("get firewall rules for cluster %s: %w", clusterID, err)
		}
		for _, v := range apiRules {
			allFirewallRules = append(allFirewallRules, ConvertFirewallRule(v, clusterID, collectedAt))
		}

		// Users
		apiUsers, err := s.client.ListAllUsers(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("list users for cluster %s: %w", clusterID, err)
		}
		for _, v := range apiUsers {
			allUsers = append(allUsers, ConvertUser(v, clusterID, collectedAt))
		}

		// Replicas
		apiReplicas, err := s.client.ListAllReplicas(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("list replicas for cluster %s: %w", clusterID, err)
		}
		for _, v := range apiReplicas {
			allReplicas = append(allReplicas, ConvertReplica(v, clusterID, collectedAt))
		}

		// Backups
		apiBackups, err := s.client.ListAllBackups(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("list backups for cluster %s: %w", clusterID, err)
		}
		for _, v := range apiBackups {
			allBackups = append(allBackups, ConvertBackup(v, clusterID, collectedAt))
		}

		// Config
		engineSlug := engineMap[clusterID]
		configJSON, err := s.client.GetConfig(ctx, clusterID, engineSlug)
		if err != nil {
			return nil, fmt.Errorf("get config for cluster %s: %w", clusterID, err)
		}
		if configJSON != nil {
			allConfigs = append(allConfigs, ConvertConfig(clusterID, engineSlug, configJSON, collectedAt))
		}

		// Connection pools
		apiPools, err := s.client.ListAllPools(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("list pools for cluster %s: %w", clusterID, err)
		}
		for _, v := range apiPools {
			allPools = append(allPools, ConvertPool(v, clusterID, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}
	}

	if err := s.saveFirewallRules(ctx, allFirewallRules); err != nil {
		return nil, fmt.Errorf("save firewall rules: %w", err)
	}
	if err := s.saveUsers(ctx, allUsers); err != nil {
		return nil, fmt.Errorf("save users: %w", err)
	}
	if err := s.saveReplicas(ctx, allReplicas); err != nil {
		return nil, fmt.Errorf("save replicas: %w", err)
	}
	if err := s.saveBackups(ctx, allBackups); err != nil {
		return nil, fmt.Errorf("save backups: %w", err)
	}
	if err := s.saveConfigs(ctx, allConfigs); err != nil {
		return nil, fmt.Errorf("save configs: %w", err)
	}
	if err := s.savePools(ctx, allPools); err != nil {
		return nil, fmt.Errorf("save pools: %w", err)
	}

	return &IngestChildrenResult{
		FirewallRuleCount: len(allFirewallRules),
		UserCount:         len(allUsers),
		ReplicaCount:      len(allReplicas),
		BackupCount:       len(allBackups),
		ConfigCount:       len(allConfigs),
		PoolCount:         len(allPools),
		CollectedAt:       collectedAt,
		DurationMillis:    time.Since(startTime).Milliseconds(),
	}, nil
}

// saveFirewallRules saves firewall rule data transactionally.
func (s *Service) saveFirewallRules(ctx context.Context, rules []*FirewallRuleData) error {
	if len(rules) == 0 {
		return nil
	}

	now := time.Now()
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range rules {
		existing, err := tx.BronzeDODatabaseFirewallRule.Query().
			Where(bronzedodatabasefirewallrule.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing FirewallRule %s: %w", data.ResourceID, err)
		}

		diff := DiffFirewallRuleData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabaseFirewallRule.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for FirewallRule %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabaseFirewallRule.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetUUID(data.UUID).
				SetType(data.Type).
				SetValue(data.Value).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create FirewallRule %s: %w", data.ResourceID, err)
			}
			if err := s.fwHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for FirewallRule %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabaseFirewallRule.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetUUID(data.UUID).
				SetType(data.Type).
				SetValue(data.Value).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update FirewallRule %s: %w", data.ResourceID, err)
			}
			if err := s.fwHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for FirewallRule %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// saveUsers saves user data transactionally.
func (s *Service) saveUsers(ctx context.Context, users []*UserData) error {
	if len(users) == 0 {
		return nil
	}

	now := time.Now()
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range users {
		existing, err := tx.BronzeDODatabaseUser.Query().
			Where(bronzedodatabaseuser.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing User %s: %w", data.ResourceID, err)
		}

		diff := DiffUserData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabaseUser.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for User %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabaseUser.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetName(data.Name).
				SetRole(data.Role).
				SetMysqlSettingsJSON(data.MySQLSettingsJSON).
				SetSettingsJSON(data.SettingsJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create User %s: %w", data.ResourceID, err)
			}
			if err := s.userHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for User %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabaseUser.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetName(data.Name).
				SetRole(data.Role).
				SetMysqlSettingsJSON(data.MySQLSettingsJSON).
				SetSettingsJSON(data.SettingsJSON).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update User %s: %w", data.ResourceID, err)
			}
			if err := s.userHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for User %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// saveReplicas saves replica data transactionally.
func (s *Service) saveReplicas(ctx context.Context, replicas []*ReplicaData) error {
	if len(replicas) == 0 {
		return nil
	}

	now := time.Now()
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range replicas {
		existing, err := tx.BronzeDODatabaseReplica.Query().
			Where(bronzedodatabasereplica.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Replica %s: %w", data.ResourceID, err)
		}

		diff := DiffReplicaData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabaseReplica.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Replica %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabaseReplica.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetName(data.Name).
				SetRegion(data.Region).
				SetStatus(data.Status).
				SetSize(data.Size).
				SetStorageSizeMib(data.StorageSizeMib).
				SetPrivateNetworkUUID(data.PrivateNetworkUUID).
				SetTagsJSON(data.TagsJSON).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Replica %s: %w", data.ResourceID, err)
			}
			if err := s.replicaHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Replica %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabaseReplica.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetName(data.Name).
				SetRegion(data.Region).
				SetStatus(data.Status).
				SetSize(data.Size).
				SetStorageSizeMib(data.StorageSizeMib).
				SetPrivateNetworkUUID(data.PrivateNetworkUUID).
				SetTagsJSON(data.TagsJSON).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Replica %s: %w", data.ResourceID, err)
			}
			if err := s.replicaHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Replica %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// saveBackups saves backup data transactionally.
func (s *Service) saveBackups(ctx context.Context, backups []*BackupData) error {
	if len(backups) == 0 {
		return nil
	}

	now := time.Now()
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range backups {
		existing, err := tx.BronzeDODatabaseBackup.Query().
			Where(bronzedodatabasebackup.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Backup %s: %w", data.ResourceID, err)
		}

		diff := DiffBackupData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabaseBackup.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Backup %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabaseBackup.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetSizeGigabytes(data.SizeGigabytes).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Backup %s: %w", data.ResourceID, err)
			}
			if err := s.backupHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Backup %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabaseBackup.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetSizeGigabytes(data.SizeGigabytes).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Backup %s: %w", data.ResourceID, err)
			}
			if err := s.backupHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Backup %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// saveConfigs saves config data transactionally.
func (s *Service) saveConfigs(ctx context.Context, configs []*ConfigData) error {
	if len(configs) == 0 {
		return nil
	}

	now := time.Now()
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range configs {
		existing, err := tx.BronzeDODatabaseConfig.Query().
			Where(bronzedodatabaseconfig.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Config %s: %w", data.ResourceID, err)
		}

		diff := DiffConfigData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabaseConfig.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Config %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabaseConfig.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetEngineSlug(data.EngineSlug).
				SetConfigJSON(data.ConfigJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Config %s: %w", data.ResourceID, err)
			}
			if err := s.configHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Config %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabaseConfig.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetEngineSlug(data.EngineSlug).
				SetConfigJSON(data.ConfigJSON).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Config %s: %w", data.ResourceID, err)
			}
			if err := s.configHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Config %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// savePools saves connection pool data transactionally.
func (s *Service) savePools(ctx context.Context, pools []*PoolData) error {
	if len(pools) == 0 {
		return nil
	}

	now := time.Now()
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range pools {
		existing, err := tx.BronzeDODatabasePool.Query().
			Where(bronzedodatabasepool.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Pool %s: %w", data.ResourceID, err)
		}

		diff := DiffPoolData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODatabasePool.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Pool %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODatabasePool.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetName(data.Name).
				SetUser(data.User).
				SetSize(data.Size).
				SetDatabase(data.Database).
				SetMode(data.Mode).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Pool %s: %w", data.ResourceID, err)
			}
			if err := s.poolHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Pool %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODatabasePool.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetName(data.Name).
				SetUser(data.User).
				SetSize(data.Size).
				SetDatabase(data.Database).
				SetMode(data.Mode).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Pool %s: %w", data.ResourceID, err)
			}
			if err := s.poolHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Pool %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleFirewallRules removes firewall rules not collected in the latest run.
func (s *Service) DeleteStaleFirewallRules(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "FirewallRule",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDODatabaseFirewallRule.Query().
				Where(bronzedodatabasefirewallrule.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDODatabaseFirewallRule.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.fwHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

// DeleteStaleUsers removes users not collected in the latest run.
func (s *Service) DeleteStaleUsers(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "User",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDODatabaseUser.Query().
				Where(bronzedodatabaseuser.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDODatabaseUser.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.userHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

// DeleteStaleReplicas removes replicas not collected in the latest run.
func (s *Service) DeleteStaleReplicas(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "Replica",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDODatabaseReplica.Query().
				Where(bronzedodatabasereplica.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDODatabaseReplica.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.replicaHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

// DeleteStaleBackups removes backups not collected in the latest run.
func (s *Service) DeleteStaleBackups(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "Backup",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDODatabaseBackup.Query().
				Where(bronzedodatabasebackup.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDODatabaseBackup.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.backupHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

// DeleteStaleConfigs removes configs not collected in the latest run.
func (s *Service) DeleteStaleConfigs(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "Config",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDODatabaseConfig.Query().
				Where(bronzedodatabaseconfig.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDODatabaseConfig.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.configHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

// DeleteStalePools removes connection pools not collected in the latest run.
func (s *Service) DeleteStalePools(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "Pool",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDODatabasePool.Query().
				Where(bronzedodatabasepool.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDODatabasePool.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.poolHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

type staleResource struct {
	id     string
	delete func(ctx context.Context) error
}

type queryStaleFunc func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error)
type closeHistoryFunc func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error

func (s *Service) deleteStale(ctx context.Context, collectedAt time.Time, typeName string, queryFn queryStaleFunc, closeFn closeHistoryFunc) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := queryFn(ctx, tx, collectedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, r := range stale {
		if err := closeFn(ctx, tx, r.id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for %s %s: %w", typeName, r.id, err)
		}
		if err := r.delete(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete %s %s: %w", typeName, r.id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
