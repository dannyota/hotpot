package httptraffic

import (
	"database/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all Silver HTTP Traffic admin routes.
func Register(db *sql.DB) {
	lh.RegisterSQL(db, sqlTables)
}

func silverHTTPTraffic(api, table, label string) lh.SQLTable {
	return lh.SQLTable{
		API:    "/api/v1/silver/httptraffic/" + api,
		Schema: "silver",
		Table:  table,
		Nav:    admin.NavMeta{Label: label, Group: []string{"Silver", "HTTP Traffic"}},
	}
}

var sqlTables = []lh.SQLTable{
	// Traffic 5m
	{
		API: "/api/v1/silver/httptraffic/traffic-5m", Schema: "silver",
		Table: "httptraffic_traffic_5m", Nav: admin.NavMeta{Label: "Traffic 5m", Group: []string{"Silver", "HTTP Traffic"}},
		Columns:             []string{"resource_id", "uri", "method", "status_code", "request_count", "avg_request_time", "max_request_time", "total_body_bytes_sent", "unique_client_count", "is_mapped", "access_level", "service", "window_start", "window_end", "collected_at", "first_collected_at", "normalized_at"},
		Filters:             []lh.SQLFilterDef{{Column: "uri", Kind: lh.Search}, {Column: "method", Kind: lh.Multi}, {Column: "is_mapped", Kind: lh.Multi}, {Column: "service", Kind: lh.Multi}, {Column: "access_level", Kind: lh.Multi}},
		DefaultSort:         "window_start", DefaultDesc: true,
		FilterOptionColumns: []string{"method", "is_mapped", "service", "access_level"},
	},
	// Client IP 5m
	{
		API: "/api/v1/silver/httptraffic/client-ip-5m", Schema: "silver",
		Table: "httptraffic_client_ip_5m", Nav: admin.NavMeta{Label: "Client IP 5m", Group: []string{"Silver", "HTTP Traffic"}},
		Columns:             []string{"resource_id", "client_ip", "uri", "method", "country_code", "country_name", "asn", "org_name", "is_internal", "request_count", "is_mapped", "window_start", "window_end", "collected_at", "first_collected_at", "normalized_at"},
		Filters:             []lh.SQLFilterDef{{Column: "client_ip", Kind: lh.Search}, {Column: "country_code", Kind: lh.Multi}, {Column: "org_name", Kind: lh.Multi}, {Column: "is_internal", Kind: lh.Multi}, {Column: "is_mapped", Kind: lh.Multi}},
		DefaultSort:         "window_start", DefaultDesc: true,
		FilterOptionColumns: []string{"country_code", "org_name", "is_internal", "is_mapped"},
	},
	// User Agent 5m
	{
		API: "/api/v1/silver/httptraffic/user-agent-5m", Schema: "silver",
		Table: "httptraffic_user_agent_5m", Nav: admin.NavMeta{Label: "User Agent 5m", Group: []string{"Silver", "HTTP Traffic"}},
		Columns:             []string{"resource_id", "user_agent", "uri", "method", "ua_family", "request_count", "is_mapped", "window_start", "window_end", "collected_at", "first_collected_at", "normalized_at"},
		Filters:             []lh.SQLFilterDef{{Column: "user_agent", Kind: lh.Search}, {Column: "ua_family", Kind: lh.Multi}, {Column: "is_mapped", Kind: lh.Multi}},
		DefaultSort:         "window_start", DefaultDesc: true,
		FilterOptionColumns: []string{"ua_family", "is_mapped"},
	},
}
