package database

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/protobuf/encoding/protojson"
)

// DatabaseData holds converted Spanner database data ready for Ent insertion.
type DatabaseData struct {
	ResourceID             string
	Name                   string
	State                  int
	CreateTime             string
	RestoreInfoJSON        json.RawMessage
	EncryptionConfigJSON   json.RawMessage
	EncryptionInfoJSON     json.RawMessage
	VersionRetentionPeriod string
	EarliestVersionTime    string
	DefaultLeader          string
	DatabaseDialect        int
	EnableDropProtection   bool
	Reconciling            bool
	InstanceName           string
	ProjectID              string
	CollectedAt            time.Time
}

// ConvertDatabase converts a GCP API Spanner database to DatabaseData.
func ConvertDatabase(db *databasepb.Database, instanceName string, projectID string, collectedAt time.Time) (*DatabaseData, error) {
	data := &DatabaseData{
		ResourceID:             db.GetName(),
		Name:                   db.GetName(),
		State:                  int(db.GetState()),
		VersionRetentionPeriod: db.GetVersionRetentionPeriod(),
		DefaultLeader:          db.GetDefaultLeader(),
		DatabaseDialect:        int(db.GetDatabaseDialect()),
		EnableDropProtection:   db.GetEnableDropProtection(),
		Reconciling:            db.GetReconciling(),
		InstanceName:           instanceName,
		ProjectID:              projectID,
		CollectedAt:            collectedAt,
	}

	// Convert create_time
	if db.GetCreateTime() != nil {
		data.CreateTime = db.GetCreateTime().AsTime().Format(time.RFC3339)
	}

	// Convert earliest_version_time
	if db.GetEarliestVersionTime() != nil {
		data.EarliestVersionTime = db.GetEarliestVersionTime().AsTime().Format(time.RFC3339)
	}

	// Convert restore_info to JSON
	if db.GetRestoreInfo() != nil {
		restoreBytes, err := protojson.Marshal(db.GetRestoreInfo())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal restore_info for database %s: %w", db.GetName(), err)
		}
		data.RestoreInfoJSON = restoreBytes
	}

	// Convert encryption_config to JSON
	if db.GetEncryptionConfig() != nil {
		configBytes, err := protojson.Marshal(db.GetEncryptionConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption_config for database %s: %w", db.GetName(), err)
		}
		data.EncryptionConfigJSON = configBytes
	}

	// Convert encryption_info to JSON
	if len(db.GetEncryptionInfo()) > 0 {
		infoBytes, err := json.Marshal(marshalEncryptionInfoList(db.GetEncryptionInfo()))
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption_info for database %s: %w", db.GetName(), err)
		}
		data.EncryptionInfoJSON = infoBytes
	}

	return data, nil
}

// marshalEncryptionInfoList converts encryption info protos to JSON-serializable maps.
func marshalEncryptionInfoList(infos []*databasepb.EncryptionInfo) []json.RawMessage {
	result := make([]json.RawMessage, 0, len(infos))
	for _, info := range infos {
		bytes, err := protojson.Marshal(info)
		if err != nil {
			continue
		}
		result = append(result, bytes)
	}
	return result
}
