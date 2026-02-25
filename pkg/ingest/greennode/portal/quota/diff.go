package quota

import (
	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
)

// QuotaDiff represents changes between old and new quota states.
type QuotaDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffQuotaData compares old Ent entity and new QuotaData.
func DiffQuotaData(old *entportal.BronzeGreenNodePortalQuota, new *QuotaData) *QuotaDiff {
	if old == nil {
		return &QuotaDiff{IsNew: true}
	}

	return &QuotaDiff{
		IsChanged: old.Name != new.Name ||
			old.Description != new.Description ||
			old.Type != new.Type ||
			old.LimitValue != new.LimitValue ||
			old.UsedValue != new.UsedValue,
	}
}

// HasAnyChange returns true if the quota changed.
func (d *QuotaDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
