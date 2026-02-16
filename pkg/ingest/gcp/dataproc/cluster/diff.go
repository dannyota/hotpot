package cluster

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ClusterDiff represents changes between old and new Dataproc cluster state.
type ClusterDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ClusterDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffClusterData compares existing Ent entity with new ClusterData and returns differences.
func DiffClusterData(old *ent.BronzeGCPDataprocCluster, new *ClusterData) *ClusterDiff {
	diff := &ClusterDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.ClusterName != new.ClusterName ||
		old.ClusterUUID != new.ClusterUUID ||
		old.Location != new.Location ||
		!bytes.Equal(old.ConfigJSON, new.ConfigJSON) ||
		!bytes.Equal(old.StatusJSON, new.StatusJSON) ||
		!bytes.Equal(old.StatusHistoryJSON, new.StatusHistoryJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.MetricsJSON, new.MetricsJSON) {
		diff.IsChanged = true
	}

	return diff
}
