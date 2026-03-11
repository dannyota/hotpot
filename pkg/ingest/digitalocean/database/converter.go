package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// DatabaseData holds converted Database cluster data ready for Ent insertion.
type DatabaseData struct {
	ResourceID            string
	Name                  string
	EngineSlug            string
	VersionSlug           string
	NumNodes              int
	SizeSlug              string
	RegionSlug            string
	Status                string
	ProjectID             string
	StorageSizeMib        uint64
	PrivateNetworkUUID    string
	TagsJSON              json.RawMessage
	MaintenanceWindowJSON json.RawMessage
	APICreatedAt          *time.Time
	CollectedAt           time.Time
}

// ConvertDatabase converts a godo Database to DatabaseData.
func ConvertDatabase(v godo.Database, collectedAt time.Time) *DatabaseData {
	data := &DatabaseData{
		ResourceID:         v.ID,
		Name:               v.Name,
		EngineSlug:         v.EngineSlug,
		VersionSlug:        v.VersionSlug,
		NumNodes:           v.NumNodes,
		SizeSlug:           v.SizeSlug,
		RegionSlug:         v.RegionSlug,
		Status:             v.Status,
		ProjectID:          v.ProjectID,
		StorageSizeMib:     v.StorageSizeMib,
		PrivateNetworkUUID: v.PrivateNetworkUUID,
		CollectedAt:        collectedAt,
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if v.MaintenanceWindow != nil {
		data.MaintenanceWindowJSON, _ = json.Marshal(v.MaintenanceWindow)
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	return data
}

// FirewallRuleData holds converted Database Firewall Rule data ready for Ent insertion.
type FirewallRuleData struct {
	ResourceID   string
	ClusterID    string
	UUID         string
	Type         string
	Value        string
	APICreatedAt *time.Time
	CollectedAt  time.Time
}

// ConvertFirewallRule converts a godo DatabaseFirewallRule to FirewallRuleData.
func ConvertFirewallRule(v godo.DatabaseFirewallRule, clusterID string, collectedAt time.Time) *FirewallRuleData {
	data := &FirewallRuleData{
		ResourceID:  fmt.Sprintf("%s:%s", clusterID, v.UUID),
		ClusterID:   clusterID,
		UUID:        v.UUID,
		Type:        v.Type,
		Value:       v.Value,
		CollectedAt: collectedAt,
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	return data
}

// UserData holds converted Database User data ready for Ent insertion.
type UserData struct {
	ResourceID       string
	ClusterID        string
	Name             string
	Role             string
	MySQLSettingsJSON json.RawMessage
	SettingsJSON     json.RawMessage
	CollectedAt      time.Time
}

// ConvertUser converts a godo DatabaseUser to UserData.
func ConvertUser(v godo.DatabaseUser, clusterID string, collectedAt time.Time) *UserData {
	data := &UserData{
		ResourceID:  fmt.Sprintf("%s:%s", clusterID, v.Name),
		ClusterID:   clusterID,
		Name:        v.Name,
		Role:        v.Role,
		CollectedAt: collectedAt,
	}

	if v.MySQLSettings != nil {
		data.MySQLSettingsJSON, _ = json.Marshal(v.MySQLSettings)
	}

	if v.Settings != nil {
		data.SettingsJSON, _ = json.Marshal(v.Settings)
	}

	return data
}

// ReplicaData holds converted Database Replica data ready for Ent insertion.
type ReplicaData struct {
	ResourceID         string
	ClusterID          string
	Name               string
	Region             string
	Status             string
	Size               string
	StorageSizeMib     uint64
	PrivateNetworkUUID string
	TagsJSON           json.RawMessage
	APICreatedAt       *time.Time
	CollectedAt        time.Time
}

// ConvertReplica converts a godo DatabaseReplica to ReplicaData.
func ConvertReplica(v godo.DatabaseReplica, clusterID string, collectedAt time.Time) *ReplicaData {
	data := &ReplicaData{
		ResourceID:         fmt.Sprintf("%s:%s", clusterID, v.Name),
		ClusterID:          clusterID,
		Name:               v.Name,
		Region:             v.Region,
		Status:             v.Status,
		Size:               v.Size,
		StorageSizeMib:     v.StorageSizeMib,
		PrivateNetworkUUID: v.PrivateNetworkUUID,
		CollectedAt:        collectedAt,
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	return data
}

// BackupData holds converted Database Backup data ready for Ent insertion.
type BackupData struct {
	ResourceID    string
	ClusterID     string
	SizeGigabytes float64
	APICreatedAt  *time.Time
	CollectedAt   time.Time
}

// ConvertBackup converts a godo DatabaseBackup to BackupData.
func ConvertBackup(v godo.DatabaseBackup, clusterID string, collectedAt time.Time) *BackupData {
	data := &BackupData{
		ResourceID:    fmt.Sprintf("%s:%s", clusterID, v.CreatedAt.Format(time.RFC3339)),
		ClusterID:     clusterID,
		SizeGigabytes: v.SizeGigabytes,
		CollectedAt:   collectedAt,
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	return data
}

// ConfigData holds converted Database Config data ready for Ent insertion.
type ConfigData struct {
	ResourceID  string
	ClusterID   string
	EngineSlug  string
	ConfigJSON  json.RawMessage
	CollectedAt time.Time
}

// ConvertConfig creates ConfigData from a cluster ID, engine slug, and raw config JSON.
func ConvertConfig(clusterID, engineSlug string, configJSON json.RawMessage, collectedAt time.Time) *ConfigData {
	return &ConfigData{
		ResourceID:  clusterID,
		ClusterID:   clusterID,
		EngineSlug:  engineSlug,
		ConfigJSON:  configJSON,
		CollectedAt: collectedAt,
	}
}

// PoolData holds converted Database Connection Pool data ready for Ent insertion.
type PoolData struct {
	ResourceID  string
	ClusterID   string
	Name        string
	User        string
	Size        int
	Database    string
	Mode        string
	CollectedAt time.Time
}

// ConvertPool converts a godo DatabasePool to PoolData.
func ConvertPool(v godo.DatabasePool, clusterID string, collectedAt time.Time) *PoolData {
	return &PoolData{
		ResourceID:  fmt.Sprintf("%s:%s", clusterID, v.Name),
		ClusterID:   clusterID,
		Name:        v.Name,
		User:        v.User,
		Size:        v.Size,
		Database:    v.Database,
		Mode:        v.Mode,
		CollectedAt: collectedAt,
	}
}
