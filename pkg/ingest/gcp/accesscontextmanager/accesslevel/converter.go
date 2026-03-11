package accesslevel

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/accesscontextmanager/apiv1/accesscontextmanagerpb"
)

// AccessLevelData holds converted access level data ready for Ent insertion.
type AccessLevelData struct {
	ID               string
	Title            string
	Description      string
	BasicJSON        json.RawMessage
	CustomJSON       json.RawMessage
	AccessPolicyName string
	OrganizationID   string
	CollectedAt      time.Time
}

// ConvertAccessLevel converts a raw GCP API access level to Ent-compatible data.
func ConvertAccessLevel(orgName string, policyName string, level *accesscontextmanagerpb.AccessLevel, collectedAt time.Time) *AccessLevelData {
	data := &AccessLevelData{
		ID:               level.GetName(),
		Title:            level.GetTitle(),
		Description:      level.GetDescription(),
		AccessPolicyName: policyName,
		OrganizationID:   orgName,
		CollectedAt:      collectedAt,
	}

	if basic := level.GetBasic(); basic != nil {
		data.BasicJSON = basicLevelToJSON(basic)
	}
	if custom := level.GetCustom(); custom != nil {
		data.CustomJSON = customLevelToJSON(custom)
	}

	return data
}
