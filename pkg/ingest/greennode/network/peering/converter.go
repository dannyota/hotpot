package peering

import (
	"time"

	networkv2 "danny.vn/greennode/services/network/v2"
)

// PeeringData represents a converted peering ready for Ent insertion.
type PeeringData struct {
	UUID        string
	Name        string
	Status      string
	FromVpcID   string
	FromCidr    string
	EndVpcID    string
	EndCidr     string
	CreatedAt   string
	Region      string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertPeering converts a GreenNode SDK Peering to PeeringData.
func ConvertPeering(p *networkv2.Peering, projectID, region string, collectedAt time.Time) *PeeringData {
	return &PeeringData{
		UUID:        p.UUID,
		Name:        p.Name,
		Status:      p.Status,
		FromVpcID:   p.FromVpcID,
		FromCidr:    p.FromCidr,
		EndVpcID:    p.EndVpcID,
		EndCidr:     p.EndCidr,
		CreatedAt:   p.CreatedAt,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
