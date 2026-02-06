package targetinstance

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertTargetInstance converts a GCP API TargetInstance to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertTargetInstance(ti *computepb.TargetInstance, projectID string, collectedAt time.Time) bronze.GCPComputeTargetInstance {
	return bronze.GCPComputeTargetInstance{
		ResourceID:        fmt.Sprintf("%d", ti.GetId()),
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
