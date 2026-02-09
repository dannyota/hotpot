package serviceaccount

import "github.com/dannyota/hotpot/pkg/storage/ent"

type ServiceAccountDiff struct {
	IsNew     bool
	IsChanged bool
}

func DiffServiceAccountData(old *ent.BronzeGCPIAMServiceAccount, new *ServiceAccountData) *ServiceAccountDiff {
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

func hasFieldsChanged(old *ent.BronzeGCPIAMServiceAccount, new *ServiceAccountData) bool {
	return old.Name != new.Name ||
		old.Email != new.Email ||
		old.DisplayName != new.DisplayName ||
		old.Description != new.Description ||
		old.Oauth2ClientID != new.Oauth2ClientId ||
		old.Disabled != new.Disabled ||
		old.Etag != new.Etag
}
