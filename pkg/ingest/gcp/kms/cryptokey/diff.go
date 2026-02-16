package cryptokey

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// CryptoKeyDiff represents changes between old and new crypto key states.
type CryptoKeyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffCryptoKeyData compares old Ent entity and new data.
func DiffCryptoKeyData(old *ent.BronzeGCPKMSCryptoKey, new *CryptoKeyData) *CryptoKeyDiff {
	if old == nil {
		return &CryptoKeyDiff{IsNew: true}
	}

	diff := &CryptoKeyDiff{}
	diff.IsChanged = old.Name != new.Name ||
		old.Purpose != new.Purpose ||
		old.CreateTime != new.CreateTime ||
		old.NextRotationTime != new.NextRotationTime ||
		old.RotationPeriod != new.RotationPeriod ||
		old.DestroyScheduledDuration != new.DestroyScheduledDuration ||
		old.ImportOnly != new.ImportOnly ||
		old.CryptoKeyBackend != new.CryptoKeyBackend ||
		!bytes.Equal(old.VersionTemplateJSON, new.VersionTemplateJSON) ||
		!bytes.Equal(old.PrimaryJSON, new.PrimaryJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON)

	return diff
}

// HasAnyChange returns true if the crypto key changed.
func (d *CryptoKeyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
