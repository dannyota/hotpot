package endpoint

import (
	"time"

	networkv1 "danny.vn/greennode/services/network/v1"
)

// EndpointData represents a converted endpoint ready for Ent insertion.
type EndpointData struct {
	UUID                string
	Name                string
	Ipv4Address         string
	EndpointURL         string
	EndpointAuthURL     string
	EndpointServiceID   string
	Status              string
	BillingStatus       string
	EndpointType        string
	Version             string
	Description         string
	CreatedAt           string
	UpdatedAt           string
	VpcID               string
	VpcName             string
	ZoneUuid            string
	EnableDnsName       bool
	EndpointDomains     []string
	SubnetID            string
	CategoryName        string
	ServiceName         string
	ServiceEndpointType string
	PackageName         string
	Region              string
	ProjectID           string
	CollectedAt         time.Time
}

// ConvertEndpoint converts a GreenNode SDK Endpoint to EndpointData.
func ConvertEndpoint(e *networkv1.Endpoint, projectID, region string, collectedAt time.Time) *EndpointData {
	data := &EndpointData{
		UUID:              e.UUID,
		Name:              e.Name,
		Ipv4Address:       e.IPv4Address,
		EndpointURL:       e.EndpointURL,
		EndpointAuthURL:   e.EndpointAuthURL,
		EndpointServiceID: e.EndpointServiceID,
		Status:            e.Status,
		BillingStatus:     e.BillingStatus,
		EndpointType:      e.EndpointType,
		Version:           e.Version,
		Description:       e.Description,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
		VpcID:             e.VpcID,
		ZoneUuid:          e.ZoneUuid,
		EnableDnsName:     e.EnableDnsName,
		EndpointDomains:   e.EndpointDomains,
		Region:            region,
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	if e.VPC != nil {
		data.VpcName = e.VPC.Name
	}
	if e.Subnet != nil {
		data.SubnetID = e.Subnet.UUID
	}
	if e.Category != nil {
		data.CategoryName = e.Category.Name
	}
	if e.Service != nil {
		data.ServiceName = e.Service.Name
		data.ServiceEndpointType = e.Service.EndpointType
	}
	if e.Package != nil {
		data.PackageName = e.Package.Name
	}

	return data
}
