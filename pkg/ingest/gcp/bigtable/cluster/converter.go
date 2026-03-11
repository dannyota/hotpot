package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
)

// ClusterData holds converted Bigtable cluster data ready for Ent insertion.
type ClusterData struct {
	ID                   string
	Location             string
	State                int32
	ServeNodes           int32
	DefaultStorageType   int32
	EncryptionConfigJSON json.RawMessage
	ClusterConfigJSON    json.RawMessage
	InstanceName         string
	ProjectID            string
	CollectedAt          time.Time
}

// clusterStateToInt32 converts a cluster state string to int32.
func clusterStateToInt32(state string) int32 {
	switch state {
	case "READY":
		return 1
	case "CREATING":
		return 2
	case "RESIZING":
		return 3
	case "DISABLED":
		return 4
	default:
		return 0 // STATE_NOT_KNOWN
	}
}

// storageTypeToInt32 converts bigtable StorageType to int32.
func storageTypeToInt32(st bigtable.StorageType) int32 {
	return int32(st)
}

// ConvertCluster converts a raw GCP Bigtable ClusterInfo to Ent-compatible data.
func ConvertCluster(raw ClusterRaw, projectID string, collectedAt time.Time) (*ClusterData, error) {
	data := &ClusterData{
		ID:                 fmt.Sprintf("%s/clusters/%s", raw.InstanceName, raw.Cluster.Name),
		Location:           raw.Cluster.Zone,
		State:              clusterStateToInt32(raw.Cluster.State),
		ServeNodes:         int32(raw.Cluster.ServeNodes),
		DefaultStorageType: storageTypeToInt32(raw.Cluster.StorageType),
		InstanceName:       raw.InstanceName,
		ProjectID:          projectID,
		CollectedAt:        collectedAt,
	}

	if raw.Cluster.KMSKeyName != "" {
		encConfig := map[string]string{"kmsKeyName": raw.Cluster.KMSKeyName}
		encJSON, err := json.Marshal(encConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption config: %w", err)
		}
		data.EncryptionConfigJSON = encJSON
	}

	if raw.Cluster.AutoscalingConfig != nil {
		autoJSON, err := json.Marshal(raw.Cluster.AutoscalingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cluster config: %w", err)
		}
		data.ClusterConfigJSON = autoJSON
	}

	return data, nil
}
