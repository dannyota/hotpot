package vpc

import (
	"time"

	networkv2 "danny.vn/greennode/services/network/v2"
)

// VPCData represents a converted VPC ready for Ent insertion.
type VPCData struct {
	UUID           string
	Name           string
	Cidr           string
	Status         string
	RouteTableID   string
	RouteTableName string
	DhcpOptionID   string
	DhcpOptionName string
	DnsStatus      string
	DnsID          string
	ZoneUuid       string
	ZoneName       string
	CreatedAt      string
	ElasticIps     []string
	Region         string
	ProjectID      string
	CollectedAt    time.Time
}

// ConvertVPC converts a GreenNode SDK Network to VPCData.
func ConvertVPC(n *networkv2.Network, projectID, region string, collectedAt time.Time) *VPCData {
	data := &VPCData{
		UUID:           n.UUID,
		Name:           n.Name,
		Cidr:           n.Cidr,
		Status:         n.Status,
		RouteTableID:   n.RouteTableID,
		RouteTableName: n.RouteTableName,
		DhcpOptionID:   n.DhcpOptionID,
		DhcpOptionName: n.DhcpOptionName,
		DnsStatus:      n.DnsStatus,
		DnsID:          n.DnsID,
		CreatedAt:      n.CreatedAt,
		ElasticIps:     n.ElasticIps,
		Region:         region,
		ProjectID:      projectID,
		CollectedAt:    collectedAt,
	}

	data.ZoneUuid = n.Zone.Uuid
	data.ZoneName = n.Zone.Name

	return data
}
