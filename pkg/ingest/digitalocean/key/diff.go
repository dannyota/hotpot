package key

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// KeyDiff represents changes between old and new SSH key states.
type KeyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffKeyData compares old Ent entity and new data.
func DiffKeyData(old *ent.BronzeDOKey, new *KeyData) *KeyDiff {
	if old == nil {
		return &KeyDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.Fingerprint != new.Fingerprint ||
		old.PublicKey != new.PublicKey

	return &KeyDiff{IsChanged: changed}
}
