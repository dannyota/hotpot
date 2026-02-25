package zone

import (
	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
)

// ZoneDiff represents changes between old and new zone states.
type ZoneDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffZoneData compares old Ent entity and new ZoneData.
func DiffZoneData(old *entportal.BronzeGreenNodePortalZone, new *ZoneData) *ZoneDiff {
	if old == nil {
		return &ZoneDiff{IsNew: true}
	}

	return &ZoneDiff{
		IsChanged: old.Name != new.Name ||
			old.OpenstackZone != new.OpenstackZone,
	}
}

// HasAnyChange returns true if the zone changed.
func (d *ZoneDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
