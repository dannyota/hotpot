package volumetype

import (
	"time"

	volumev1 "danny.vn/gnode/services/volume/v1"
)

// VolumeTypeData represents a converted volume type ready for Ent insertion.
type VolumeTypeData struct {
	ID          string
	Name        string
	Iops        int
	MaxSize     int
	MinSize     int
	ThroughPut  int
	ZoneID      string
	Region      string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertVolumeType converts a GreenNode SDK VolumeType to VolumeTypeData.
func ConvertVolumeType(vt *volumev1.VolumeType, projectID, region string, collectedAt time.Time) *VolumeTypeData {
	return &VolumeTypeData{
		ID:          vt.ID,
		Name:        vt.Name,
		Iops:        vt.Iops,
		MaxSize:     vt.MaxSize,
		MinSize:     vt.MinSize,
		ThroughPut:  vt.ThroughPut,
		ZoneID:      vt.ZoneID,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
