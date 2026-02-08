package serviceaccount

import (
	"encoding/hex"
	"time"

	"cloud.google.com/go/iam/admin/apiv1/adminpb"
)

// ServiceAccountData holds converted service account data ready for Ent insertion.
type ServiceAccountData struct {
	ResourceID     string
	Name           string
	Email          string
	DisplayName    string
	Description    string
	Oauth2ClientId string
	Disabled       bool
	Etag           string
	ProjectID      string
	CollectedAt    time.Time
}

func ConvertServiceAccount(sa *adminpb.ServiceAccount, projectID string, collectedAt time.Time) *ServiceAccountData {
	return &ServiceAccountData{
		ResourceID:     sa.GetUniqueId(),
		Name:           sa.GetName(),
		Email:          sa.GetEmail(),
		DisplayName:    sa.GetDisplayName(),
		Description:    sa.GetDescription(),
		Oauth2ClientId: sa.GetOauth2ClientId(),
		Disabled:       sa.GetDisabled(),
		Etag:           hex.EncodeToString(sa.GetEtag()),
		ProjectID:      projectID,
		CollectedAt:    collectedAt,
	}
}
