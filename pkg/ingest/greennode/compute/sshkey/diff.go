package sshkey

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SSHKeyDiff represents changes between old and new SSH key states.
type SSHKeyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSSHKeyData compares old Ent entity and new SSHKeyData.
func DiffSSHKeyData(old *ent.BronzeGreenNodeComputeSSHKey, new *SSHKeyData) *SSHKeyDiff {
	if old == nil {
		return &SSHKeyDiff{IsNew: true}
	}

	return &SSHKeyDiff{
		IsChanged: old.Name != new.Name ||
			old.CreatedAtAPI != new.CreatedAtAPI ||
			old.PubKey != new.PubKey ||
			old.Status != new.Status,
	}
}

// HasAnyChange returns true if the SSH key changed.
func (d *SSHKeyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
