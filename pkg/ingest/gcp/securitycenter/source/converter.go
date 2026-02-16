package source

import (
	"time"

	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
)

// SourceData holds converted SCC source data ready for Ent insertion.
type SourceData struct {
	ID             string
	DisplayName    string
	Description    string
	CanonicalName  string
	OrganizationID string
	CollectedAt    time.Time
}

// ConvertSource converts a raw GCP API SCC source to Ent-compatible data.
func ConvertSource(orgName string, source *securitycenterpb.Source, collectedAt time.Time) *SourceData {
	return &SourceData{
		ID:             source.GetName(),
		DisplayName:    source.GetDisplayName(),
		Description:    source.GetDescription(),
		CanonicalName:  source.GetCanonicalName(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}
}
