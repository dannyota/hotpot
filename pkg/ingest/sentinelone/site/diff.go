package site

import (
	"bytes"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

// SiteDiff represents changes between old and new site states.
type SiteDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSiteData compares old Ent entity and new data.
func DiffSiteData(old *ents1.BronzeS1Site, new *SiteData) *SiteDiff {
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
		old.Description != new.Description ||
		old.ExternalID != new.ExternalID ||
		old.Sku != new.SKU ||
		old.UsageType != new.UsageType ||
		old.UnlimitedExpiration != new.UnlimitedExpiration ||
		old.InheritAccountExpiration != new.InheritAccountExpiration ||
		!bytes.Equal(old.LicensesJSON, new.LicensesJSON)

	return &SiteDiff{IsChanged: changed}
}
