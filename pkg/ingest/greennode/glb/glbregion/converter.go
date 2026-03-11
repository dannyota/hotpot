package glbregion

import (
	"time"

	glbv1 "danny.vn/gnode/services/glb/v1"
)

// GLBRegionData represents a converted global region ready for Ent insertion.
type GLBRegionData struct {
	ID               string
	Name             string
	VserverEndpoint  string
	VlbEndpoint      string
	UIServerEndpoint string
	ProjectID        string
	CollectedAt      time.Time
}

// ConvertGLBRegion converts a GreenNode SDK GlobalRegion to GLBRegionData.
func ConvertGLBRegion(r *glbv1.GlobalRegion, projectID string, collectedAt time.Time) *GLBRegionData {
	return &GLBRegionData{
		ID:               r.ID,
		Name:             r.Name,
		VserverEndpoint:  r.VServerEndpoint,
		VlbEndpoint:      r.VlbEndpoint,
		UIServerEndpoint: r.UIServerEndpoint,
		ProjectID:        projectID,
		CollectedAt:      collectedAt,
	}
}
