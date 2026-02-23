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
	Region      string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertQuota converts a GreenNode SDK Quota to QuotaData.
func ConvertQuota(q *portalv2.Quota, projectID, region string, collectedAt time.Time) *QuotaData {
	return &QuotaData{
		ID:          fmt.Sprintf("%s:%s:%s", projectID, region, q.Name),
		Name:        q.Name,
		Description: q.Description,
		Type:        q.Type,
		LimitValue:  q.Limit,
		UsedValue:   q.Used,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
