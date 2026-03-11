package compute

import (
	"context"
	"database/sql"
	"net/http"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	entcompute "danny.vn/hotpot/pkg/storage/ent/greennode/compute"
	p "danny.vn/hotpot/pkg/storage/ent/greennode/compute/bronzegreennodecomputeserver"
	"danny.vn/hotpot/pkg/storage/ent/greennode/compute/predicate"
)

// Register registers all GreenNode Compute admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entcompute.NewClient(
		entcompute.Driver(driver),
		entcompute.AlternateSchema(entcompute.DefaultSchemaConfig()),
	)

	filterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "greennode_compute_servers",
		Columns: []string{"status", "region", "location", "product", "server_group_name"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/greennode/compute/servers",
		Nav:    &admin.NavMeta{Label: "Servers", Group: []string{"Bronze", "GreenNode", "Compute"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "servers",
			AllowedFields: map[string]bool{
				"name": true, "q": true, "status": true, "location": true, "region": true,
				"project_id": true, "product": true, "flavor_name": true,
				"flavor_cpu": true, "flavor_memory": true, "image_type": true,
				"server_group_name": true,
				"created_at_api": true, "first_collected_at": true, "collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeGreenNodeComputeServer.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeGreenNodeComputeServer](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[p.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "q", Kind: lh.Search, Pred: nameOrIPSearchPred},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(p.StatusIn), EqFn: lh.Pred(p.StatusEQ)},
				{Field: "location", Kind: lh.Multi, InFn: lh.PredIn(p.LocationIn), EqFn: lh.Pred(p.LocationEQ)},
				{Field: "region", Kind: lh.Multi, InFn: lh.PredIn(p.RegionIn), EqFn: lh.Pred(p.RegionEQ)},
				{Field: "project_id", Kind: lh.Exact, Pred: lh.Pred(p.ProjectIDEQ)},
				{Field: "product", Kind: lh.Multi, InFn: lh.PredIn(p.ProductIn), EqFn: lh.Pred(p.ProductEQ)},
				{Field: "server_group_name", Kind: lh.Multi, InFn: lh.PredIn(p.ServerGroupNameIn), EqFn: lh.Pred(p.ServerGroupNameEQ)},
				{Field: "flavor_name", Kind: lh.Search, Pred: lh.Pred(p.FlavorNameContainsFold)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":               lh.Sort(p.ByName),
				"status":             lh.Sort(p.ByStatus),
				"location":           lh.Sort(p.ByLocation),
				"region":             lh.Sort(p.ByRegion),
				"project_id":         lh.Sort(p.ByProjectID),
				"product":            lh.Sort(p.ByProduct),
				"flavor_name":        lh.Sort(p.ByFlavorName),
				"flavor_cpu":         lh.Sort(p.ByFlavorCPU),
				"flavor_memory":      lh.Sort(p.ByFlavorMemory),
				"image_type":         lh.Sort(p.ByImageType),
				"server_group_name": lh.Sort(p.ByServerGroupName),
				"created_at_api":     lh.Sort(p.ByCreatedAtAPI),
				"first_collected_at": lh.Sort(p.ByFirstCollectedAt),
				"collected_at":       lh.Sort(p.ByCollectedAt),
			},
			DefaultOrder:  p.ByCreatedAtAPI(entsql.OrderDesc()),
			FilterOptions: filterOpts,
		}),
	})

	admin.RegisterRoute(admin.RouteRegistration{
		Method:  "GET",
		Path:    "/api/v1/bronze/greennode/compute/servers/stats",
		Handler: serverStatsHandler(db),
	})

	lh.RegisterSQL(db, sqlTables)
}

func nameOrIPSearchPred(v string) lh.Predicate {
	return func(s *entsql.Selector) {
		s.Where(entsql.Or(
			entsql.P(func(b *entsql.Builder) {
				b.WriteString("LOWER(name) LIKE LOWER(")
				b.Arg("%" + v + "%")
				b.WriteByte(')')
			}),
			entsql.P(func(b *entsql.Builder) {
				b.WriteString("CAST(interfaces_json AS TEXT) ILIKE ")
				b.Arg("%" + v + "%")
			}),
		))
	}
}

func serverStatsHandler(db *sql.DB) http.HandlerFunc {
	statsFilters := map[string]admin.StatsFilter{
		"status":            {Column: "status"},
		"location":          {Column: "location"},
		"region":            {Column: "region"},
		"product":           {Column: "product"},
		"server_group_name": {Column: "server_group_name"},
		"project_id":        {Column: "project_id"},
	}

	type statusGroup struct {
		Count     int            `json:"count"`
		Breakdown map[string]int `json:"breakdown"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		where, args := admin.StatsWhere(r, statsFilters)
		query := `SELECT COALESCE(status, '') AS status, COALESCE(region, '') AS region, COUNT(*) AS cnt
			FROM "bronze"."greennode_compute_servers"` + where + ` GROUP BY 1, 2 ORDER BY 1, 3 DESC`
		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			admin.WriteServerError(w, "stats query failed", err)
			return
		}
		defer rows.Close()

		result := map[string]*statusGroup{}
		for rows.Next() {
			var status, region string
			var cnt int
			if err := rows.Scan(&status, &region, &cnt); err != nil {
				admin.WriteServerError(w, "stats scan failed", err)
				return
			}
			g, ok := result[status]
			if !ok {
				g = &statusGroup{Breakdown: map[string]int{}}
				result[status] = g
			}
			g.Count += cnt
			g.Breakdown[region] = cnt
		}
		if err := rows.Err(); err != nil {
			admin.WriteServerError(w, "stats query failed", err)
			return
		}

		admin.WriteJSON(w, http.StatusOK, result)
	}
}

func bronzeGNCompute(api, table, label string, columns []string, filters []lh.SQLFilterDef, filterOptionCols []string) lh.SQLTable {
	return lh.SQLTable{
		API:                 "/api/v1/bronze/greennode/compute/" + api,
		Schema:              "bronze",
		Table:               table,
		Nav:                 admin.NavMeta{Label: label, Group: []string{"Bronze", "GreenNode", "Compute"}},
		Columns:             columns,
		Filters:             filters,
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: filterOptionCols,
	}
}

var sqlTables = []lh.SQLTable{
	bronzeGNCompute("server-groups", "greennode_compute_server_groups", "Server Groups",
		[]string{"resource_id", "name", "description", "policy_id", "policy_name", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"region", "project_id"},
	),
	bronzeGNCompute("ssh-keys", "greennode_compute_ssh_keys", "SSH Keys",
		[]string{"resource_id", "name", "pub_key", "status", "created_at_api", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "region", "project_id"},
	),
	bronzeGNCompute("os-images", "greennode_compute_os_images", "OS Images",
		[]string{"resource_id", "image_type", "image_version", "licence", "license_key", "description", "zone_id", "package_limit_cpu", "package_limit_memory", "package_limit_disk_size", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "description", Kind: lh.Search},
			{Column: "image_type", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"image_type", "region", "project_id"},
	),
	bronzeGNCompute("user-images", "greennode_compute_user_images", "User Images",
		[]string{"resource_id", "name", "status", "min_disk", "image_size", "meta_data", "created_at", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "region", "project_id"},
	),
}
