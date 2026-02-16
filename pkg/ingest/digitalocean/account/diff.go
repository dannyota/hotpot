package account

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AccountDiff represents changes between old and new Account states.
type AccountDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffAccountData compares old Ent entity and new data.
func DiffAccountData(old *ent.BronzeDOAccount, new *AccountData) *AccountDiff {
	if old == nil {
		return &AccountDiff{IsNew: true}
	}

	changed := old.Email != new.Email ||
		old.Name != new.Name ||
		old.Status != new.Status ||
		old.StatusMessage != new.StatusMessage ||
		old.DropletLimit != new.DropletLimit ||
		old.FloatingIPLimit != new.FloatingIPLimit ||
		old.ReservedIPLimit != new.ReservedIPLimit ||
		old.VolumeLimit != new.VolumeLimit ||
		old.EmailVerified != new.EmailVerified ||
		old.TeamName != new.TeamName ||
		old.TeamUUID != new.TeamUUID

	return &AccountDiff{IsChanged: changed}
}
