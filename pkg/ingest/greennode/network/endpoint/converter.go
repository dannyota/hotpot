package endpoint

import (
	"time"

	networkv1 "danny.vn/greennode/services/network/v1"
)

// EndpointData represents a converted endpoint ready for Ent insertion.
type EndpointData struct {
	ID          string
	Name        string
	Ipv4Address string
	EndpointURL string
	Status      string
	VpcID       string
	Region      string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertEndpoint converts a GreenNode SDK Endpoint to EndpointData.
func ConvertEndpoint(e *networkv1.Endpoint, projectID, region string, collectedAt time.Time) *EndpointData {
	return &EndpointData{
		ID:          e.ID,
		Name:        e.Name,
		Ipv4Address: e.IPv4Address,
		EndpointURL: e.EndpointURL,
		Status:      e.Status,
		VpcID:       e.VpcID,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}
}
