package account

import "github.com/dannyota/hotpot/pkg/storage/ent"

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

	return &AccountDiff{
		IsChanged: old.Name != new.Name,
	}
}
