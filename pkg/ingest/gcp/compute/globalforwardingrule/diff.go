package globalforwardingrule

import (
	"encoding/json"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// GlobalForwardingRuleDiff represents changes between old and new global forwarding rule states.
type GlobalForwardingRuleDiff struct {
	IsNew      bool
	IsChanged  bool
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if any part of the global forwarding rule changed.
func (d *GlobalForwardingRuleDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelsDiff.HasChanges
}

// DiffGlobalForwardingRuleData compares existing Ent entity with new GlobalForwardingRuleData.
func DiffGlobalForwardingRuleData(old *ent.BronzeGCPComputeGlobalForwardingRule, new *GlobalForwardingRuleData) *GlobalForwardingRuleDiff {
	diff := &GlobalForwardingRuleDiff{}

	// New global forwarding rule
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	diff.IsChanged = hasGlobalForwardingRuleFieldsChanged(old, new)

	// Compare labels
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)

	return diff
}

// hasGlobalForwardingRuleFieldsChanged compares global forwarding rule-level fields (excluding children).
func hasGlobalForwardingRuleFieldsChanged(old *ent.BronzeGCPComputeGlobalForwardingRule, new *GlobalForwardingRuleData) bool {
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
		old.IPCollection != new.IpCollection ||
		old.IPVersion != new.IpVersion ||
		old.IsMirroringCollector != new.IsMirroringCollector ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.LoadBalancingScheme != new.LoadBalancingScheme ||
		old.Network != new.Network ||
		old.NetworkTier != new.NetworkTier ||
		old.NoAutomateDNSZone != new.NoAutomateDnsZone ||
		old.PortRange != new.PortRange ||
		old.PscConnectionID != new.PscConnectionId ||
		old.PscConnectionStatus != new.PscConnectionStatus ||
		old.Region != new.Region ||
		old.SelfLink != new.SelfLink ||
		old.SelfLinkWithID != new.SelfLinkWithId ||
		old.ServiceLabel != new.ServiceLabel ||
		old.ServiceName != new.ServiceName ||
		old.Subnetwork != new.Subnetwork ||
		old.Target != new.Target ||
		jsonChanged(old.PortsJSON, new.PortsJSON) ||
		jsonChanged(old.SourceIPRangesJSON, new.SourceIpRangesJSON) ||
		jsonChanged(old.MetadataFiltersJSON, new.MetadataFiltersJSON) ||
		jsonChanged(old.ServiceDirectoryRegistrationsJSON, new.ServiceDirectoryRegistrationsJSON)
}

func jsonChanged(old, new []interface{}) bool {
	oldBytes, _ := json.Marshal(old)
	newBytes, _ := json.Marshal(new)
	return string(oldBytes) != string(newBytes)
}

func diffLabelsData(old []*ent.BronzeGCPComputeGlobalForwardingRuleLabel, new []LabelData) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old labels
	oldMap := make(map[string]string, len(old))
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}

	// Compare with new labels
	for _, l := range new {
		if oldValue, ok := oldMap[l.Key]; !ok || oldValue != l.Value {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}
