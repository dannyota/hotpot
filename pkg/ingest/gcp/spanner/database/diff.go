package database

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// DatabaseDiff represents changes between old and new Spanner database state.
type DatabaseDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *DatabaseDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffDatabaseData compares existing Ent entity with new DatabaseData and returns differences.
func DiffDatabaseData(old *ent.BronzeGCPSpannerDatabase, new *DatabaseData) *DatabaseDiff {
	diff := &DatabaseDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.State != new.State ||
		old.CreateTime != new.CreateTime ||
		old.VersionRetentionPeriod != new.VersionRetentionPeriod ||
		old.EarliestVersionTime != new.EarliestVersionTime ||
		old.DefaultLeader != new.DefaultLeader ||
		old.DatabaseDialect != new.DatabaseDialect ||
		old.EnableDropProtection != new.EnableDropProtection ||
		old.Reconciling != new.Reconciling ||
		old.InstanceName != new.InstanceName ||
		!bytes.Equal(old.RestoreInfoJSON, new.RestoreInfoJSON) ||
		!bytes.Equal(old.EncryptionConfigJSON, new.EncryptionConfigJSON) ||
		!bytes.Equal(old.EncryptionInfoJSON, new.EncryptionInfoJSON) {
		diff.IsChanged = true
	}

	return diff
}
