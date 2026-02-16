package organization

import (
	"time"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
)

// OrganizationData holds converted organization data ready for Ent insertion.
type OrganizationData struct {
	ID                  string
	Name                string
	DisplayName         string
	State               string
	DirectoryCustomerID string
	Etag                string
	CreateTime          string
	UpdateTime          string
	DeleteTime          string
	CollectedAt         time.Time
}

// ConvertOrganization converts a GCP API Organization to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertOrganization(org *resourcemanagerpb.Organization, collectedAt time.Time) *OrganizationData {
	data := &OrganizationData{
		ID:                  org.GetName(),
		Name:                org.GetName(),
		DisplayName:         org.GetDisplayName(),
		State:               org.GetState().String(),
		DirectoryCustomerID: org.GetDirectoryCustomerId(),
		Etag:                org.GetEtag(),
		CollectedAt:         collectedAt,
	}

	// Convert timestamps
	if org.CreateTime != nil {
		data.CreateTime = org.CreateTime.AsTime().Format(time.RFC3339)
	}
	if org.UpdateTime != nil {
		data.UpdateTime = org.UpdateTime.AsTime().Format(time.RFC3339)
	}
	if org.DeleteTime != nil {
		data.DeleteTime = org.DeleteTime.AsTime().Format(time.RFC3339)
	}

	return data
}
