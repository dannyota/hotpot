package volume

import (
	"time"

	"github.com/digitalocean/godo"
)

// VolumeData holds converted Volume data ready for Ent insertion.
type VolumeData struct {
	ResourceID      string
	Name            string
	Region          string
	SizeGigabytes   int64
	Description     string
	DropletIdsJSON  []int
	FilesystemType  string
	FilesystemLabel string
	TagsJSON        []string
	APICreatedAt    *time.Time
	CollectedAt     time.Time
}

// ConvertVolume converts a godo Volume to VolumeData.
func ConvertVolume(v godo.Volume, collectedAt time.Time) *VolumeData {
	data := &VolumeData{
		ResourceID:      v.ID,
		Name:            v.Name,
		SizeGigabytes:   v.SizeGigaBytes,
		Description:     v.Description,
		DropletIdsJSON:  v.DropletIDs,
		FilesystemType:  v.FilesystemType,
		FilesystemLabel: v.FilesystemLabel,
		TagsJSON:        v.Tags,
		CollectedAt:     collectedAt,
	}

	if v.Region != nil {
		data.Region = v.Region.Slug
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	return data
}
