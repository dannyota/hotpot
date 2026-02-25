package interconnect

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InterconnectDiff represents changes between old and new interconnect states.
type InterconnectDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffInterconnectData compares old Ent entity and new InterconnectData.
func DiffInterconnectData(old *ent.BronzeGreenNodeNetworkInterconnect, new *InterconnectData) *InterconnectDiff {
	if old == nil {
		return &InterconnectDiff{IsNew: true}
	}

	return &InterconnectDiff{
		IsChanged: old.Name != new.Name ||
			old.Description != new.Description ||
			old.Status != new.Status ||
			old.EnableGw2 != new.EnableGw2 ||
			old.Gw01IP != new.Gw01IP ||
			old.Gw02IP != new.Gw02IP ||
			old.GwVip != new.GwVIP ||
			old.RemoteGw01IP != new.RemoteGw01IP ||
			old.RemoteGw02IP != new.RemoteGw02IP ||
			old.PackageID != new.PackageID ||
			old.TypeID != new.TypeID,
	}
}

// HasAnyChange returns true if the interconnect changed.
func (d *InterconnectDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
