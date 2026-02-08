package targetinstance

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetInstanceData holds converted target instance data ready for Ent insertion.
type TargetInstanceData struct {
	ID                string
	Name              string
	Description       string
	Zone              string
	Instance          string
	Network           string
	NatPolicy         string
	SecurityPolicy    string
	SelfLink          string
	CreationTimestamp string
	ProjectID         string
	CollectedAt       time.Time
}

// ConvertTargetInstance converts a GCP API TargetInstance to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertTargetInstance(ti *computepb.TargetInstance, projectID string, collectedAt time.Time) *TargetInstanceData {
	return &TargetInstanceData{
		ID:                fmt.Sprintf("%d", ti.GetId()),
		Name:              ti.GetName(),
		Description:       ti.GetDescription(),
		Zone:              ti.GetZone(),
		Instance:          ti.GetInstance(),
		Network:           ti.GetNetwork(),
		NatPolicy:         ti.GetNatPolicy(),
		SecurityPolicy:    ti.GetSecurityPolicy(),
		SelfLink:          ti.GetSelfLink(),
		CreationTimestamp: ti.GetCreationTimestamp(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}
}
