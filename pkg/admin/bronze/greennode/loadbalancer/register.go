package loadbalancer

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
	plb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer/bronzegreennodeloadbalancerlb"
	"danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer/predicate"
)

// Register registers all GreenNode Load Balancer admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entlb.NewClient(
		entlb.Driver(driver),
		entlb.AlternateSchema(entlb.DefaultSchemaConfig()),
	)

	filterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "greennode_loadbalancer_lbs",
		Columns: []string{"status", "region", "location", "type"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/greennode/loadbalancer/lbs",
		Nav:    &admin.NavMeta{Label: "Load Balancers", Group: []string{"Bronze", "GreenNode", "Load Balancer"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "lbs",
			AllowedFields: map[string]bool{
				"name": true, "q": true, "status": true, "region": true,
				"location": true, "project_id": true, "type": true,
				"address": true, "total_nodes": true, "created_at_api": true,
				"first_collected_at": true, "collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeGreenNodeLoadBalancerLB.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeGreenNodeLoadBalancerLB](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[plb.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "q", Kind: lh.Search, Pred: lh.Pred(plb.NameContainsFold)},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(plb.StatusIn), EqFn: lh.Pred(plb.StatusEQ)},
				{Field: "region", Kind: lh.Multi, InFn: lh.PredIn(plb.RegionIn), EqFn: lh.Pred(plb.RegionEQ)},
				{Field: "location", Kind: lh.Multi, InFn: lh.PredIn(plb.LocationIn), EqFn: lh.Pred(plb.LocationEQ)},
				{Field: "project_id", Kind: lh.Exact, Pred: lh.Pred(plb.ProjectIDEQ)},
				{Field: "type", Kind: lh.Multi, InFn: lh.PredIn(plb.TypeIn), EqFn: lh.Pred(plb.TypeEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":               lh.Sort(plb.ByName),
				"status":             lh.Sort(plb.ByStatus),
				"address":            lh.Sort(plb.ByAddress),
				"type":               lh.Sort(plb.ByType),
				"location":           lh.Sort(plb.ByLocation),
				"region":             lh.Sort(plb.ByRegion),
				"project_id":         lh.Sort(plb.ByProjectID),
				"total_nodes":        lh.Sort(plb.ByTotalNodes),
				"created_at_api":     lh.Sort(plb.ByCreatedAtAPI),
				"collected_at":       lh.Sort(plb.ByCollectedAt),
				"first_collected_at": lh.Sort(plb.ByFirstCollectedAt),
			},
			DefaultOrder:  plb.ByCollectedAt(entsql.OrderDesc()),
			FilterOptions: filterOpts,
		}),
	})

	lh.RegisterSQL(db, sqlTables)
}

func bronzeGNLB(api, table, label string, columns []string, filters []lh.SQLFilterDef, filterOptionCols []string) lh.SQLTable {
	return lh.SQLTable{
		API:                 "/api/v1/bronze/greennode/loadbalancer/" + api,
		Schema:              "bronze",
		Table:               table,
		Nav:                 admin.NavMeta{Label: label, Group: []string{"Bronze", "GreenNode", "Load Balancer"}},
		Columns:             columns,
		Filters:             filters,
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: filterOptionCols,
	}
}

var sqlTables = []lh.SQLTable{
	bronzeGNLB("certificates", "greennode_loadbalancer_certificates", "Certificates",
		[]string{"resource_id", "name", "certificate_type", "expired_at", "imported_at", "not_after", "key_algorithm", "serial", "subject", "domain_name", "in_use", "issuer", "signature_algorithm", "not_before", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "certificate_type", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"certificate_type", "region", "project_id"},
	),
	bronzeGNLB("packages", "greennode_loadbalancer_packages", "Packages",
		[]string{"resource_id", "name", "type", "connection_number", "data_transfer", "mode", "lb_type", "display_lb_type", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "type", Kind: lh.Multi},
			{Column: "lb_type", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"type", "lb_type", "region", "project_id"},
	),
}
