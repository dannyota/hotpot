package cluster

import (
	"hotpot/pkg/base/models/bronze"
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

// DiffCluster compares old and new cluster states.
func DiffCluster(old, new *bronze.GCPContainerCluster) *ClusterDiff {
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

	// Compare children
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)
	diff.AddonsDiff = diffAddons(old.Addons, new.Addons)
	diff.ConditionsDiff = diffConditions(old.Conditions, new.Conditions)
	diff.NodePoolsDiff = diffNodePools(old.NodePools, new.NodePools)

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
func hasClusterFieldsChanged(old, new *bronze.GCPContainerCluster) bool {
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
		old.AddonsConfigJSON != new.AddonsConfigJSON ||
		old.PrivateClusterConfigJSON != new.PrivateClusterConfigJSON ||
		old.IpAllocationPolicyJSON != new.IpAllocationPolicyJSON ||
		old.NetworkConfigJSON != new.NetworkConfigJSON ||
		old.AutoscalingJSON != new.AutoscalingJSON ||
		old.MaintenancePolicyJSON != new.MaintenancePolicyJSON ||
		old.AutopilotJSON != new.AutopilotJSON ||
		old.ReleaseChannelJSON != new.ReleaseChannelJSON
}

func diffLabels(old, new []bronze.GCPContainerClusterLabel) ChildDiff {
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

func diffAddons(old, new []bronze.GCPContainerClusterAddon) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]bronze.GCPContainerClusterAddon)
	for _, a := range old {
		oldMap[a.AddonName] = a
	}
	for _, a := range new {
		if oldAddon, ok := oldMap[a.AddonName]; !ok ||
			oldAddon.Enabled != a.Enabled ||
			oldAddon.ConfigJSON != a.ConfigJSON {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func diffConditions(old, new []bronze.GCPContainerClusterCondition) ChildDiff {
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

func diffNodePools(old, new []bronze.GCPContainerClusterNodePool) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]bronze.GCPContainerClusterNodePool)
	for _, np := range old {
		oldMap[np.Name] = np
	}
	for _, np := range new {
		oldNP, ok := oldMap[np.Name]
		if !ok || hasNodePoolChanged(&oldNP, &np) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasNodePoolChanged(old, new *bronze.GCPContainerClusterNodePool) bool {
	return old.Version != new.Version ||
		old.Status != new.Status ||
		old.StatusMessage != new.StatusMessage ||
		old.InitialNodeCount != new.InitialNodeCount ||
		old.PodIpv4CidrSize != new.PodIpv4CidrSize ||
		old.LocationsJSON != new.LocationsJSON ||
		old.ConfigJSON != new.ConfigJSON ||
		old.AutoscalingJSON != new.AutoscalingJSON ||
		old.ManagementJSON != new.ManagementJSON ||
		old.UpgradeSettingsJSON != new.UpgradeSettingsJSON ||
		old.NetworkConfigJSON != new.NetworkConfigJSON
}
