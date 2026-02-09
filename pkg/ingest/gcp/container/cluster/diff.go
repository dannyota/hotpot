package cluster

import (
	"bytes"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ClusterDiff represents changes between old and new cluster states.
type ClusterDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelsDiff     ChildDiff
	AddonsDiff     ChildDiff
	ConditionsDiff ChildDiff
	NodePoolsDiff  ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffClusterData compares old Ent entity and new data.
func DiffClusterData(old *ent.BronzeGCPContainerCluster, new *ClusterData) *ClusterDiff {
	if old == nil {
		return &ClusterDiff{
			IsNew:          true,
			LabelsDiff:     ChildDiff{Changed: true},
			AddonsDiff:     ChildDiff{Changed: true},
			ConditionsDiff: ChildDiff{Changed: true},
			NodePoolsDiff:  ChildDiff{Changed: true},
		}
	}

	diff := &ClusterDiff{}

	// Compare cluster-level fields
	diff.IsChanged = hasClusterFieldsChanged(old, new)

	// Compare children (need to load edges from old)
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)
	diff.AddonsDiff = diffAddonsData(old.Edges.Addons, new.Addons)
	diff.ConditionsDiff = diffConditionsData(old.Edges.Conditions, new.Conditions)
	diff.NodePoolsDiff = diffNodePoolsData(old.Edges.NodePools, new.NodePools)

	return diff
}

// HasAnyChange returns true if any part of the cluster changed.
func (d *ClusterDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed ||
		d.AddonsDiff.Changed ||
		d.ConditionsDiff.Changed ||
		d.NodePoolsDiff.Changed
}

// hasClusterFieldsChanged compares cluster-level fields (excluding children).
func hasClusterFieldsChanged(old *ent.BronzeGCPContainerCluster, new *ClusterData) bool {
	return old.Name != new.Name ||
		old.Location != new.Location ||
		old.Zone != new.Zone ||
		old.Description != new.Description ||
		old.InitialClusterVersion != new.InitialClusterVersion ||
		old.CurrentMasterVersion != new.CurrentMasterVersion ||
		old.CurrentNodeVersion != new.CurrentNodeVersion ||
		old.Status != new.Status ||
		old.StatusMessage != new.StatusMessage ||
		old.CurrentNodeCount != new.CurrentNodeCount ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.ClusterIpv4Cidr != new.ClusterIpv4Cidr ||
		old.ServicesIpv4Cidr != new.ServicesIpv4Cidr ||
		old.Endpoint != new.Endpoint ||
		old.LoggingService != new.LoggingService ||
		old.MonitoringService != new.MonitoringService ||
		old.EnableKubernetesAlpha != new.EnableKubernetesAlpha ||
		old.EnableTpu != new.EnableTpu ||
		!bytes.Equal(old.AddonsConfigJSON, new.AddonsConfigJSON) ||
		!bytes.Equal(old.PrivateClusterConfigJSON, new.PrivateClusterConfigJSON) ||
		!bytes.Equal(old.IPAllocationPolicyJSON, new.IPAllocationPolicyJSON) ||
		!bytes.Equal(old.NetworkConfigJSON, new.NetworkConfigJSON) ||
		!bytes.Equal(old.AutoscalingJSON, new.AutoscalingJSON) ||
		!bytes.Equal(old.MaintenancePolicyJSON, new.MaintenancePolicyJSON) ||
		!bytes.Equal(old.AutopilotJSON, new.AutopilotJSON) ||
		!bytes.Equal(old.ReleaseChannelJSON, new.ReleaseChannelJSON)
}

func diffLabelsData(old []*ent.BronzeGCPContainerClusterLabel, new []LabelData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]string)
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}
	for _, l := range new {
		if v, ok := oldMap[l.Key]; !ok || v != l.Value {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffAddonsData(old []*ent.BronzeGCPContainerClusterAddon, new []AddonData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*ent.BronzeGCPContainerClusterAddon)
	for _, a := range old {
		oldMap[a.AddonName] = a
	}
	for _, a := range new {
		if oldAddon, ok := oldMap[a.AddonName]; !ok ||
			oldAddon.Enabled != a.Enabled ||
			!bytes.Equal(oldAddon.ConfigJSON, a.ConfigJSON) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffConditionsData(old []*ent.BronzeGCPContainerClusterCondition, new []ConditionData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if old[i].Code != new[i].Code ||
			old[i].Message != new[i].Message ||
			old[i].CanonicalCode != new[i].CanonicalCode {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffNodePoolsData(old []*ent.BronzeGCPContainerClusterNodePool, new []NodePoolData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*ent.BronzeGCPContainerClusterNodePool)
	for _, np := range old {
		oldMap[np.Name] = np
	}
	for _, np := range new {
		oldNP, ok := oldMap[np.Name]
		if !ok || hasNodePoolChangedData(oldNP, &np) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasNodePoolChangedData(old *ent.BronzeGCPContainerClusterNodePool, new *NodePoolData) bool {
	return old.Version != new.Version ||
		old.Status != new.Status ||
		old.StatusMessage != new.StatusMessage ||
		old.InitialNodeCount != new.InitialNodeCount ||
		old.PodIpv4CidrSize != new.PodIpv4CidrSize ||
		!bytes.Equal(old.LocationsJSON, new.LocationsJSON) ||
		!bytes.Equal(old.ConfigJSON, new.ConfigJSON) ||
		!bytes.Equal(old.AutoscalingJSON, new.AutoscalingJSON) ||
		!bytes.Equal(old.ManagementJSON, new.ManagementJSON) ||
		!bytes.Equal(old.UpgradeSettingsJSON, new.UpgradeSettingsJSON) ||
		!bytes.Equal(old.NetworkConfigJSON, new.NetworkConfigJSON)
}
