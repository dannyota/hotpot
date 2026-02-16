package droplet

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// DropletDiff represents changes between old and new Droplet states.
type DropletDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffDropletData compares old Ent entity and new data.
func DiffDropletData(old *ent.BronzeDODroplet, new *DropletData) *DropletDiff {
	if old == nil {
		return &DropletDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.Memory != new.Memory ||
		old.Vcpus != new.Vcpus ||
		old.Disk != new.Disk ||
		old.Region != new.Region ||
		old.SizeSlug != new.SizeSlug ||
		old.Status != new.Status ||
		old.Locked != new.Locked ||
		old.VpcUUID != new.VpcUUID ||
		old.APICreatedAt != new.APICreatedAt ||
		!bytes.Equal(old.ImageJSON, new.ImageJSON) ||
		!bytes.Equal(old.SizeJSON, new.SizeJSON) ||
		!bytes.Equal(old.NetworksJSON, new.NetworksJSON) ||
		!bytes.Equal(old.KernelJSON, new.KernelJSON) ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!bytes.Equal(old.FeaturesJSON, new.FeaturesJSON) ||
		!bytes.Equal(old.VolumeIdsJSON, new.VolumeIdsJSON) ||
		!bytes.Equal(old.BackupIdsJSON, new.BackupIdsJSON) ||
		!bytes.Equal(old.SnapshotIdsJSON, new.SnapshotIdsJSON)

	return &DropletDiff{IsChanged: changed}
}
