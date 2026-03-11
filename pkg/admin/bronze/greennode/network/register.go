package network

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	entnetwork "danny.vn/hotpot/pkg/storage/ent/greennode/network"
	psec "danny.vn/hotpot/pkg/storage/ent/greennode/network/bronzegreennodenetworksecgroup"
	pvpc "danny.vn/hotpot/pkg/storage/ent/greennode/network/bronzegreennodenetworkvpc"
	"danny.vn/hotpot/pkg/storage/ent/greennode/network/predicate"
)

// Register registers all GreenNode Network admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entnetwork.NewClient(
		entnetwork.Driver(driver),
		entnetwork.AlternateSchema(entnetwork.DefaultSchemaConfig()),
	)

	// VPCs
	vpcFilterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "greennode_network_vpcs",
		Columns: []string{"status", "region"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/greennode/network/vpcs",
		Nav:    &admin.NavMeta{Label: "VPCs", Group: []string{"Bronze", "GreenNode", "Network"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "vpcs",
			AllowedFields: map[string]bool{
				"id": true, "name": true, "q": true, "status": true, "region": true,
				"project_id": true, "cidr": true,
				"first_collected_at": true, "collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeGreenNodeNetworkVpc.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeGreenNodeNetworkVpc](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[pvpc.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "q", Kind: lh.Search, Pred: lh.Pred(pvpc.NameContainsFold)},
				{Field: "id", Kind: lh.Multi, InFn: lh.PredIn(pvpc.IDIn), EqFn: lh.Pred(pvpc.IDEQ)},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(pvpc.StatusIn), EqFn: lh.Pred(pvpc.StatusEQ)},
				{Field: "region", Kind: lh.Multi, InFn: lh.PredIn(pvpc.RegionIn), EqFn: lh.Pred(pvpc.RegionEQ)},
				{Field: "project_id", Kind: lh.Exact, Pred: lh.Pred(pvpc.ProjectIDEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":               lh.Sort(pvpc.ByName),
				"status":             lh.Sort(pvpc.ByStatus),
				"cidr":               lh.Sort(pvpc.ByCidr),
				"region":             lh.Sort(pvpc.ByRegion),
				"project_id":         lh.Sort(pvpc.ByProjectID),
				"collected_at":       lh.Sort(pvpc.ByCollectedAt),
				"first_collected_at": lh.Sort(pvpc.ByFirstCollectedAt),
			},
			DefaultOrder:  pvpc.ByCollectedAt(entsql.OrderDesc()),
			FilterOptions: vpcFilterOpts,
		}),
	})

	// Security Groups
	secFilterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "greennode_network_secgroups",
		Columns: []string{"status", "region"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/greennode/network/secgroups",
		Nav:    &admin.NavMeta{Label: "Security Groups", Group: []string{"Bronze", "GreenNode", "Network"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "secgroups",
			AllowedFields: map[string]bool{
				"name": true, "q": true, "status": true, "region": true,
				"project_id": true, "description": true,
				"first_collected_at": true, "collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeGreenNodeNetworkSecgroup.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeGreenNodeNetworkSecgroup](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[psec.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "q", Kind: lh.Search, Pred: lh.Pred(psec.NameContainsFold)},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(psec.StatusIn), EqFn: lh.Pred(psec.StatusEQ)},
				{Field: "region", Kind: lh.Multi, InFn: lh.PredIn(psec.RegionIn), EqFn: lh.Pred(psec.RegionEQ)},
				{Field: "project_id", Kind: lh.Exact, Pred: lh.Pred(psec.ProjectIDEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":               lh.Sort(psec.ByName),
				"status":             lh.Sort(psec.ByStatus),
				"description":        lh.Sort(psec.ByDescription),
				"region":             lh.Sort(psec.ByRegion),
				"project_id":         lh.Sort(psec.ByProjectID),
				"collected_at":       lh.Sort(psec.ByCollectedAt),
				"first_collected_at": lh.Sort(psec.ByFirstCollectedAt),
			},
			DefaultOrder:  psec.ByCollectedAt(entsql.OrderDesc()),
			FilterOptions: secFilterOpts,
		}),
	})

	lh.RegisterSQL(db, sqlTables)
}

func bronzeGNNetwork(api, table, label string, columns []string, filters []lh.SQLFilterDef, filterOptionCols []string) lh.SQLTable {
	return lh.SQLTable{
		API:                 "/api/v1/bronze/greennode/network/" + api,
		Schema:              "bronze",
		Table:               table,
		Nav:                 admin.NavMeta{Label: label, Group: []string{"Bronze", "GreenNode", "Network"}},
		Columns:             columns,
		Filters:             filters,
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: filterOptionCols,
	}
}

var sqlTables = []lh.SQLTable{
	bronzeGNNetwork("subnets", "greennode_network_subnets", "Subnets",
		[]string{"resource_id", "name", "network_id", "cidr", "status", "route_table_id", "interface_acl_policy_id", "interface_acl_policy_name", "zone_id", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
			{Column: "network_id", Kind: lh.Exact},
		},
		[]string{"status", "region", "project_id"},
	),
	bronzeGNNetwork("endpoints", "greennode_network_endpoints", "Endpoints",
		[]string{"resource_id", "name", "ipv4_address", "endpoint_url", "endpoint_service_id", "status", "billing_status", "endpoint_type", "version", "description", "created_at", "updated_at", "vpc_id", "vpc_name", "zone_uuid", "category_name", "service_name", "service_endpoint_type", "package_name", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "billing_status", Kind: lh.Multi},
			{Column: "endpoint_type", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "billing_status", "endpoint_type", "region", "project_id"},
	),
	bronzeGNNetwork("interconnects", "greennode_network_interconnects", "Interconnects",
		[]string{"resource_id", "name", "description", "status", "enable_gw2", "circuit_id", "gw01_ip", "gw02_ip", "gw_vip", "remote_gw01_ip", "remote_gw02_ip", "package_id", "type_id", "type_name", "created_at", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "type_name", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "type_name", "region", "project_id"},
	),
	bronzeGNNetwork("peerings", "greennode_network_peerings", "Peerings",
		[]string{"resource_id", "name", "status", "from_vpc_id", "from_cidr", "end_vpc_id", "end_cidr", "created_at", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "region", "project_id"},
	),
	bronzeGNNetwork("route-tables", "greennode_network_route_tables", "Route Tables",
		[]string{"resource_id", "name", "status", "network_id", "created_at", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "region", "project_id"},
	),
}
