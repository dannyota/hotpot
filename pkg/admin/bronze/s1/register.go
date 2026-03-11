package s1

import (
	"context"
	"database/sql"
	"net/http"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzes1agent"
	"danny.vn/hotpot/pkg/storage/ent/s1/predicate"
)

// Register registers all SentinelOne admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := ents1.NewClient(
		ents1.Driver(driver),
		ents1.AlternateSchema(ents1.DefaultSchemaConfig()),
	)

	agentFilterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "s1_agents",
		Columns: []string{"os_type", "site_name", "network_status", "is_active", "is_infected"},
		ColumnExprs: map[string]string{
			"is_active":   `CAST("is_active" AS TEXT)`,
			"is_infected": `CAST("is_infected" AS TEXT)`,
		},
	}

	p := bronzes1agent.ByLastActiveDate

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/s1/agents",
		Nav:    &admin.NavMeta{Label: "Agents", Group: []string{"Bronze", "SentinelOne"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "agents",
			AllowedFields: map[string]bool{
				"computer_name": true, "os_name": true, "agent_version": true,
				"is_active": true, "is_infected": true, "network_status": true,
				"site_name": true, "last_active_date": true, "os_type": true,
				"collected_at": true, "first_collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeS1Agent.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeS1Agent](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[bronzes1agent.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "computer_name", Kind: lh.Search, Pred: lh.Pred(bronzes1agent.ComputerNameContainsFold)},
				{Field: "os_type", Kind: lh.Multi, InFn: lh.PredIn(bronzes1agent.OsTypeIn), EqFn: lh.Pred(bronzes1agent.OsTypeEQ)},
				{Field: "site_name", Kind: lh.Multi, InFn: lh.PredIn(bronzes1agent.SiteNameIn), EqFn: lh.Pred(bronzes1agent.SiteNameEQ)},
				{Field: "network_status", Kind: lh.Multi, InFn: lh.PredIn(bronzes1agent.NetworkStatusIn), EqFn: lh.Pred(bronzes1agent.NetworkStatusEQ)},
				{Field: "is_active", Kind: lh.Exact, Pred: lh.BoolPred(bronzes1agent.IsActiveEQ)},
				{Field: "is_infected", Kind: lh.Exact, Pred: lh.BoolPred(bronzes1agent.IsInfectedEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"computer_name":    lh.Sort(bronzes1agent.ByComputerName),
				"os_type":          lh.Sort(bronzes1agent.ByOsType),
				"site_name":        lh.Sort(bronzes1agent.BySiteName),
				"network_status":   lh.Sort(bronzes1agent.ByNetworkStatus),
				"last_active_date": lh.Sort(bronzes1agent.ByLastActiveDate),
				"collected_at":     lh.Sort(bronzes1agent.ByCollectedAt),
				"first_collected_at": lh.Sort(bronzes1agent.ByFirstCollectedAt),
			},
			DefaultOrder:  p(entsql.OrderDesc()),
			FilterOptions: agentFilterOpts,
		}),
	})

	admin.RegisterRoute(admin.RouteRegistration{
		Method:  "GET",
		Path:    "/api/v1/bronze/s1/agents/stats",
		Handler: agentStatsHandler(db),
	})

	lh.RegisterSQL(db, sqlTables)
}

func agentStatsHandler(db *sql.DB) http.HandlerFunc {
	statsFilters := map[string]admin.StatsFilter{
		"os_type":        {Column: "os_type"},
		"site_name":      {Column: "site_name"},
		"network_status": {Column: "network_status"},
		"is_active":      {Column: "is_active"},
		"is_infected":    {Column: "is_infected"},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		where, args := admin.StatsWhere(r, statsFilters)
		q := `SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE is_active = true) AS active,
			COUNT(*) FILTER (WHERE is_active = false) AS inactive,
			COUNT(*) FILTER (WHERE is_infected = true) AS infected,
			COUNT(*) FILTER (WHERE network_status = 'connected') AS connected,
			COUNT(*) FILTER (WHERE network_status = 'disconnected') AS disconnected
		FROM "bronze"."s1_agents"` + where
		var total, active, inactive, infected, connected, disconnected int
		if err := db.QueryRowContext(r.Context(), q, args...).Scan(&total, &active, &inactive, &infected, &connected, &disconnected); err != nil {
			admin.WriteServerError(w, "stats query failed", err)
			return
		}
		admin.WriteJSON(w, http.StatusOK, map[string]int{
			"total": total, "active": active, "inactive": inactive,
			"infected": infected, "connected": connected, "disconnected": disconnected,
		})
	}
}

func bronzeS1(api, table, label string) lh.SQLTable {
	return lh.SQLTable{
		API:    "/api/v1/bronze/s1/" + api,
		Schema: "bronze",
		Table:  table,
		Nav:    admin.NavMeta{Label: label, Group: []string{"Bronze", "SentinelOne"}},
	}
}

var sqlTables = []lh.SQLTable{
	// Accounts
	{
		API: "/api/v1/bronze/s1/accounts", Schema: "bronze",
		Table: "s1_accounts", Nav: admin.NavMeta{Label: "Accounts", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "name", "state", "account_type", "active_agents", "total_licenses", "usage_type", "billing_mode", "unlimited_expiration", "api_created_at", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "state", Kind: lh.Multi}, {Column: "account_type", Kind: lh.Multi}, {Column: "usage_type", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"state", "account_type", "usage_type"},
	},
	// Sites
	{
		API: "/api/v1/bronze/s1/sites", Schema: "bronze",
		Table: "s1_sites", Nav: admin.NavMeta{Label: "Sites", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "name", "state", "site_type", "account_name", "active_licenses", "total_licenses", "health_status", "is_default", "suite", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "state", Kind: lh.Multi}, {Column: "site_type", Kind: lh.Multi}, {Column: "account_name", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"state", "site_type", "account_name"},
	},
	// Groups
	{
		API: "/api/v1/bronze/s1/groups", Schema: "bronze",
		Table: "s1_groups", Nav: admin.NavMeta{Label: "Groups", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "name", "type", "site_id", "is_default", "inherits", "total_agents", "rank", "creator", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "type", Kind: lh.Multi}, {Column: "is_default", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"type", "is_default"},
	},
	// App Inventory
	{
		API: "/api/v1/bronze/s1/app-inventory", Schema: "bronze",
		Table: "s1_app_inventory", Nav: admin.NavMeta{Label: "App Inventory", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "application_name", "application_vendor", "endpoints_count", "application_versions_count", "estimate", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "application_name", Kind: lh.Search}, {Column: "application_vendor", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"application_vendor"},
	},
	// Endpoint Apps
	{
		API: "/api/v1/bronze/s1/endpoint-apps", Schema: "bronze",
		Table: "s1_endpoint_apps", Nav: admin.NavMeta{Label: "Endpoint Apps", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "agent_id", "name", "version", "publisher", "size", "installed_date", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "publisher", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"publisher"},
	},
	// Network Discoveries
	{
		API: "/api/v1/bronze/s1/network-discoveries", Schema: "bronze",
		Table: "s1_network_discoveries", Nav: admin.NavMeta{Label: "Network Discoveries", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "name", "ip_address", "category", "sub_category", "os", "os_family", "manufacturer", "asset_status", "infection_status", "device_review", "detected_from_site", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "category", Kind: lh.Multi}, {Column: "os_family", Kind: lh.Multi}, {Column: "asset_status", Kind: lh.Multi}, {Column: "infection_status", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"category", "os_family", "asset_status", "infection_status"},
	},
	// Ranger Devices
	{
		API: "/api/v1/bronze/s1/ranger-devices", Schema: "bronze",
		Table: "s1_ranger_devices", Nav: admin.NavMeta{Label: "Ranger Devices", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "local_ip", "external_ip", "mac_address", "os_type", "os_name", "device_type", "device_function", "manufacturer", "managed_state", "site_name", "first_seen", "last_seen", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "local_ip", Kind: lh.Search}, {Column: "os_type", Kind: lh.Multi}, {Column: "device_type", Kind: lh.Multi}, {Column: "managed_state", Kind: lh.Multi}, {Column: "site_name", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"os_type", "device_type", "managed_state", "site_name"},
	},
	// Ranger Gateways
	{
		API: "/api/v1/bronze/s1/ranger-gateways", Schema: "bronze",
		Table: "s1_ranger_gateways", Nav: admin.NavMeta{Label: "Ranger Gateways", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "ip", "mac_address", "external_ip", "manufacturer", "network_name", "account_name", "number_of_agents", "number_of_rangers", "connected_rangers", "allow_scan", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "ip", Kind: lh.Search}, {Column: "network_name", Kind: lh.Multi}, {Column: "account_name", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"network_name", "account_name"},
	},
	// Ranger Settings
	{
		API: "/api/v1/bronze/s1/ranger-settings", Schema: "bronze",
		Table: "s1_ranger_settings", Nav: admin.NavMeta{Label: "Ranger Settings", Group: []string{"Bronze", "SentinelOne"}},
		Columns:             []string{"resource_id", "account_id", "scope_id", "enabled", "tcp_port_scan", "udp_port_scan", "icmp_scan", "smb_scan", "mdns_scan", "rdns_scan", "snmp_scan", "use_periodic_snapshots", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "account_id", Kind: lh.Search}, {Column: "enabled", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"enabled"},
	},
}
