package serviceaccountkey

import "hotpot/pkg/base/models/bronze"

type ServiceAccountKeyDiff struct {
	IsNew     bool
	IsChanged bool
}

func DiffServiceAccountKey(old, new *bronze.GCPIAMServiceAccountKey) *ServiceAccountKeyDiff {
	if old == nil {
		return &ServiceAccountKeyDiff{IsNew: true}
	}
	return &ServiceAccountKeyDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

func (d *ServiceAccountKeyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old, new *bronze.GCPIAMServiceAccountKey) bool {
	return old.Name != new.Name ||
		old.ServiceAccountEmail != new.ServiceAccountEmail ||
		old.KeyOrigin != new.KeyOrigin ||
		old.KeyType != new.KeyType ||
		old.KeyAlgorithm != new.KeyAlgorithm ||
		!old.ValidAfterTime.Equal(new.ValidAfterTime) ||
		!old.ValidBeforeTime.Equal(new.ValidBeforeTime) ||
		old.Disabled != new.Disabled
}
