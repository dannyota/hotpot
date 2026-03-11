package routetable

import (
	"time"

	networkv2 "danny.vn/gnode/services/network/v2"
)

// RouteTableData represents a converted route table ready for Ent insertion.
type RouteTableData struct {
	UUID        string
	Name        string
	Status      string
	NetworkID   string
	CreatedAt   string
	Region      string
	ProjectID   string
	CollectedAt time.Time

	Routes []RouteData
}

// RouteData represents a route within a route table ready for Ent insertion.
type RouteData struct {
	RouteID              string
	RouteTableID         string
	RoutingType          string
	DestinationCidrBlock string
	Target               string
	Status               string
}

// ConvertRouteTable converts a GreenNode SDK RouteTable to RouteTableData.
// Routes are embedded in the RouteTable response — no separate API call needed.
func ConvertRouteTable(rt *networkv2.RouteTable, projectID, region string, collectedAt time.Time) *RouteTableData {
	data := &RouteTableData{
		UUID:        rt.UUID,
		Name:        rt.Name,
		Status:      rt.Status,
		NetworkID:   rt.NetworkID,
		CreatedAt:   rt.CreatedAt,
		Region:      region,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	data.Routes = ConvertRoutes(rt.Routes)

	return data
}

// ConvertRoutes converts SDK routes to RouteData.
func ConvertRoutes(routes []*networkv2.Route) []RouteData {
	if len(routes) == 0 {
		return nil
	}
	result := make([]RouteData, 0, len(routes))
	for _, r := range routes {
		result = append(result, RouteData{
			RouteID:              r.UUID,
			RouteTableID:         r.RouteTableID,
			RoutingType:          r.RoutingType,
			DestinationCidrBlock: r.DestinationCidrBlock,
			Target:               r.Target,
			Status:               r.Status,
		})
	}
	return result
}
