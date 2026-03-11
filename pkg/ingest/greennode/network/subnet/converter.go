package subnet

import (
	"encoding/json"
	"time"

	networkv2 "danny.vn/gnode/services/network/v2"
)

// SubnetData represents a converted subnet ready for Ent insertion.
type SubnetData struct {
	UUID                   string
	Name                   string
	NetworkID              string
	Cidr                   string
	Status                 string
	RouteTableID           string
	InterfaceAclPolicyID   string
	InterfaceAclPolicyName string
	ZoneID                 string
	SecondarySubnets       json.RawMessage
	Region                 string
	ProjectID              string
	CollectedAt            time.Time
}

// ConvertSubnet converts a GreenNode SDK Subnet to SubnetData.
func ConvertSubnet(s *networkv2.Subnet, projectID, region string, collectedAt time.Time) *SubnetData {
	data := &SubnetData{
		UUID:                   s.UUID,
		Name:                   s.Name,
		NetworkID:              s.NetworkID,
		Cidr:                   s.Cidr,
		Status:                 s.Status,
		RouteTableID:           s.RouteTableID,
		InterfaceAclPolicyID:   s.InterfaceAclPolicyID,
		InterfaceAclPolicyName: s.InterfaceAclPolicyName,
		ZoneID:                 s.ZoneID,
		Region:                 region,
		ProjectID:              projectID,
		CollectedAt:            collectedAt,
	}

	if len(s.SecondarySubnets) > 0 {
		raw, err := json.Marshal(s.SecondarySubnets)
		if err == nil {
			data.SecondarySubnets = raw
		}
	}

	return data
}
