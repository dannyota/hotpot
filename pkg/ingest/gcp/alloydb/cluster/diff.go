package cluster

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ClusterDiff represents changes between old and new AlloyDB cluster states.
type ClusterDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ClusterDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffClusterData compares existing Ent entity with new ClusterData and returns differences.
func DiffClusterData(old *ent.BronzeGCPAlloyDBCluster, new *ClusterData) *ClusterDiff {
	diff := &ClusterDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.DisplayName != new.DisplayName ||
		old.UID != new.UID ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime ||
		old.DeleteTime != new.DeleteTime ||
		old.State != new.State ||
		old.ClusterType != new.ClusterType ||
		old.DatabaseVersion != new.DatabaseVersion ||
		old.Network != new.Network ||
		old.Etag != new.Etag ||
		old.Reconciling != new.Reconciling ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.SubscriptionType != new.SubscriptionType ||
		old.Location != new.Location ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.NetworkConfigJSON, new.NetworkConfigJSON) ||
		!bytes.Equal(old.AnnotationsJSON, new.AnnotationsJSON) ||
		!bytes.Equal(old.InitialUserJSON, new.InitialUserJSON) ||
		!bytes.Equal(old.AutomatedBackupPolicyJSON, new.AutomatedBackupPolicyJSON) ||
		!bytes.Equal(old.SslConfigJSON, new.SslConfigJSON) ||
		!bytes.Equal(old.EncryptionConfigJSON, new.EncryptionConfigJSON) ||
		!bytes.Equal(old.EncryptionInfoJSON, new.EncryptionInfoJSON) ||
		!bytes.Equal(old.ContinuousBackupConfigJSON, new.ContinuousBackupConfigJSON) ||
		!bytes.Equal(old.ContinuousBackupInfoJSON, new.ContinuousBackupInfoJSON) ||
		!bytes.Equal(old.SecondaryConfigJSON, new.SecondaryConfigJSON) ||
		!bytes.Equal(old.PrimaryConfigJSON, new.PrimaryConfigJSON) ||
		!bytes.Equal(old.PscConfigJSON, new.PscConfigJSON) ||
		!bytes.Equal(old.MaintenanceUpdatePolicyJSON, new.MaintenanceUpdatePolicyJSON) ||
		!bytes.Equal(old.MaintenanceScheduleJSON, new.MaintenanceScheduleJSON) ||
		!bytes.Equal(old.TrialMetadataJSON, new.TrialMetadataJSON) ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) {
		diff.IsChanged = true
	}

	return diff
}
