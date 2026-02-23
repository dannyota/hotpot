package quota

import (
	"fmt"
	"time"

	portalv2 "danny.vn/greennode/services/portal/v2"
)

// QuotaData represents a converted quota ready for Ent insertion.
type QuotaData struct {
	ID          string
	Name        string
	Description string
	Type        string
	LimitValue  int
	UsedValue   int
	ProjectID   string
	CollectedAt time.Time
}

// ConvertQuota converts a GreenNode SDK Quota to QuotaData.
func ConvertQuota(q *portalv2.Quota, projectID string, collectedAt time.Time) *QuotaData {
	return &QuotaData{
		ID:          fmt.Sprintf("%s:%s", projectID, q.Name),
		Name:        q.Name,
		Description: q.Description,
		Type:        q.Type,
		LimitValue:  q.Limit,
		UsedValue:   q.Used,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
