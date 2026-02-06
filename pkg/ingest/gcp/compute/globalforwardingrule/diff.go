package globalforwardingrule

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// GlobalForwardingRuleDiff represents changes between old and new global forwarding rule states.
type GlobalForwardingRuleDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffGlobalForwardingRule compares old and new global forwarding rule states.
// Returns nil if old is nil (new forwarding rule).
func DiffGlobalForwardingRule(old, new *bronze.GCPComputeGlobalForwardingRule) *GlobalForwardingRuleDiff {
	if old == nil {
		return &GlobalForwardingRuleDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &GlobalForwardingRuleDiff{}

	// Compare forwarding rule-level fields
	diff.IsChanged = hasGlobalForwardingRuleFieldsChanged(old, new)

	// Compare children
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the global forwarding rule changed.
func (d *GlobalForwardingRuleDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

// hasGlobalForwardingRuleFieldsChanged compares forwarding rule-level fields (excluding children).
func hasGlobalForwardingRuleFieldsChanged(old, new *bronze.GCPComputeGlobalForwardingRule) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.IPAddress != new.IPAddress ||
		old.IPProtocol != new.IPProtocol ||
		old.AllPorts != new.AllPorts ||
		old.AllowGlobalAccess != new.AllowGlobalAccess ||
		old.AllowPscGlobalAccess != new.AllowPscGlobalAccess ||
		old.BackendService != new.BackendService ||
		old.BaseForwardingRule != new.BaseForwardingRule ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.ExternalManagedBackendBucketMigrationState != new.ExternalManagedBackendBucketMigrationState ||
		old.ExternalManagedBackendBucketMigrationTestingPercentage != new.ExternalManagedBackendBucketMigrationTestingPercentage ||
		old.Fingerprint != new.Fingerprint ||
		old.IpCollection != new.IpCollection ||
		old.IpVersion != new.IpVersion ||
		old.IsMirroringCollector != new.IsMirroringCollector ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.LoadBalancingScheme != new.LoadBalancingScheme ||
		old.Network != new.Network ||
		old.NetworkTier != new.NetworkTier ||
		old.NoAutomateDnsZone != new.NoAutomateDnsZone ||
		old.PortRange != new.PortRange ||
		old.PscConnectionId != new.PscConnectionId ||
		old.PscConnectionStatus != new.PscConnectionStatus ||
		old.Region != new.Region ||
		old.SelfLink != new.SelfLink ||
		old.SelfLinkWithId != new.SelfLinkWithId ||
		old.ServiceLabel != new.ServiceLabel ||
		old.ServiceName != new.ServiceName ||
		old.Subnetwork != new.Subnetwork ||
		old.Target != new.Target ||
		jsonb.Changed(old.PortsJSON, new.PortsJSON) ||
		jsonb.Changed(old.SourceIpRangesJSON, new.SourceIpRangesJSON) ||
		jsonb.Changed(old.MetadataFiltersJSON, new.MetadataFiltersJSON) ||
		jsonb.Changed(old.ServiceDirectoryRegistrationsJSON, new.ServiceDirectoryRegistrationsJSON)
}

func diffLabels(old, new []bronze.GCPComputeGlobalForwardingRuleLabel) ChildDiff {
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
