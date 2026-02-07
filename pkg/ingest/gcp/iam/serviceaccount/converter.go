package serviceaccount

import (
	"encoding/hex"
	"time"

	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"hotpot/pkg/base/models/bronze"
)

func ConvertServiceAccount(sa *adminpb.ServiceAccount, projectID string, collectedAt time.Time) bronze.GCPIAMServiceAccount {
	return bronze.GCPIAMServiceAccount{
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
