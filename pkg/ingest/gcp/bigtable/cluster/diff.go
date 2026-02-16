package cluster

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ClusterDiff represents changes between old and new Bigtable cluster state.
type ClusterDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ClusterDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffClusterData compares existing Ent entity with new ClusterData and returns differences.
func DiffClusterData(old *ent.BronzeGCPBigtableCluster, new *ClusterData) *ClusterDiff {
	diff := &ClusterDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Location != new.Location ||
		old.State != new.State ||
		old.ServeNodes != new.ServeNodes ||
		old.DefaultStorageType != new.DefaultStorageType ||
		!bytes.Equal(old.EncryptionConfigJSON, new.EncryptionConfigJSON) ||
		!bytes.Equal(old.ClusterConfigJSON, new.ClusterConfigJSON) {
		diff.IsChanged = true
	}

	return diff
}
