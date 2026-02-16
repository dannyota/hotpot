package keyring

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// KeyRingDiff represents changes between old and new key ring states.
type KeyRingDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffKeyRingData compares old Ent entity and new data.
func DiffKeyRingData(old *ent.BronzeGCPKMSKeyRing, new *KeyRingData) *KeyRingDiff {
	if old == nil {
		return &KeyRingDiff{IsNew: true}
	}

	diff := &KeyRingDiff{}
	diff.IsChanged = old.Name != new.Name ||
		old.CreateTime != new.CreateTime ||
		old.Location != new.Location

	return diff
}

// HasAnyChange returns true if the key ring changed.
func (d *KeyRingDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
