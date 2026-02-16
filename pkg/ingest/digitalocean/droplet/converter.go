package droplet

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
)

// DropletData holds converted Droplet data ready for Ent insertion.
type DropletData struct {
	ResourceID      string
	Name            string
	Memory          int
	Vcpus           int
	Disk            int
	Region          string
	SizeSlug        string
	Status          string
	Locked          bool
	VpcUUID         string
	APICreatedAt    string
	ImageJSON       json.RawMessage
	SizeJSON        json.RawMessage
	NetworksJSON    json.RawMessage
	KernelJSON      json.RawMessage
	TagsJSON        json.RawMessage
	FeaturesJSON    json.RawMessage
	VolumeIdsJSON   json.RawMessage
	BackupIdsJSON   json.RawMessage
	SnapshotIdsJSON json.RawMessage
	CollectedAt     time.Time
}

// ConvertDroplet converts a godo Droplet to DropletData.
func ConvertDroplet(v godo.Droplet, collectedAt time.Time) *DropletData {
	data := &DropletData{
		ResourceID:   strconv.Itoa(v.ID),
		Name:         v.Name,
		Memory:       v.Memory,
		Vcpus:        v.Vcpus,
		Disk:         v.Disk,
		SizeSlug:     v.SizeSlug,
		Status:       v.Status,
		Locked:       v.Locked,
		VpcUUID:      v.VPCUUID,
		APICreatedAt: v.Created,
		CollectedAt:  collectedAt,
	}

	if v.Region != nil {
		data.Region = v.Region.Slug
	}

	if v.Image != nil {
		data.ImageJSON, _ = json.Marshal(v.Image)
	}

	if v.Size != nil {
		data.SizeJSON, _ = json.Marshal(v.Size)
	}

	if v.Networks != nil {
		data.NetworksJSON, _ = json.Marshal(v.Networks)
	}

	if v.Kernel != nil {
		data.KernelJSON, _ = json.Marshal(v.Kernel)
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if len(v.Features) > 0 {
		data.FeaturesJSON, _ = json.Marshal(v.Features)
	}

	if len(v.VolumeIDs) > 0 {
		data.VolumeIdsJSON, _ = json.Marshal(v.VolumeIDs)
	}

	if len(v.BackupIDs) > 0 {
		data.BackupIdsJSON, _ = json.Marshal(v.BackupIDs)
	}

	if len(v.SnapshotIDs) > 0 {
		data.SnapshotIdsJSON, _ = json.Marshal(v.SnapshotIDs)
	}

	return data
}
