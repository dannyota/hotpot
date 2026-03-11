package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	entmachine "danny.vn/hotpot/pkg/storage/ent/inventory/machine"
	"danny.vn/hotpot/pkg/storage/ent/inventory/machine/inventorymachine"
	"danny.vn/hotpot/pkg/storage/ent/inventory/machine/predicate"
)

// Register registers all Silver Inventory admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entmachine.NewClient(
		entmachine.Driver(driver),
		entmachine.AlternateSchema(entmachine.DefaultSchemaConfig()),
	)

	machineFilterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "silver",
		Table:   "inventory_machines",
		Columns: []string{"os_type", "status", "environment", "cloud_project"},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/silver/inventory/machines",
		Nav:    &admin.NavMeta{Label: "Machines", Group: []string{"Silver", "Inventory"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "machines",
			AllowedFields: map[string]bool{
				"resource_id": true,
				"hostname": true, "os_type": true, "os_name": true,
				"status": true, "internal_ip": true, "external_ip": true,
				"environment": true, "cloud_project": true, "cloud_zone": true,
				"created": true, "collected_at": true, "first_collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.InventoryMachine.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.InventoryMachine](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[inventorymachine.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "resource_id", Kind: lh.Exact, Pred: lh.Pred(inventorymachine.IDEQ)},
				{Field: "hostname", Kind: lh.Search, Pred: lh.Pred(inventorymachine.HostnameContainsFold)},
				{Field: "os_type", Kind: lh.Multi, InFn: lh.PredIn(inventorymachine.OsTypeIn), EqFn: lh.Pred(inventorymachine.OsTypeEQ)},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(inventorymachine.StatusIn), EqFn: lh.Pred(inventorymachine.StatusEQ)},
				{Field: "environment", Kind: lh.Multi, InFn: lh.PredIn(inventorymachine.EnvironmentIn), EqFn: lh.Pred(inventorymachine.EnvironmentEQ)},
				{Field: "cloud_project", Kind: lh.Multi, InFn: lh.PredIn(inventorymachine.CloudProjectIn), EqFn: lh.Pred(inventorymachine.CloudProjectEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"hostname":         lh.Sort(inventorymachine.ByHostname),
				"os_type":          lh.Sort(inventorymachine.ByOsType),
				"status":           lh.Sort(inventorymachine.ByStatus),
				"environment":      lh.Sort(inventorymachine.ByEnvironment),
				"cloud_project":    lh.Sort(inventorymachine.ByCloudProject),
				"created":           lh.Sort(inventorymachine.ByCreated),
				"collected_at":     lh.Sort(inventorymachine.ByCollectedAt),
				"first_collected_at": lh.Sort(inventorymachine.ByFirstCollectedAt),
			},
			DefaultOrder:  inventorymachine.ByCreated(entsql.OrderDesc()),
			FilterOptions: machineFilterOpts,
		}),
	})

	// Machine name lookup: returns {machine_id: hostname} for all machines.
	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/silver/inventory/machines/names",
		Handler: machineNamesHandler(db),
	})

	lh.RegisterSQL(db, sqlTables)
}

func machineNamesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.QueryContext(r.Context(),
			`SELECT "resource_id", "hostname" FROM "silver"."inventory_machines"`)
		if err != nil {
			admin.WriteServerError(w, "failed to load machine names", err)
			return
		}
		defer rows.Close()

		names := make(map[string]string)
		for rows.Next() {
			var id, hostname string
			if err := rows.Scan(&id, &hostname); err != nil {
				admin.WriteServerError(w, "failed to scan machine names", err)
				return
			}
			names[id] = hostname
		}
		if err := rows.Err(); err != nil {
			admin.WriteServerError(w, "failed to load machine names", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(names)
	}
}

var sqlTables = []lh.SQLTable{
	// K8s Nodes
	{
		API: "/api/v1/silver/inventory/k8s-nodes", Schema: "silver",
		Table: "inventory_k8s_nodes", Nav: admin.NavMeta{Label: "K8s Nodes", Group: []string{"Silver", "Inventory"}},
		Columns:             []string{"resource_id", "node_name", "cluster_name", "node_pool", "status", "provisioning", "cloud_project", "cloud_zone", "internal_ip", "external_ip", "collected_at", "first_collected_at", "normalized_at"},
		Filters:             []lh.SQLFilterDef{{Column: "node_name", Kind: lh.Search}, {Column: "cluster_name", Kind: lh.Multi}, {Column: "status", Kind: lh.Multi}, {Column: "cloud_project", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"cluster_name", "status", "cloud_project"},
	},
	// Software — joined with machines to expose hostname, status, environment.
	{
		API: "/api/v1/silver/inventory/software", Schema: "silver",
		Table: "inventory_software", Nav: admin.NavMeta{Label: "Software", Group: []string{"Silver", "Inventory"}},
		From: `SELECT s."resource_id", s."machine_id", s."name", s."version", s."publisher", s."installed_on",
			s."collected_at", s."first_collected_at", s."normalized_at",
			m."hostname" AS machine_hostname, m."status" AS machine_status, m."environment" AS machine_environment
			FROM "silver"."inventory_software" s
			LEFT JOIN "silver"."inventory_machines" m ON s."machine_id" = m."resource_id"`,
		Columns: []string{"resource_id", "machine_id", "name", "version", "publisher", "installed_on",
			"collected_at", "first_collected_at", "normalized_at",
			"machine_hostname", "machine_status", "machine_environment"},
		Filters: []lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "publisher", Kind: lh.Multi},
			{Column: "machine_id", Kind: lh.Multi},
			{Column: "machine_status", Kind: lh.Multi},
			{Column: "machine_environment", Kind: lh.Multi},
		},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"publisher", "machine_status", "machine_environment"},
	},
	// API Endpoints
	{
		API: "/api/v1/silver/inventory/api-endpoints", Schema: "silver",
		Table: "inventory_api_endpoints", Nav: admin.NavMeta{Label: "API Endpoints", Group: []string{"Silver", "Inventory"}},
		Columns:             []string{"resource_id", "name", "service", "uri_pattern", "is_active", "access_level", "collected_at", "first_collected_at", "normalized_at"},
		Filters:             []lh.SQLFilterDef{{Column: "uri_pattern", Kind: lh.Search}, {Column: "service", Kind: lh.Multi}, {Column: "access_level", Kind: lh.Multi}, {Column: "is_active", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"service", "access_level", "is_active"},
	},
}
