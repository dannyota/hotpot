package zone

import (
	"time"

	portalv1 "danny.vn/greennode/services/portal/v1"
)

// ZoneData represents a converted zone ready for Ent insertion.
type ZoneData struct {
	ID            string
	Name          string
	OpenstackZone string
	ProjectID     string
	CollectedAt   time.Time
}

// ConvertZone converts a GreenNode SDK Zone to ZoneData.
func ConvertZone(z *portalv1.Zone, projectID string, collectedAt time.Time) *ZoneData {
	return &ZoneData{
		ID:            z.Uuid,
		Name:          z.Name,
		OpenstackZone: z.OpenstackZone,
		ProjectID:     projectID,
		CollectedAt:   collectedAt,
	}
}
