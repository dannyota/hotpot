package settings

import (
	"bytes"

	entiap "danny.vn/hotpot/pkg/storage/ent/gcp/iap"
)

// SettingsDiff represents changes between old and new IAP settings state.
type SettingsDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *SettingsDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffSettingsData compares existing Ent entity with new SettingsData and returns differences.
func DiffSettingsData(old *entiap.BronzeGCPIAPSettings, new *SettingsData) *SettingsDiff {
	diff := &SettingsDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		!bytes.Equal(old.AccessSettingsJSON, new.AccessSettingsJSON) ||
		!bytes.Equal(old.ApplicationSettingsJSON, new.ApplicationSettingsJSON) {
		diff.IsChanged = true
	}

	return diff
}
