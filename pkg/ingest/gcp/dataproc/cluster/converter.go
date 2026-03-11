package cluster

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/dataproc/v2/apiv1/dataprocpb"
)

// ClusterData holds converted Dataproc cluster data ready for Ent insertion.
type ClusterData struct {
	ID                string
	ClusterName       string
	ClusterUUID       string
	ConfigJSON        json.RawMessage
	StatusJSON        json.RawMessage
	StatusHistoryJSON json.RawMessage
	LabelsJSON        json.RawMessage
	MetricsJSON       json.RawMessage
	ProjectID         string
	Location          string
	CollectedAt       time.Time
}

// ConvertCluster converts a GCP API Dataproc Cluster to Ent-compatible data.
func ConvertCluster(c *dataprocpb.Cluster, projectID string, collectedAt time.Time) (*ClusterData, error) {
	clusterProject := c.GetProjectId()
	if clusterProject == "" {
		clusterProject = projectID
	}

	region := extractRegion(c)

	data := &ClusterData{
		ID:          fmt.Sprintf("projects/%s/regions/%s/clusters/%s", clusterProject, region, c.GetClusterName()),
		ClusterName: c.GetClusterName(),
		ClusterUUID: c.GetClusterUuid(),
		ProjectID:   clusterProject,
		Location:    region,
		CollectedAt: collectedAt,
	}

	if c.GetConfig() != nil {
		j, err := json.Marshal(c.GetConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config for cluster %s: %w", c.GetClusterName(), err)
		}
		data.ConfigJSON = j
	}

	if c.GetStatus() != nil {
		j, err := json.Marshal(c.GetStatus())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal status for cluster %s: %w", c.GetClusterName(), err)
		}
		data.StatusJSON = j
	}

	if len(c.GetStatusHistory()) > 0 {
		j, err := json.Marshal(c.GetStatusHistory())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal status_history for cluster %s: %w", c.GetClusterName(), err)
		}
		data.StatusHistoryJSON = j
	}

	if len(c.GetLabels()) > 0 {
		j, err := json.Marshal(c.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for cluster %s: %w", c.GetClusterName(), err)
		}
		data.LabelsJSON = j
	}

	if c.GetMetrics() != nil {
		j, err := json.Marshal(c.GetMetrics())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metrics for cluster %s: %w", c.GetClusterName(), err)
		}
		data.MetricsJSON = j
	}

	return data, nil
}

// extractRegion derives the region from the cluster's GCE zone URI.
// Zone URI format: projects/{project}/zones/{zone} where zone is like "us-central1-a".
// The region is the zone minus the last segment (e.g., "us-central1").
func extractRegion(c *dataprocpb.Cluster) string {
	if cfg := c.GetConfig(); cfg != nil {
		if gceCfg := cfg.GetGceClusterConfig(); gceCfg != nil {
			zoneURI := gceCfg.GetZoneUri()
			if zoneURI != "" {
				// Extract zone name from URI
				parts := strings.Split(zoneURI, "/")
				zone := parts[len(parts)-1]
				// Strip last segment to get region (e.g., "us-central1-a" -> "us-central1")
				if idx := strings.LastIndex(zone, "-"); idx > 0 {
					return zone[:idx]
				}
				return zone
			}
		}
	}
	return ""
}
