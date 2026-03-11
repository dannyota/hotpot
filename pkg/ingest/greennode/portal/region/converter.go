package region

import (
	"time"

	portalv2 "danny.vn/gnode/services/portal/v2"
)

// RegionData represents a converted region ready for Ent insertion.
type RegionData struct {
	ID          string
	Name        string
	Description string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertRegion converts a GreenNode SDK Region to RegionData.
func ConvertRegion(r *portalv2.Region, projectID string, collectedAt time.Time) *RegionData {
	return &RegionData{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
