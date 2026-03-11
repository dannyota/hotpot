package apicatalog

import (
	"database/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all API Catalog admin routes.
func Register(db *sql.DB) {
	lh.RegisterSQL(db, sqlTables)
}

var sqlTables = []lh.SQLTable{
	{
		API:    "/api/v1/bronze/apicatalog/endpoints",
		Schema: "bronze",
		Table:  "apicatalog_endpoints_raw",
		Nav:    admin.NavMeta{Label: "Endpoints", Group: []string{"Bronze", "API Catalog"}},
		Columns: []string{
			"resource_id", "name", "service_name", "upstream", "uri",
			"method", "route_status", "plugin_auth", "source_file",
			"collected_at", "first_collected_at",
		},
		Filters: []lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "uri", Kind: lh.Search},
			{Column: "upstream", Kind: lh.Multi},
			{Column: "route_status", Kind: lh.Multi},
			{Column: "method", Kind: lh.Multi},
		},
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: []string{"upstream", "route_status", "method"},
	},
}
