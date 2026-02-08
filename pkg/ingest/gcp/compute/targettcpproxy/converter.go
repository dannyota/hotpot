package targettcpproxy

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetTcpProxyData holds converted target TCP proxy data ready for Ent insertion.
type TargetTcpProxyData struct {
	ID                string
	Name              string
	Description       string
	CreationTimestamp string
	SelfLink          string
	Service           string
	ProxyBind         bool
	ProxyHeader       string
	Region            string
	ProjectID         string
	CollectedAt       time.Time
}

// ConvertTargetTcpProxy converts a GCP API TargetTcpProxy to Ent-compatible data.
func ConvertTargetTcpProxy(ttp *computepb.TargetTcpProxy, projectID string, collectedAt time.Time) *TargetTcpProxyData {
	return &TargetTcpProxyData{
		ID:                fmt.Sprintf("%d", ttp.GetId()),
		Name:              ttp.GetName(),
		Description:       ttp.GetDescription(),
		CreationTimestamp: ttp.GetCreationTimestamp(),
		SelfLink:          ttp.GetSelfLink(),
		Service:           ttp.GetService(),
		ProxyBind:         ttp.GetProxyBind(),
		ProxyHeader:       ttp.GetProxyHeader(),
		Region:            ttp.GetRegion(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}
}
