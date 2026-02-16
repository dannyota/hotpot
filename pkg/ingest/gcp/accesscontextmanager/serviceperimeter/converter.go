package serviceperimeter

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/accesscontextmanager/apiv1/accesscontextmanagerpb"
)

// ServicePerimeterData holds converted service perimeter data ready for Ent insertion.
type ServicePerimeterData struct {
	ID                     string
	Title                  string
	Description            string
	PerimeterType          int
	Etag                   string
	UseExplicitDryRunSpec  bool
	StatusJSON             json.RawMessage
	SpecJSON               json.RawMessage
	AccessPolicyName       string
	OrganizationID         string
	CollectedAt            time.Time
}

// ConvertServicePerimeter converts a raw GCP API service perimeter to Ent-compatible data.
func ConvertServicePerimeter(orgName string, policyName string, perimeter *accesscontextmanagerpb.ServicePerimeter, collectedAt time.Time) *ServicePerimeterData {
	data := &ServicePerimeterData{
		ID:                    perimeter.GetName(),
		Title:                 perimeter.GetTitle(),
		Description:           perimeter.GetDescription(),
		PerimeterType:         int(perimeter.GetPerimeterType()),
		UseExplicitDryRunSpec: perimeter.GetUseExplicitDryRunSpec(),
		AccessPolicyName:      policyName,
		OrganizationID:        orgName,
		CollectedAt:           collectedAt,
	}

	if status := perimeter.GetStatus(); status != nil {
		data.StatusJSON = servicePerimeterConfigToJSON(status)
	}
	if spec := perimeter.GetSpec(); spec != nil {
		data.SpecJSON = servicePerimeterConfigToJSON(spec)
	}

	return data
}
