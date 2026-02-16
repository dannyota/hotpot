package accesspolicy

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/accesscontextmanager/apiv1/accesscontextmanagerpb"
)

// AccessPolicyData holds converted access policy data ready for Ent insertion.
type AccessPolicyData struct {
	ID             string
	Parent         string
	Title          string
	Etag           string
	ScopesJSON     json.RawMessage
	OrganizationID string
	CollectedAt    time.Time
}

// ConvertAccessPolicy converts a raw GCP API access policy to Ent-compatible data.
func ConvertAccessPolicy(orgName string, policy *accesscontextmanagerpb.AccessPolicy, collectedAt time.Time) *AccessPolicyData {
	return &AccessPolicyData{
		ID:             policy.GetName(),
		Parent:         policy.GetParent(),
		Title:          policy.GetTitle(),
		Etag:           policy.GetEtag(),
		ScopesJSON:     scopesToJSON(policy.GetScopes()),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}
}
