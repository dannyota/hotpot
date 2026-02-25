package routetable

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// RouteTableDiff represents changes between old and new route table states.
type RouteTableDiff struct {
	IsNew     bool
	IsChanged bool

	RoutesDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffRouteTableData compares old Ent entity and new RouteTableData.
func DiffRouteTableData(old *ent.BronzeGreenNodeNetworkRouteTable, new *RouteTableData) *RouteTableDiff {
	if old == nil {
		return &RouteTableDiff{
			IsNew:      true,
			RoutesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &RouteTableDiff{}
	diff.IsChanged = hasRouteTableFieldsChanged(old, new)
	diff.RoutesDiff = diffRoutes(old.Edges.Routes, new.Routes)

	return diff
}

// HasAnyChange returns true if any part of the route table changed.
func (d *RouteTableDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.RoutesDiff.Changed
}

func hasRouteTableFieldsChanged(old *ent.BronzeGreenNodeNetworkRouteTable, new *RouteTableData) bool {
	return old.Name != new.Name ||
		old.Status != new.Status ||
		old.NetworkID != new.NetworkID
}

func diffRoutes(old []*ent.BronzeGreenNodeNetworkRouteTableRoute, new []RouteData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*ent.BronzeGreenNodeNetworkRouteTableRoute)
	for _, r := range old {
		oldMap[r.RouteID] = r
	}
	for _, r := range new {
		oldRoute, ok := oldMap[r.RouteID]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if oldRoute.RoutingType != r.RoutingType ||
			oldRoute.DestinationCidrBlock != r.DestinationCidrBlock ||
			oldRoute.Target != r.Target ||
			oldRoute.Status != r.Status {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
