package lbpackage

import (
	"time"

	lbv2 "danny.vn/gnode/services/loadbalancer/v2"
)

// PackageData represents a converted load balancer package ready for Ent insertion.
type PackageData struct {
	ID               string
	Name             string
	Type             string
	ConnectionNumber int
	DataTransfer     int
	Mode             string
	LbType           string
	DisplayLbType    string
	Region           string
	ProjectID        string
	CollectedAt      time.Time
}

// ConvertPackage converts a GreenNode SDK LoadBalancerPackage to PackageData.
func ConvertPackage(p *lbv2.LoadBalancerPackage, projectID, region string, collectedAt time.Time) *PackageData {
	return &PackageData{
		ID:               p.UUID,
		Name:             p.Name,
		Type:             p.Type,
		ConnectionNumber: p.ConnectionNumber,
		DataTransfer:     p.DataTransfer,
		Mode:             p.Mode,
		LbType:           p.LbType,
		DisplayLbType:    p.DisplayLbType,
		Region:           region,
		ProjectID:        projectID,
		CollectedAt:      collectedAt,
	}
}
