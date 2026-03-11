package volumetypezone

import (
	"encoding/json"
	"fmt"
	"time"

	volumev1 "danny.vn/gnode/services/volume/v1"
)

// VolumeTypeZoneData represents a converted volume type zone ready for Ent insertion.
type VolumeTypeZoneData struct {
	ID           string
	Name         string
	PoolNameJSON json.RawMessage
	Region       string
	ProjectID    string
	CollectedAt  time.Time
}

// ConvertVolumeTypeZone converts a GreenNode SDK VolumeTypeZone to VolumeTypeZoneData.
func ConvertVolumeTypeZone(z *volumev1.VolumeTypeZone, projectID, region string, collectedAt time.Time) (*VolumeTypeZoneData, error) {
	data := &VolumeTypeZoneData{
		ID:          z.ID,
		Name:        z.Name,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if len(z.PoolName) > 0 {
		poolJSON, err := json.Marshal(z.PoolName)
		if err != nil {
			return nil, fmt.Errorf("marshal pool names for zone %s: %w", z.ID, err)
		}
		data.PoolNameJSON = poolJSON
	}

	return data, nil
}
