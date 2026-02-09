package account

import "time"

// AccountData holds converted account data ready for Ent insertion.
type AccountData struct {
	ResourceID  string
	Name        string
	CollectedAt time.Time
}

// ConvertAccount converts an API account to AccountData.
func ConvertAccount(acct APIAccount, collectedAt time.Time) *AccountData {
	return &AccountData{
		ResourceID:  acct.ID,
		Name:        acct.Name,
		CollectedAt: collectedAt,
	}
}
