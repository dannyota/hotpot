package lifecycle

import (
	"context"
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	entlifecycle "danny.vn/hotpot/pkg/storage/ent/lifecycle"
	"danny.vn/hotpot/pkg/storage/ent/lifecycle/goldlifecyclesoftware"
	"danny.vn/hotpot/pkg/storage/ent/lifecycle/predicate"
)

// Register registers all Gold Lifecycle admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entlifecycle.NewClient(
		entlifecycle.Driver(driver),
		entlifecycle.AlternateSchema(entlifecycle.DefaultSchemaConfig()),
	)

	softwareFilterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "gold",
		Table:   "lifecycle_software",
		Columns: []string{"classification", "eol_status"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/gold/lifecycle/software",
		Nav:    &admin.NavMeta{Label: "Software EOL", Group: []string{"Gold", "Lifecycle"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "software",
			AllowedFields: map[string]bool{
				"name": true, "version": true, "classification": true,
				"eol_status": true, "eol_product_name": true, "eol_cycle": true,
				"eol_date": true, "machine_id": true, "detected_at": true, "first_detected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.GoldLifecycleSoftware.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.GoldLifecycleSoftware](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[goldlifecyclesoftware.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "name", Kind: lh.Search, Pred: lh.Pred(goldlifecyclesoftware.NameContainsFold)},
				{Field: "classification", Kind: lh.Multi, InFn: lh.PredIn(goldlifecyclesoftware.ClassificationIn), EqFn: lh.Pred(goldlifecyclesoftware.ClassificationEQ)},
				{Field: "eol_status", Kind: lh.Multi, InFn: lh.PredIn(goldlifecyclesoftware.EolStatusIn), EqFn: lh.Pred(goldlifecyclesoftware.EolStatusEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":              lh.Sort(goldlifecyclesoftware.ByName),
				"classification":    lh.Sort(goldlifecyclesoftware.ByClassification),
				"eol_status":        lh.Sort(goldlifecyclesoftware.ByEolStatus),
				"eol_date":          lh.Sort(goldlifecyclesoftware.ByEolDate),
				"detected_at":       lh.Sort(goldlifecyclesoftware.ByDetectedAt),
				"first_detected_at": lh.Sort(goldlifecyclesoftware.ByFirstDetectedAt),
			},
			DefaultOrder:  goldlifecyclesoftware.ByDetectedAt(entsql.OrderDesc()),
			FilterOptions: softwareFilterOpts,
		}),
	})

	lh.RegisterSQL(db, sqlTables)
}

func goldLifecycle(api, table, label string) lh.SQLTable {
	return lh.SQLTable{
		API:    "/api/v1/gold/lifecycle/" + api,
		Schema: "gold",
		Table:  table,
		Nav:    admin.NavMeta{Label: label, Group: []string{"Gold", "Lifecycle"}},
	}
}

var sqlTables = []lh.SQLTable{
	// OS EOL
	{
		API: "/api/v1/gold/lifecycle/os", Schema: "gold",
		Table: "lifecycle_os", Nav: admin.NavMeta{Label: "OS EOL", Group: []string{"Gold", "Lifecycle"}},
		Columns:             []string{"resource_id", "machine_id", "hostname", "os_type", "os_name", "eol_status", "eol_product_name", "eol_cycle", "eol_date", "eoas_date", "latest_version", "detected_at", "first_detected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "hostname", Kind: lh.Search}, {Column: "eol_status", Kind: lh.Multi}, {Column: "os_type", Kind: lh.Multi}, {Column: "eol_product_name", Kind: lh.Multi}},
		DefaultSort:         "detected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"eol_status", "os_type", "eol_product_name"},
	},
}
