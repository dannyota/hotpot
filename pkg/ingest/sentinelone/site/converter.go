package site

import "time"

// SiteData holds converted site data ready for Ent insertion.
type SiteData struct {
	ResourceID        string
	Name              string
	AccountID         string
	AccountName       string
	State             string
	SiteType          string
	Suite             string
	Creator           string
	CreatorID         string
	HealthStatus      bool
	ActiveLicenses    int
	TotalLicenses     int
	UnlimitedLicenses bool
	IsDefault         bool
	Description       string
	APICreatedAt      *time.Time
	Expiration        *time.Time
	CollectedAt       time.Time
}

// ConvertSite converts an API site to SiteData.
func ConvertSite(s APISite, collectedAt time.Time) *SiteData {
	data := &SiteData{
		ResourceID:        s.ID,
		Name:              s.Name,
		AccountID:         s.AccountID,
		AccountName:       s.AccountName,
		State:             s.State,
		SiteType:          s.SiteType,
		Suite:             s.Suite,
		Creator:           s.Creator,
		CreatorID:         s.CreatorID,
		HealthStatus:      s.HealthStatus,
		ActiveLicenses:    s.ActiveLicenses,
		TotalLicenses:     s.TotalLicenses,
		UnlimitedLicenses: s.UnlimitedLicenses,
		IsDefault:         s.IsDefault,
		Description:       s.Description,
		CollectedAt:       collectedAt,
	}

	if s.CreatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *s.CreatedAt); err == nil {
			data.APICreatedAt = &t
		}
	}
	if s.Expiration != nil {
		if t, err := time.Parse(time.RFC3339, *s.Expiration); err == nil {
			data.Expiration = &t
		}
	}

	return data
}
