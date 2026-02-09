package account

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AccountDiff represents changes between old and new account states.
type AccountDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffAccountData compares old Ent entity and new data.
func DiffAccountData(old *ent.BronzeS1Account, new *AccountData) *AccountDiff {
	if old == nil {
		return &AccountDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.State != new.State ||
		old.AccountType != new.AccountType ||
		old.UnlimitedExpiration != new.UnlimitedExpiration ||
		old.ActiveAgents != new.ActiveAgents ||
		old.TotalLicenses != new.TotalLicenses ||
		old.UsageType != new.UsageType ||
		old.BillingMode != new.BillingMode ||
		old.Creator != new.Creator ||
		old.CreatorID != new.CreatorID ||
		old.NumberOfSites != new.NumberOfSites ||
		old.ExternalID != new.ExternalID ||
		!bytes.Equal(old.LicensesJSON, new.LicensesJSON)

	return &AccountDiff{IsChanged: changed}
}
