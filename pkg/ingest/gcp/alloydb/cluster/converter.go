package cluster

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/alloydb/apiv1/alloydbpb"
)

// ClusterData holds converted AlloyDB cluster data ready for Ent insertion.
type ClusterData struct {
	ID                         string
	Name                       string
	DisplayName                string
	UID                        string
	CreateTime                 string
	UpdateTime                 string
	DeleteTime                 string
	LabelsJSON                 json.RawMessage
	State                      int
	ClusterType                int
	DatabaseVersion            int
	NetworkConfigJSON          json.RawMessage
	Network                    string
	Etag                       string
	AnnotationsJSON            json.RawMessage
	Reconciling                bool
	InitialUserJSON            json.RawMessage
	AutomatedBackupPolicyJSON  json.RawMessage
	SslConfigJSON              json.RawMessage
	EncryptionConfigJSON       json.RawMessage
	EncryptionInfoJSON         json.RawMessage
	ContinuousBackupConfigJSON json.RawMessage
	ContinuousBackupInfoJSON   json.RawMessage
	SecondaryConfigJSON        json.RawMessage
	PrimaryConfigJSON          json.RawMessage
	SatisfiesPzs               bool
	PscConfigJSON              json.RawMessage
	MaintenanceUpdatePolicyJSON json.RawMessage
	MaintenanceScheduleJSON    json.RawMessage
	SubscriptionType           int
	TrialMetadataJSON          json.RawMessage
	TagsJSON                   json.RawMessage
	ProjectID                  string
	Location                   string
	CollectedAt                time.Time
}

// ConvertCluster converts a GCP API AlloyDB Cluster to Ent-compatible data.
func ConvertCluster(c *alloydbpb.Cluster, projectID string, collectedAt time.Time) (*ClusterData, error) {
	data := &ClusterData{
		ID:              c.GetName(),
		Name:            c.GetName(),
		DisplayName:     c.GetDisplayName(),
		UID:             c.GetUid(),
		State:           int(c.GetState()),
		ClusterType:     int(c.GetClusterType()),
		DatabaseVersion: int(c.GetDatabaseVersion()),
		Network:         c.GetNetwork(),
		Etag:            c.GetEtag(),
		Reconciling:     c.GetReconciling(),
		SatisfiesPzs:    c.GetSatisfiesPzs(),
		SubscriptionType: int(c.GetSubscriptionType()),
		ProjectID:       projectID,
		Location:        extractLocation(c.GetName()),
		CollectedAt:     collectedAt,
	}

	if c.GetCreateTime() != nil {
		data.CreateTime = c.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if c.GetUpdateTime() != nil {
		data.UpdateTime = c.GetUpdateTime().AsTime().Format(time.RFC3339)
	}
	if c.GetDeleteTime() != nil {
		data.DeleteTime = c.GetDeleteTime().AsTime().Format(time.RFC3339)
	}

	if len(c.GetLabels()) > 0 {
		j, err := json.Marshal(c.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for cluster %s: %w", c.GetName(), err)
		}
		data.LabelsJSON = j
	}

	if c.GetNetworkConfig() != nil {
		j, err := json.Marshal(c.GetNetworkConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal network_config for cluster %s: %w", c.GetName(), err)
		}
		data.NetworkConfigJSON = j
	}

	if len(c.GetAnnotations()) > 0 {
		j, err := json.Marshal(c.GetAnnotations())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal annotations for cluster %s: %w", c.GetName(), err)
		}
		data.AnnotationsJSON = j
	}

	if c.GetInitialUser() != nil {
		j, err := json.Marshal(c.GetInitialUser())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal initial_user for cluster %s: %w", c.GetName(), err)
		}
		data.InitialUserJSON = j
	}

	if c.GetAutomatedBackupPolicy() != nil {
		j, err := json.Marshal(c.GetAutomatedBackupPolicy())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal automated_backup_policy for cluster %s: %w", c.GetName(), err)
		}
		data.AutomatedBackupPolicyJSON = j
	}

	if c.GetSslConfig() != nil {
		j, err := json.Marshal(c.GetSslConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ssl_config for cluster %s: %w", c.GetName(), err)
		}
		data.SslConfigJSON = j
	}

	if c.GetEncryptionConfig() != nil {
		j, err := json.Marshal(c.GetEncryptionConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption_config for cluster %s: %w", c.GetName(), err)
		}
		data.EncryptionConfigJSON = j
	}

	if c.GetEncryptionInfo() != nil {
		j, err := json.Marshal(c.GetEncryptionInfo())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption_info for cluster %s: %w", c.GetName(), err)
		}
		data.EncryptionInfoJSON = j
	}

	if c.GetContinuousBackupConfig() != nil {
		j, err := json.Marshal(c.GetContinuousBackupConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal continuous_backup_config for cluster %s: %w", c.GetName(), err)
		}
		data.ContinuousBackupConfigJSON = j
	}

	if c.GetContinuousBackupInfo() != nil {
		j, err := json.Marshal(c.GetContinuousBackupInfo())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal continuous_backup_info for cluster %s: %w", c.GetName(), err)
		}
		data.ContinuousBackupInfoJSON = j
	}

	if c.GetSecondaryConfig() != nil {
		j, err := json.Marshal(c.GetSecondaryConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal secondary_config for cluster %s: %w", c.GetName(), err)
		}
		data.SecondaryConfigJSON = j
	}

	if c.GetPrimaryConfig() != nil {
		j, err := json.Marshal(c.GetPrimaryConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal primary_config for cluster %s: %w", c.GetName(), err)
		}
		data.PrimaryConfigJSON = j
	}

	if c.GetPscConfig() != nil {
		j, err := json.Marshal(c.GetPscConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal psc_config for cluster %s: %w", c.GetName(), err)
		}
		data.PscConfigJSON = j
	}

	if c.GetMaintenanceUpdatePolicy() != nil {
		j, err := json.Marshal(c.GetMaintenanceUpdatePolicy())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal maintenance_update_policy for cluster %s: %w", c.GetName(), err)
		}
		data.MaintenanceUpdatePolicyJSON = j
	}

	if c.GetMaintenanceSchedule() != nil {
		j, err := json.Marshal(c.GetMaintenanceSchedule())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal maintenance_schedule for cluster %s: %w", c.GetName(), err)
		}
		data.MaintenanceScheduleJSON = j
	}

	if c.GetTrialMetadata() != nil {
		j, err := json.Marshal(c.GetTrialMetadata())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal trial_metadata for cluster %s: %w", c.GetName(), err)
		}
		data.TrialMetadataJSON = j
	}

	if len(c.GetTags()) > 0 {
		j, err := json.Marshal(c.GetTags())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags for cluster %s: %w", c.GetName(), err)
		}
		data.TagsJSON = j
	}

	return data, nil
}

// extractLocation extracts the location from a cluster resource name.
// Format: projects/{project}/locations/{location}/clusters/{cluster}
func extractLocation(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		if p == "locations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
