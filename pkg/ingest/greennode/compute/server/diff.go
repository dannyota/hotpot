package server

import (
	"bytes"

	entcompute "danny.vn/hotpot/pkg/storage/ent/greennode/compute"
)

// ServerDiff represents changes between old and new server states.
type ServerDiff struct {
	IsNew     bool
	IsChanged bool

	SecGroupsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffServerData compares old Ent entity and new ServerData.
func DiffServerData(old *entcompute.BronzeGreenNodeComputeServer, new *ServerData) *ServerDiff {
	if old == nil {
		return &ServerDiff{
			IsNew:         true,
			SecGroupsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &ServerDiff{}
	diff.IsChanged = hasServerFieldsChanged(old, new)
	diff.SecGroupsDiff = diffSecGroups(old.Edges.SecGroups, new.SecGroups)

	return diff
}

// HasAnyChange returns true if any part of the server changed.
func (d *ServerDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.SecGroupsDiff.Changed
}

func hasServerFieldsChanged(old *entcompute.BronzeGreenNodeComputeServer, new *ServerData) bool {
	return old.Name != new.Name ||
		old.Status != new.Status ||
		old.Location != new.Location ||
		old.ZoneID != new.ZoneID ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		old.BootVolumeID != new.BootVolumeID ||
		old.EncryptionVolume != new.EncryptionVolume ||
		old.Licence != new.Licence ||
		old.Metadata != new.Metadata ||
		old.MigrateState != new.MigrateState ||
		old.Product != new.Product ||
		old.ServerGroupID != new.ServerGroupID ||
		old.ServerGroupName != new.ServerGroupName ||
		old.SSHKeyName != new.SSHKeyName ||
		old.StopBeforeMigrate != new.StopBeforeMigrate ||
		old.User != new.User ||
		old.ImageID != new.ImageID ||
		old.ImageType != new.ImageType ||
		old.ImageVersion != new.ImageVersion ||
		old.FlavorID != new.FlavorID ||
		old.FlavorName != new.FlavorName ||
		old.FlavorCPU != new.FlavorCPU ||
		old.FlavorMemory != new.FlavorMemory ||
		old.FlavorGpu != new.FlavorGPU ||
		old.FlavorBandwidth != new.FlavorBandwidth ||
		!bytes.Equal(old.InterfacesJSON, new.InterfacesJSON)
}

func diffSecGroups(old []*entcompute.BronzeGreenNodeComputeServerSecGroup, new []SecGroupData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]string)
	for _, sg := range old {
		oldMap[sg.UUID] = sg.Name
	}
	for _, sg := range new {
		if name, ok := oldMap[sg.UUID]; !ok || name != sg.Name {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
