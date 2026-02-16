package kubernetes

import (
	"bytes"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ClusterDiff represents changes between old and new Kubernetes cluster states.
type ClusterDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffClusterData compares old Ent entity and new data.
func DiffClusterData(old *ent.BronzeDOKubernetesCluster, new *ClusterData) *ClusterDiff {
	if old == nil {
		return &ClusterDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.RegionSlug != new.RegionSlug ||
		old.VersionSlug != new.VersionSlug ||
		old.ClusterSubnet != new.ClusterSubnet ||
		old.ServiceSubnet != new.ServiceSubnet ||
		old.Ipv4 != new.IPv4 ||
		old.Endpoint != new.Endpoint ||
		old.VpcUUID != new.VPCUUID ||
		old.Ha != new.HA ||
		old.AutoUpgrade != new.AutoUpgrade ||
		old.SurgeUpgrade != new.SurgeUpgrade ||
		old.RegistryEnabled != new.RegistryEnabled ||
		old.StatusState != new.StatusState ||
		old.StatusMessage != new.StatusMessage ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!bytes.Equal(old.MaintenancePolicyJSON, new.MaintenancePolicyJSON) ||
		!bytes.Equal(old.ControlPlaneFirewallJSON, new.ControlPlaneFirewallJSON) ||
		!bytes.Equal(old.AutoscalerConfigJSON, new.AutoscalerConfigJSON) ||
		!ptrTimeEqual(old.APICreatedAt, new.APICreatedAt) ||
		!ptrTimeEqual(old.APIUpdatedAt, new.APIUpdatedAt)

	return &ClusterDiff{IsChanged: changed}
}

// NodePoolDiff represents changes between old and new Kubernetes node pool states.
type NodePoolDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffNodePoolData compares old Ent entity and new data.
func DiffNodePoolData(old *ent.BronzeDOKubernetesNodePool, new *NodePoolData) *NodePoolDiff {
	if old == nil {
		return &NodePoolDiff{IsNew: true}
	}

	changed := old.ClusterID != new.ClusterID ||
		old.NodePoolID != new.NodePoolID ||
		old.Name != new.Name ||
		old.Size != new.Size ||
		old.Count != new.Count ||
		old.AutoScale != new.AutoScale ||
		old.MinNodes != new.MinNodes ||
		old.MaxNodes != new.MaxNodes ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.TaintsJSON, new.TaintsJSON) ||
		!bytes.Equal(old.NodesJSON, new.NodesJSON)

	return &NodePoolDiff{IsChanged: changed}
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
