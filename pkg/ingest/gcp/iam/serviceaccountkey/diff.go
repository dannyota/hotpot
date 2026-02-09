package serviceaccountkey

import "github.com/dannyota/hotpot/pkg/storage/ent"

// ServiceAccountKeyDiff represents changes between old and new service account key states.
type ServiceAccountKeyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffServiceAccountKeyData compares existing Ent entity with new ServiceAccountKeyData and returns differences.
func DiffServiceAccountKeyData(old *ent.BronzeGCPIAMServiceAccountKey, new *ServiceAccountKeyData) *ServiceAccountKeyDiff {
	diff := &ServiceAccountKeyDiff{}

	// New service account key
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.ServiceAccountEmail != new.ServiceAccountEmail ||
		old.KeyOrigin != new.KeyOrigin ||
		old.KeyType != new.KeyType ||
		old.KeyAlgorithm != new.KeyAlgorithm ||
		!old.ValidAfterTime.Equal(new.ValidAfterTime) ||
		!old.ValidBeforeTime.Equal(new.ValidBeforeTime) ||
		old.Disabled != new.Disabled {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the service account key changed.
func (d *ServiceAccountKeyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
