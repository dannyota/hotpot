package organization

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// OrganizationDiff represents changes between old and new organization state.
type OrganizationDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *OrganizationDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffOrganizationData compares existing Ent entity with new OrganizationData and returns differences.
func DiffOrganizationData(old *ent.BronzeGCPOrganization, new *OrganizationData) *OrganizationDiff {
	diff := &OrganizationDiff{}

	// New organization
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	if old.DisplayName != new.DisplayName ||
		old.State != new.State ||
		old.DirectoryCustomerID != new.DirectoryCustomerID ||
		old.UpdateTime != new.UpdateTime ||
		old.DeleteTime != new.DeleteTime ||
		old.Etag != new.Etag {
		diff.IsChanged = true
	}

	return diff
}
