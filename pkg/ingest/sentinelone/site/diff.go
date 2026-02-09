package site

import "hotpot/pkg/storage/ent"

// SiteDiff represents changes between old and new site states.
type SiteDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSiteData compares old Ent entity and new data.
func DiffSiteData(old *ent.BronzeS1Site, new *SiteData) *SiteDiff {
	if old == nil {
		return &SiteDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.AccountID != new.AccountID ||
		old.AccountName != new.AccountName ||
		old.State != new.State ||
		old.SiteType != new.SiteType ||
		old.Suite != new.Suite ||
		old.Creator != new.Creator ||
		old.CreatorID != new.CreatorID ||
		old.HealthStatus != new.HealthStatus ||
		old.ActiveLicenses != new.ActiveLicenses ||
		old.TotalLicenses != new.TotalLicenses ||
		old.UnlimitedLicenses != new.UnlimitedLicenses ||
		old.IsDefault != new.IsDefault ||
		old.Description != new.Description

	return &SiteDiff{IsChanged: changed}
}
