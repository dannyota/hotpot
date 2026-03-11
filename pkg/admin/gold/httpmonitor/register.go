package httpmonitor

import (
	"database/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all Gold HTTP Monitor admin routes.
func Register(db *sql.DB) {
	lh.RegisterSQL(db, sqlTables)
}

var sqlTables = []lh.SQLTable{
	// Anomalies
	{
		API: "/api/v1/gold/httpmonitor/anomalies", Schema: "gold",
		Table: "httpmonitor_anomalies", Nav: admin.NavMeta{Label: "Anomalies", Group: []string{"Gold", "HTTP Monitor"}},
		Columns:             []string{"resource_id", "anomaly_type", "severity", "uri", "method", "baseline_value", "actual_value", "deviation", "description", "window_start", "window_end", "detected_at", "first_detected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "uri", Kind: lh.Search}, {Column: "anomaly_type", Kind: lh.Multi}, {Column: "severity", Kind: lh.Multi}, {Column: "method", Kind: lh.Multi}},
		DefaultSort:         "detected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"anomaly_type", "severity", "method"},
	},
}
