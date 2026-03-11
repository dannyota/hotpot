package stats

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"danny.vn/hotpot/pkg/admin"
)

// Register registers stats admin routes.
func Register(db *sql.DB) {
	admin.RegisterRoute(admin.RouteRegistration{
		Method:  "GET",
		Path:    "/api/v1/stats/overview",
		Handler: overviewHandler(db),
	})
}

// stat holds a count and an optional delta compared to the last blueprint.
type stat struct {
	Count int  `json:"count"`
	Delta *int `json:"delta,omitempty"`
}

// bronzeResource is a highlighted resource with its row count.
type bronzeResource struct {
	Label string `json:"label"`
	Count int    `json:"count"`
	Delta *int   `json:"delta,omitempty"`
}

// bronzeProvider holds stats for a single bronze provider.
type bronzeProvider struct {
	Resources []bronzeResource `json:"resources"`
}

// bronzeHighlight pairs a table name with its display label.
type bronzeHighlight struct{ table, label string }

// bronzeProviderDef defines a bronze provider and its highlighted resources.
var bronzeProviderDefs = []struct {
	key        string
	highlights []bronzeHighlight
}{
	{"gcp", []bronzeHighlight{
		{"gcp_compute_instances", "Compute Instances"},
		{"gcp_container_clusters", "GKE Clusters"},
		{"gcp_compute_disks", "Disks"},
		{"gcp_compute_snapshots", "Snapshots"},
		{"gcp_compute_firewalls", "Firewalls"},
		{"gcp_storage_buckets", "Storage Buckets"},
	}},
	{"greennode", []bronzeHighlight{
		{"greennode_compute_servers", "Servers"},
		{"greennode_network_vpcs", "VPCs"},
		{"greennode_network_secgroups", "Security Groups"},
		{"greennode_volume_block_volumes", "Block Volumes"},
	}},
	{"s1", []bronzeHighlight{
		{"s1_agents", "Agents"},
		{"s1_app_inventory", "App Inventory"},
		{"s1_network_discoveries", "Network Discoveries"},
	}},
	{"meec", []bronzeHighlight{
		{"meec_inventory_computers", "Computers"},
		{"meec_inventory_installed_software", "Installed Software"},
	}},
	{"vault", []bronzeHighlight{
		{"vault_pki_certificates", "PKI Certificates"},
	}},
	{"apicatalog", []bronzeHighlight{
		{"apicatalog_endpoints_raw", "API Endpoints"},
	}},
}

func overviewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		admin.WriteJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"bronze": queryBronze(ctx, db),
				"silver": querySilver(ctx, db),
				"gold":   queryGold(ctx, db),
			},
		})
	}
}

// queryBronze counts rows for each highlighted bronze table using COUNT(*).
func queryBronze(ctx context.Context, db *sql.DB) map[string]*bronzeProvider {
	result := map[string]*bronzeProvider{}
	for _, p := range bronzeProviderDefs {
		bp := &bronzeProvider{}
		for _, h := range p.highlights {
			n := countRows(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM bronze.%s`, h.table))
			bp.Resources = append(bp.Resources, bronzeResource{
				Label: h.label,
				Count: n,
				// Delta: populated when blueprint comparison is implemented.
			})
		}
		result[p.key] = bp
	}
	return result
}

func querySilver(ctx context.Context, db *sql.DB) map[string]stat {
	return map[string]stat{
		"machines":       {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.inventory_machines`)},
		"k8s_nodes":      {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.inventory_k8s_nodes`)},
		"software":       {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.inventory_software`)},
		"api_endpoints":  {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.inventory_api_endpoints`)},
		"traffic_5m":     {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.httptraffic_traffic_5m`)},
		"client_ips_5m":  {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.httptraffic_client_ip_5m`)},
		"user_agents_5m": {Count: countRows(ctx, db, `SELECT COUNT(*) FROM silver.httptraffic_user_agent_5m`)},
	}
}

func queryGold(ctx context.Context, db *sql.DB) map[string]stat {
	return map[string]stat{
		"software_eol":  {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.lifecycle_software WHERE eol_status = 'eol_expired'`)},
		"software_eoes": {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.lifecycle_software WHERE eol_status = 'eoes_expired'`)},
		"os_eol":        {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.lifecycle_os WHERE eol_status = 'eol_expired'`)},
		"os_eoes":       {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.lifecycle_os WHERE eol_status = 'eoes_expired'`)},
		"anomalies":          {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.httpmonitor_anomalies`)},
		"anomalies_critical": {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.httpmonitor_anomalies WHERE severity = 'critical'`)},
		"anomalies_high":     {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.httpmonitor_anomalies WHERE severity = 'high'`)},
		"anomalies_medium":   {Count: countRows(ctx, db, `SELECT COUNT(*) FROM gold.httpmonitor_anomalies WHERE severity = 'medium'`)},
	}
}

// countRows runs a COUNT query, returning 0 if the table doesn't exist or query fails.
func countRows(ctx context.Context, db *sql.DB, query string) int {
	var n int
	_ = db.QueryRowContext(ctx, query).Scan(&n)
	return n
}
