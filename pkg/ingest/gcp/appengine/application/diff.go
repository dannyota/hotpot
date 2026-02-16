package application

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ApplicationDiff represents changes between old and new App Engine application state.
type ApplicationDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ApplicationDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffApplicationData compares existing Ent entity with new ApplicationData and returns differences.
func DiffApplicationData(old *ent.BronzeGCPAppEngineApplication, new *ApplicationData) *ApplicationDiff {
	diff := &ApplicationDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.AuthDomain != new.AuthDomain ||
		old.LocationID != new.LocationID ||
		old.CodeBucket != new.CodeBucket ||
		old.DefaultCookieExpiration != new.DefaultCookieExpiration ||
		old.ServingStatus != new.ServingStatus ||
		old.DefaultHostname != new.DefaultHostname ||
		old.DefaultBucket != new.DefaultBucket ||
		old.GcrDomain != new.GcrDomain ||
		old.DatabaseType != new.DatabaseType ||
		!bytes.Equal(old.FeatureSettingsJSON, new.FeatureSettingsJSON) ||
		!bytes.Equal(old.IapJSON, new.IapJSON) ||
		!bytes.Equal(old.DispatchRulesJSON, new.DispatchRulesJSON) {
		diff.IsChanged = true
	}

	return diff
}
