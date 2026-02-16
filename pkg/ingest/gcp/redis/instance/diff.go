package instance

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InstanceDiff represents changes between old and new Redis instance state.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *InstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffInstanceData compares existing Ent entity with new InstanceData and returns differences.
func DiffInstanceData(old *ent.BronzeGCPRedisInstance, new *InstanceData) *InstanceDiff {
	diff := &InstanceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.DisplayName != new.DisplayName ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		old.LocationID != new.LocationID ||
		old.AlternativeLocationID != new.AlternativeLocationID ||
		old.RedisVersion != new.RedisVersion ||
		old.ReservedIPRange != new.ReservedIPRange ||
		old.SecondaryIPRange != new.SecondaryIPRange ||
		old.Host != new.Host ||
		old.Port != new.Port ||
		old.CurrentLocationID != new.CurrentLocationID ||
		old.CreateTime != new.CreateTime ||
		old.State != new.State ||
		old.StatusMessage != new.StatusMessage ||
		!bytes.Equal(old.RedisConfigsJSON, new.RedisConfigsJSON) ||
		old.Tier != new.Tier ||
		old.MemorySizeGB != new.MemorySizeGB ||
		old.AuthorizedNetwork != new.AuthorizedNetwork ||
		old.PersistenceIamIdentity != new.PersistenceIamIdentity ||
		old.ConnectMode != new.ConnectMode ||
		old.AuthEnabled != new.AuthEnabled ||
		!bytes.Equal(old.ServerCaCertsJSON, new.ServerCaCertsJSON) ||
		old.TransitEncryptionMode != new.TransitEncryptionMode ||
		!bytes.Equal(old.MaintenancePolicyJSON, new.MaintenancePolicyJSON) ||
		!bytes.Equal(old.MaintenanceScheduleJSON, new.MaintenanceScheduleJSON) ||
		old.ReplicaCount != new.ReplicaCount ||
		!bytes.Equal(old.NodesJSON, new.NodesJSON) ||
		old.ReadEndpoint != new.ReadEndpoint ||
		old.ReadEndpointPort != new.ReadEndpointPort ||
		old.ReadReplicasMode != new.ReadReplicasMode ||
		old.CustomerManagedKey != new.CustomerManagedKey ||
		!bytes.Equal(old.PersistenceConfigJSON, new.PersistenceConfigJSON) ||
		!bytes.Equal(old.SuspensionReasonsJSON, new.SuspensionReasonsJSON) ||
		old.MaintenanceVersion != new.MaintenanceVersion ||
		!bytes.Equal(old.AvailableMaintenanceVersionsJSON, new.AvailableMaintenanceVersionsJSON) {
		diff.IsChanged = true
	}

	return diff
}
