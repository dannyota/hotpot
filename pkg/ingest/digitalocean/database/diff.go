package database

import (
	"bytes"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// DatabaseDiff represents changes between old and new Database cluster states.
type DatabaseDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffDatabaseData compares old Ent entity and new data.
func DiffDatabaseData(old *ent.BronzeDODatabase, new *DatabaseData) *DatabaseDiff {
	if old == nil {
		return &DatabaseDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.EngineSlug != new.EngineSlug ||
		old.VersionSlug != new.VersionSlug ||
		old.NumNodes != new.NumNodes ||
		old.SizeSlug != new.SizeSlug ||
		old.RegionSlug != new.RegionSlug ||
		old.Status != new.Status ||
		old.ProjectID != new.ProjectID ||
		old.StorageSizeMib != new.StorageSizeMib ||
		old.PrivateNetworkUUID != new.PrivateNetworkUUID ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!bytes.Equal(old.MaintenanceWindowJSON, new.MaintenanceWindowJSON) ||
		!ptrTimeEqual(old.APICreatedAt, new.APICreatedAt)

	return &DatabaseDiff{IsChanged: changed}
}

// FirewallRuleDiff represents changes between old and new Database Firewall Rule states.
type FirewallRuleDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffFirewallRuleData compares old Ent entity and new data.
func DiffFirewallRuleData(old *ent.BronzeDODatabaseFirewallRule, new *FirewallRuleData) *FirewallRuleDiff {
	if old == nil {
		return &FirewallRuleDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.UUID != new.UUID ||
		old.Type != new.Type ||
		old.Value != new.Value ||
		!ptrTimeEqual(old.APICreatedAt, new.APICreatedAt)

	return &FirewallRuleDiff{IsChanged: changed}
}

// UserDiff represents changes between old and new Database User states.
type UserDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffUserData compares old Ent entity and new data.
func DiffUserData(old *ent.BronzeDODatabaseUser, new *UserData) *UserDiff {
	if old == nil {
		return &UserDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.Name != new.Name ||
		old.Role != new.Role ||
		!bytes.Equal(old.MysqlSettingsJSON, new.MySQLSettingsJSON) ||
		!bytes.Equal(old.SettingsJSON, new.SettingsJSON)

	return &UserDiff{IsChanged: changed}
}

// ReplicaDiff represents changes between old and new Database Replica states.
type ReplicaDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffReplicaData compares old Ent entity and new data.
func DiffReplicaData(old *ent.BronzeDODatabaseReplica, new *ReplicaData) *ReplicaDiff {
	if old == nil {
		return &ReplicaDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.Name != new.Name ||
		old.Region != new.Region ||
		old.Status != new.Status ||
		old.Size != new.Size ||
		old.StorageSizeMib != new.StorageSizeMib ||
		old.PrivateNetworkUUID != new.PrivateNetworkUUID ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!ptrTimeEqual(old.APICreatedAt, new.APICreatedAt)

	return &ReplicaDiff{IsChanged: changed}
}

// BackupDiff represents changes between old and new Database Backup states.
type BackupDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffBackupData compares old Ent entity and new data.
func DiffBackupData(old *ent.BronzeDODatabaseBackup, new *BackupData) *BackupDiff {
	if old == nil {
		return &BackupDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.SizeGigabytes != new.SizeGigabytes ||
		!ptrTimeEqual(old.APICreatedAt, new.APICreatedAt)

	return &BackupDiff{IsChanged: changed}
}

// ConfigDiff represents changes between old and new Database Config states.
type ConfigDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffConfigData compares old Ent entity and new data.
func DiffConfigData(old *ent.BronzeDODatabaseConfig, new *ConfigData) *ConfigDiff {
	if old == nil {
		return &ConfigDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.EngineSlug != new.EngineSlug ||
		!bytes.Equal(old.ConfigJSON, new.ConfigJSON)

	return &ConfigDiff{IsChanged: changed}
}

// PoolDiff represents changes between old and new Database Pool states.
type PoolDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffPoolData compares old Ent entity and new data.
func DiffPoolData(old *ent.BronzeDODatabasePool, new *PoolData) *PoolDiff {
	if old == nil {
		return &PoolDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.Name != new.Name ||
		old.User != new.User ||
		old.Size != new.Size ||
		old.Database != new.Database ||
		old.Mode != new.Mode

	return &PoolDiff{IsChanged: changed}
}

func ptrTimeEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
