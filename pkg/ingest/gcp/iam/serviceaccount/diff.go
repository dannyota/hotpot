package serviceaccount

import "hotpot/pkg/base/models/bronze"

type ServiceAccountDiff struct {
	IsNew     bool
	IsChanged bool
}

func DiffServiceAccount(old, new *bronze.GCPIAMServiceAccount) *ServiceAccountDiff {
	if old == nil {
		return &ServiceAccountDiff{IsNew: true}
	}
	return &ServiceAccountDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

func (d *ServiceAccountDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old, new *bronze.GCPIAMServiceAccount) bool {
	return old.Name != new.Name ||
		old.Email != new.Email ||
		old.DisplayName != new.DisplayName ||
		old.Description != new.Description ||
		old.Oauth2ClientId != new.Oauth2ClientId ||
		old.Disabled != new.Disabled ||
		old.Etag != new.Etag
}
