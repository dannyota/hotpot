package attestor

import (
	"bytes"

	entbinaryauthorization "danny.vn/hotpot/pkg/storage/ent/gcp/binaryauthorization"
)

// AttestorDiff represents changes between old and new Binary Authorization attestor state.
type AttestorDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *AttestorDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffAttestorData compares existing Ent entity with new AttestorData and returns differences.
func DiffAttestorData(old *entbinaryauthorization.BronzeGCPBinaryAuthorizationAttestor, new *AttestorData) *AttestorDiff {
	diff := &AttestorDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Description != new.Description ||
		!bytes.Equal(old.UserOwnedGrafeasNoteJSON, new.UserOwnedGrafeasNoteJSON) ||
		old.UpdateTime != new.UpdateTime ||
		old.Etag != new.Etag {
		diff.IsChanged = true
	}

	return diff
}
