package account

import (
	"encoding/json"
	"time"
)

// AccountData holds converted account data ready for Ent insertion.
type AccountData struct {
	ResourceID          string
	Name                string
	State               string
	AccountType         string
	APICreatedAt        *time.Time
	APIUpdatedAt        *time.Time
	Expiration          *time.Time
	UnlimitedExpiration bool
	ActiveAgents        int
	TotalLicenses       int
	UsageType           string
	BillingMode         string
	Creator             string
	CreatorID           string
	NumberOfSites       int
	ExternalID          string
	LicensesJSON        json.RawMessage
	CollectedAt         time.Time
}

// ConvertAccount converts an API account to AccountData.
func ConvertAccount(acct APIAccount, collectedAt time.Time) *AccountData {
	data := &AccountData{
		ResourceID:          acct.ID,
		Name:                acct.Name,
		State:               acct.State,
		AccountType:         acct.AccountType,
		APICreatedAt:        acct.CreatedAt,
		APIUpdatedAt:        acct.UpdatedAt,
		UnlimitedExpiration: acct.UnlimitedExpiration,
		ActiveAgents:        acct.ActiveAgents,
		TotalLicenses:       acct.TotalLicenses,
		UsageType:           acct.UsageType,
		BillingMode:         acct.BillingMode,
		Creator:             acct.Creator,
		CreatorID:           acct.CreatorID,
		NumberOfSites:       acct.NumberOfSites,
		ExternalID:          acct.ExternalID,
		LicensesJSON:        acct.Licenses,
		CollectedAt:         collectedAt,
	}

	if acct.Expiration != nil {
		if t, err := time.Parse(time.RFC3339, *acct.Expiration); err == nil {
			data.Expiration = &t
		}
	}

	return data
}
