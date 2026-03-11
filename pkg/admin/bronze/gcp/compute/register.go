package compute

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
	"danny.vn/hotpot/pkg/bronzerel"
	gcpcompute "danny.vn/hotpot/pkg/bronzerel/gcp/compute"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	p "danny.vn/hotpot/pkg/storage/ent/gcp/compute/bronzegcpcomputeinstance"
	labelp "danny.vn/hotpot/pkg/storage/ent/gcp/compute/bronzegcpcomputeinstancelabel"
	"danny.vn/hotpot/pkg/storage/ent/gcp/compute/predicate"
)

// Register registers all GCP Compute admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	entClient := entcompute.NewClient(
		entcompute.Driver(driver),
		entcompute.AlternateSchema(entcompute.DefaultSchemaConfig()),
	)

	pathShort := `SUBSTRING("%s" FROM '[^/]*$')`
	instanceFilterOpts := &lh.FilterOptionsConfig{
		DB:      db,
		Schema:  "bronze",
		Table:   "gcp_compute_instances",
		Columns: []string{"instance_type", "status", "zone", "machine_type", "cpu_platform", "project_id"},
		ColumnExprs: map[string]string{
			"zone":         fmt.Sprintf(pathShort, "zone"),
			"machine_type": fmt.Sprintf(pathShort, "machine_type"),
			"instance_type": `CASE WHEN EXISTS (
				SELECT 1 FROM "bronze"."gcp_compute_instance_labels" l
				WHERE l."bronze_gcp_compute_instance_labels" = "gcp_compute_instances"."resource_id"
				AND l."key" = 'goog-gke-node'
			) THEN 'GKE Node' ELSE 'VM' END`,
		},
	}

	admin.RegisterRoute(admin.RouteRegistration{
		Method: "GET",
		Path:   "/api/v1/bronze/gcp/compute/instances",
		Nav:    &admin.NavMeta{Label: "Instances", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		Handler: lh.Handler(lh.Config{
			EntityName: "instances",
			AllowedFields: map[string]bool{
				"name": true, "q": true, "instance_type": true, "status": true, "zone": true, "machine_type": true,
				"project_id": true, "cpu_platform": true, "deletion_protection": true,
				"creation_timestamp": true, "first_collected_at": true, "collected_at": true,
			},
			NewQuery: func() lh.QueryAdapter {
				q := entClient.BronzeGCPComputeInstance.Query()
				return lh.QueryAdapter{
					Where:      func(ps ...lh.Predicate) { q.Where(lh.ConvertSlice[predicate.BronzeGCPComputeInstance](ps)...) },
					CloneCount: func(ctx context.Context) (int, error) { return q.Clone().Count(ctx) },
					Order:      func(os ...lh.Predicate) { q.Order(lh.ConvertSlice[p.OrderOption](os)...) },
					Fetch:      func(ctx context.Context, off, lim int) (any, error) { return q.Offset(off).Limit(lim).All(ctx) },
				}
			},
			Filters: []lh.FilterDef{
				{Field: "q", Kind: lh.Search, Pred: lh.Pred(p.NameContainsFold)},
				{Field: "instance_type", Kind: lh.Multi,
					InFn: func(vs ...string) lh.Predicate { return func(*entsql.Selector) {} }, // both selected = no filter
					EqFn: func(v string) lh.Predicate {
						gke := predicate.BronzeGCPComputeInstance(p.HasLabelsWith(labelp.KeyEQ("goog-gke-node")))
						if v == "GKE Node" {
							return gke
						}
						return p.Not(gke)
					},
				},
				{Field: "status", Kind: lh.Multi, InFn: lh.PredIn(p.StatusIn), EqFn: lh.Pred(p.StatusEQ)},
				{Field: "zone", Kind: lh.Multi,
					InFn: lh.SuffixIn(p.ZoneHasSuffix),
					EqFn: lh.Suffix(p.ZoneHasSuffix),
				},
				{Field: "machine_type", Kind: lh.Multi,
					InFn: lh.SuffixIn(p.MachineTypeHasSuffix),
					EqFn: lh.Suffix(p.MachineTypeHasSuffix),
				},
				{Field: "project_id", Kind: lh.Multi, InFn: lh.PredIn(p.ProjectIDIn), EqFn: lh.Pred(p.ProjectIDEQ)},
			},
			SortFields: map[string]lh.SortFunc{
				"name":               lh.Sort(p.ByName),
				"status":             lh.Sort(p.ByStatus),
				"zone":               lh.Sort(p.ByZone),
				"machine_type":       lh.Sort(p.ByMachineType),
				"cpu_platform":       lh.Sort(p.ByCPUPlatform),
				"project_id":         lh.Sort(p.ByProjectID),
				"creation_timestamp": lh.Sort(p.ByCreationTimestamp),
				"first_collected_at": lh.Sort(p.ByFirstCollectedAt),
				"collected_at":       lh.Sort(p.ByCollectedAt),
			},
			DefaultOrder:  p.ByCreationTimestamp(entsql.OrderDesc()),
			FilterOptions: instanceFilterOpts,
		}),
	})

	admin.RegisterRoute(admin.RouteRegistration{
		Method:  "GET",
		Path:    "/api/v1/bronze/gcp/compute/instances/stats",
		Handler: instanceStatsHandler(db),
	})

	lh.RegisterSQL(db, sqlTables)

	lh.RegisterSQLDetail(db, lh.SQLDetail{
		API:      "/api/v1/bronze/gcp/compute/instances",
		Schema:   "bronze",
		Table:    "gcp_compute_instances",
		IDColumn: "resource_id",
		Edges: []lh.SQLDetailEdge{
			{Key: "nics", Table: "gcp_compute_instance_nics", FKColumn: "bronze_gcp_compute_instance_nics"},
			{Key: "labels", Table: "gcp_compute_instance_labels", FKColumn: "bronze_gcp_compute_instance_labels"},
			{Key: "tags", Table: "gcp_compute_instance_tags", FKColumn: "bronze_gcp_compute_instance_tags"},
			{Key: "metadata", Table: "gcp_compute_instance_metadata", FKColumn: "bronze_gcp_compute_instance_metadata"},
			{Key: "service_accounts", Table: "gcp_compute_instance_service_accounts", FKColumn: "bronze_gcp_compute_instance_service_accounts"},
		},
		Related: []lh.SQLRelated{
			{
				Key: "disks", Schema: "bronze", Table: "gcp_compute_instance_disks",
				Columns:     []string{"device_name", "boot", "auto_delete", "mode", "type", "disk_size_gb", "disk_name", "disk_status", "disk_type", "disk_size", "disk_architecture", "source_image", "index"},
				DefaultSort: "index",
				From: `SELECT
						ad."device_name", ad."boot", ad."auto_delete", ad."mode", ad."type", ad."disk_size_gb", ad."index",
						d."name" AS disk_name, d."status" AS disk_status,
						SUBSTRING(d."type" FROM '[^/]*$') AS disk_type,
						d."size_gb" AS disk_size, d."architecture" AS disk_architecture,
						img."name" AS source_image
					FROM "bronze"."gcp_compute_instance_disks" ad
					LEFT JOIN "bronze"."gcp_compute_disks" d ON ad."source" = d."self_link"
					LEFT JOIN "bronze"."gcp_compute_images" img ON d."source_image" = img."self_link"
					WHERE ad."bronze_gcp_compute_instance_disks" = $1`,
			},
			relatedImages(gcpcompute.InstanceImages()),
			relatedSnapshots(gcpcompute.InstanceSnapshots()),
			relatedFirewalls(gcpcompute.InstanceFirewalls()),
			relatedInstanceGroups(gcpcompute.InstanceGroups()),
			relatedForwardingRules(gcpcompute.InstanceForwardingRules()),
			relatedAddresses(gcpcompute.InstanceAddresses()),
		},
	})
}

// --- Instance detail: related tab helpers (wrap bronzerel with admin-specific columns/sort) ---

func relatedImages(rel bronzerel.Relation) lh.SQLRelated {
	return asRelated(rel, "images",
		[]string{"resource_id", "name", "status", "family", "architecture", "disk_size_gb", "source_type", "source_disk", "creation_timestamp", "project_id"},
		"creation_timestamp", true)
}

func relatedSnapshots(rel bronzerel.Relation) lh.SQLRelated {
	return asRelated(rel, "snapshots",
		[]string{"resource_id", "name", "status", "snapshot_type", "disk_size_gb", "storage_bytes", "architecture", "source_disk", "creation_timestamp", "project_id"},
		"creation_timestamp", true)
}

func relatedFirewalls(rel bronzerel.Relation) lh.SQLRelated {
	return asRelated(rel, "firewalls",
		[]string{"resource_id", "name", "direction", "priority", "disabled", "network", "project_id", "creation_timestamp"},
		"priority", false)
}

func relatedInstanceGroups(rel bronzerel.Relation) lh.SQLRelated {
	return asRelated(rel, "instance-groups",
		[]string{"resource_id", "name", "zone", "size", "network", "creation_timestamp", "project_id"},
		"name", false)
}

func relatedForwardingRules(rel bronzerel.Relation) lh.SQLRelated {
	// Wrap base relation with computed ports column (admin display concern).
	portsExpr := `CASE WHEN t."all_ports" THEN 'ALL'
		WHEN t."ports_json" IS NOT NULL AND t."ports_json"::text NOT IN ('[]','null','') THEN (SELECT string_agg(p, ', ') FROM jsonb_array_elements_text(t."ports_json"::jsonb) AS p)
		WHEN t."port_range" IS NOT NULL AND t."port_range" != '' THEN t."port_range"
		ELSE '' END`
	r := asRelated(rel, "forwarding-rules",
		[]string{"resource_id", "name", "ip_address", "ip_protocol", "ports", "load_balancing_scheme", "backend_service", "project_id", "creation_timestamp"},
		"creation_timestamp", true)
	r.From = fmt.Sprintf(`SELECT t.*, %s AS ports FROM (%s) t`, portsExpr, rel.From)
	return r
}

func relatedAddresses(rel bronzerel.Relation) lh.SQLRelated {
	return asRelated(rel, "addresses",
		[]string{"resource_id", "name", "address", "status", "address_type", "ip_version", "region", "network_tier", "purpose", "project_id", "creation_timestamp"},
		"creation_timestamp", true)
}

// asRelated converts a bronzerel.Relation into an admin SQLRelated with display config.
func asRelated(rel bronzerel.Relation, key string, columns []string, defaultSort string, desc bool) lh.SQLRelated {
	return lh.SQLRelated{
		Key: key, Schema: rel.Schema, Table: rel.Table,
		Columns: columns, DefaultSort: defaultSort, DefaultDesc: desc,
		From: rel.From,
	}
}

func instanceStatsHandler(db *sql.DB) http.HandlerFunc {
	instanceTypeExpr := `CASE WHEN EXISTS (
			SELECT 1 FROM "bronze"."gcp_compute_instance_labels" l
			WHERE l."bronze_gcp_compute_instance_labels" = "gcp_compute_instances"."resource_id"
			AND l."key" = 'goog-gke-node'
		) THEN 'GKE Node' ELSE 'VM' END`

	statsFilters := map[string]admin.StatsFilter{
		"instance_type": {Expr: instanceTypeExpr},
		"status":        {Column: "status"},
		"zone":          {Column: "zone", Suffix: true},
		"machine_type":  {Column: "machine_type", Suffix: true},
		"cpu_platform":  {Column: "cpu_platform"},
		"project_id":    {Column: "project_id"},
	}

	type statusGroup struct {
		Count     int            `json:"count"`
		Breakdown map[string]int `json:"breakdown"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		where, args := admin.StatsWhere(r, statsFilters)
		query := `SELECT COALESCE(status, '') AS status, COALESCE(zone, '') AS zone, COUNT(*) AS cnt
			FROM "bronze"."gcp_compute_instances"` + where + ` GROUP BY 1, 2 ORDER BY 1, 3 DESC`
		rows, err := db.QueryContext(r.Context(), query, args...)
		if err != nil {
			admin.WriteServerError(w, "stats query failed", err)
			return
		}
		defer rows.Close()

		result := map[string]*statusGroup{}
		for rows.Next() {
			var status, zone string
			var cnt int
			if err := rows.Scan(&status, &zone, &cnt); err != nil {
				admin.WriteServerError(w, "stats scan failed", err)
				return
			}
			// Extract short zone name from full path (zones/us-central1-a → us-central1-a)
			if i := len(zone) - 1; i > 0 {
				for ; i >= 0; i-- {
					if zone[i] == '/' {
						zone = zone[i+1:]
						break
					}
				}
			}
			g, ok := result[status]
			if !ok {
				g = &statusGroup{Breakdown: map[string]int{}}
				result[status] = g
			}
			g.Count += cnt
			g.Breakdown[zone] += cnt
		}
		if err := rows.Err(); err != nil {
			admin.WriteServerError(w, "stats query failed", err)
			return
		}

		admin.WriteJSON(w, http.StatusOK, result)
	}
}

func bronzeGCPCompute(api, table, label, subcat string) lh.SQLTable {
	return lh.SQLTable{
		API:    "/api/v1/bronze/gcp/compute/" + api,
		Schema: "bronze",
		Table:  table,
		Nav:    admin.NavMeta{Label: label, Group: []string{"Bronze", "GCP", "Compute", subcat}},
	}
}

var sqlTables = []lh.SQLTable{
	// Instances
	{
		API: "/api/v1/bronze/gcp/compute/instance-groups", Schema: "bronze",
		Table: "gcp_compute_instance_groups", Nav: admin.NavMeta{Label: "Instance Groups", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		Columns:             []string{"resource_id", "name", "zone", "network", "subnetwork", "size", "creation_timestamp", "project_id", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "zone", Kind: lh.Multi, Suffix: true}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "subnetwork", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"zone", "network", "subnetwork", "project_id"},
		ColumnExprs:         map[string]string{"zone": `SUBSTRING("zone" FROM '[^/]*$')`, "network": `SUBSTRING("network" FROM '[^/]*$')`, "subnetwork": `SUBSTRING("subnetwork" FROM '[^/]*$')`},
	},
	{
		API: "/api/v1/bronze/gcp/compute/target-instances", Schema: "bronze",
		Table: "gcp_compute_target_instances", Nav: admin.NavMeta{Label: "Target Instances", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		Columns:             []string{"resource_id", "name", "zone", "instance", "nat_policy", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "zone", Kind: lh.Multi, Suffix: true}, {Column: "nat_policy", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"zone", "nat_policy", "project_id"},
		ColumnExprs:         map[string]string{"zone": `SUBSTRING("zone" FROM '[^/]*$')`},
	},
	{
		API: "/api/v1/bronze/gcp/compute/target-pools", Schema: "bronze",
		Table: "gcp_compute_target_pools", Nav: admin.NavMeta{Label: "Target Pools", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		Columns:             []string{"resource_id", "name", "region", "session_affinity", "failover_ratio", "creation_timestamp", "project_id", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "session_affinity", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"region", "session_affinity", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`},
	},

	// Storage — Disks
	{
		API: "/api/v1/bronze/gcp/compute/disks", Schema: "bronze",
		Table: "gcp_compute_disks", Nav: admin.NavMeta{Label: "Disks", Group: []string{"Bronze", "GCP", "Compute", "Storage"}},
		Columns:             []string{"resource_id", "name", "status", "zone", "type", "size_gb", "architecture", "creation_timestamp", "project_id", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "zone", Kind: lh.Multi, Suffix: true}, {Column: "type", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "zone", "type", "project_id"},
		ColumnExprs:         map[string]string{"zone": `SUBSTRING("zone" FROM '[^/]*$')`, "type": `SUBSTRING("type" FROM '[^/]*$')`},
	},
	// Storage — Snapshots
	{
		API: "/api/v1/bronze/gcp/compute/snapshots", Schema: "bronze",
		Table: "gcp_compute_snapshots", Nav: admin.NavMeta{Label: "Snapshots", Group: []string{"Bronze", "GCP", "Compute", "Storage"}},
		Columns:             []string{"resource_id", "name", "status", "snapshot_type", "disk_size_gb", "storage_bytes", "architecture", "auto_created", "creation_timestamp", "project_id", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "snapshot_type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "snapshot_type", "project_id"},
	},
	// Storage — Images
	{
		API: "/api/v1/bronze/gcp/compute/images", Schema: "bronze",
		Table: "gcp_compute_images", Nav: admin.NavMeta{Label: "Images", Group: []string{"Bronze", "GCP", "Compute", "Storage"}},
		Columns:             []string{"resource_id", "name", "status", "family", "architecture", "disk_size_gb", "source_type", "creation_timestamp", "project_id", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "family", Kind: lh.Multi}, {Column: "architecture", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "family", "architecture", "project_id"},
	},

	// Networking — Networks
	{
		API: "/api/v1/bronze/gcp/compute/networks", Schema: "bronze",
		Table: "gcp_compute_networks", Nav: admin.NavMeta{Label: "Networks", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		Columns:             []string{"resource_id", "name", "routing_mode", "auto_create_subnetworks", "mtu", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "routing_mode", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"routing_mode", "project_id"},
	},
	// Networking — Subnetworks
	{
		API: "/api/v1/bronze/gcp/compute/subnetworks", Schema: "bronze",
		Table: "gcp_compute_subnetworks", Nav: admin.NavMeta{Label: "Subnetworks", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		Columns:             []string{"resource_id", "name", "region", "network", "ip_cidr_range", "purpose", "stack_type", "private_ip_google_access", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "purpose", Kind: lh.Multi}, {Column: "stack_type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"region", "network", "purpose", "stack_type", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`, "network": `SUBSTRING("network" FROM '[^/]*$')`},
	},
	// Networking — Addresses
	{
		API: "/api/v1/bronze/gcp/compute/addresses", Schema: "bronze",
		Table: "gcp_compute_addresses", Nav: admin.NavMeta{Label: "Addresses", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		Columns:             []string{"resource_id", "name", "address", "status", "address_type", "ip_version", "region", "network_tier", "purpose", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "address_type", Kind: lh.Multi}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "network_tier", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "address_type", "region", "network_tier", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`},
	},
	// Networking — Global Addresses
	{
		API: "/api/v1/bronze/gcp/compute/global-addresses", Schema: "bronze",
		Table: "gcp_compute_global_addresses", Nav: admin.NavMeta{Label: "Global Addresses", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		Columns:             []string{"resource_id", "name", "address", "status", "address_type", "ip_version", "network_tier", "purpose", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "address_type", Kind: lh.Multi}, {Column: "purpose", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "address_type", "purpose", "project_id"},
	},
	// Networking — Firewalls
	{
		API: "/api/v1/bronze/gcp/compute/firewalls", Schema: "bronze",
		Table: "gcp_compute_firewalls", Nav: admin.NavMeta{Label: "Firewalls", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		Columns:             []string{"resource_id", "name", "direction", "priority", "disabled", "network", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "direction", Kind: lh.Multi}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"direction", "network", "project_id"},
		ColumnExprs:         map[string]string{"network": `SUBSTRING("network" FROM '[^/]*$')`},
	},
	// Networking — Routers
	{
		API: "/api/v1/bronze/gcp/compute/routers", Schema: "bronze",
		Table: "gcp_compute_routers", Nav: admin.NavMeta{Label: "Routers", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		Columns:             []string{"resource_id", "name", "region", "network", "bgp_asn", "bgp_advertise_mode", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"region", "network", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`, "network": `SUBSTRING("network" FROM '[^/]*$')`},
	},

	// Interconnect & VPN — Interconnects
	{
		API: "/api/v1/bronze/gcp/compute/interconnects", Schema: "bronze",
		Table: "gcp_compute_interconnects", Nav: admin.NavMeta{Label: "Interconnects", Group: []string{"Bronze", "GCP", "Compute", "Interconnect & VPN"}},
		Columns:             []string{"resource_id", "name", "interconnect_type", "link_type", "operational_status", "state", "admin_enabled", "location", "provisioned_link_count", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "interconnect_type", Kind: lh.Multi}, {Column: "link_type", Kind: lh.Multi}, {Column: "operational_status", Kind: lh.Multi}, {Column: "state", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"interconnect_type", "link_type", "operational_status", "state", "project_id"},
	},
	// Interconnect & VPN — VPN Gateways (HA)
	{
		API: "/api/v1/bronze/gcp/compute/vpn-gateways", Schema: "bronze",
		Table: "gcp_compute_vpn_gateways", Nav: admin.NavMeta{Label: "VPN Gateways", Group: []string{"Bronze", "GCP", "Compute", "Interconnect & VPN"}},
		Columns:             []string{"resource_id", "name", "region", "network", "gateway_ip_version", "stack_type", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "stack_type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"region", "network", "stack_type", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`, "network": `SUBSTRING("network" FROM '[^/]*$')`},
	},
	// Interconnect & VPN — Target VPN Gateways (Classic)
	{
		API: "/api/v1/bronze/gcp/compute/target-vpn-gateways", Schema: "bronze",
		Table: "gcp_compute_target_vpn_gateways", Nav: admin.NavMeta{Label: "Target VPN Gateways", Group: []string{"Bronze", "GCP", "Compute", "Interconnect & VPN"}},
		Columns:             []string{"resource_id", "name", "status", "region", "network", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "region", "network", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`, "network": `SUBSTRING("network" FROM '[^/]*$')`},
	},
	// Interconnect & VPN — VPN Tunnels
	{
		API: "/api/v1/bronze/gcp/compute/vpn-tunnels", Schema: "bronze",
		Table: "gcp_compute_vpn_tunnels", Nav: admin.NavMeta{Label: "VPN Tunnels", Group: []string{"Bronze", "GCP", "Compute", "Interconnect & VPN"}},
		Columns:             []string{"resource_id", "name", "status", "region", "peer_ip", "ike_version", "vpn_gateway", "target_vpn_gateway", "router", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "region", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`},
	},
	// Interconnect & VPN — Packet Mirrorings
	{
		API: "/api/v1/bronze/gcp/compute/packet-mirrorings", Schema: "bronze",
		Table: "gcp_compute_packet_mirrorings", Nav: admin.NavMeta{Label: "Packet Mirrorings", Group: []string{"Bronze", "GCP", "Compute", "Interconnect & VPN"}},
		Columns:             []string{"resource_id", "name", "region", "network", "priority", "enable", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "network", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"region", "network", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`, "network": `SUBSTRING("network" FROM '[^/]*$')`},
	},

	// Load Balancing — Backend Services
	{
		API: "/api/v1/bronze/gcp/compute/backend-services", Schema: "bronze",
		Table: "gcp_compute_backend_services", Nav: admin.NavMeta{Label: "Backend Services", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "load_balancing_scheme", "protocol", "port", "region", "enable_cdn", "session_affinity", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "load_balancing_scheme", Kind: lh.Multi}, {Column: "protocol", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"load_balancing_scheme", "protocol", "project_id"},
	},
	// Load Balancing — Forwarding Rules
	{
		API: "/api/v1/bronze/gcp/compute/forwarding-rules", Schema: "bronze",
		Table: "gcp_compute_forwarding_rules", Nav: admin.NavMeta{Label: "Forwarding Rules", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "ip_address", "ip_protocol", "port_range", "load_balancing_scheme", "network_tier", "region", "target", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "ip_protocol", Kind: lh.Multi}, {Column: "load_balancing_scheme", Kind: lh.Multi}, {Column: "network_tier", Kind: lh.Multi}, {Column: "region", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"ip_protocol", "load_balancing_scheme", "network_tier", "region", "project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`},
	},
	// Load Balancing — Global Forwarding Rules
	{
		API: "/api/v1/bronze/gcp/compute/global-forwarding-rules", Schema: "bronze",
		Table: "gcp_compute_global_forwarding_rules", Nav: admin.NavMeta{Label: "Global Forwarding Rules", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "ip_address", "ip_protocol", "port_range", "load_balancing_scheme", "network_tier", "target", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "ip_protocol", Kind: lh.Multi}, {Column: "load_balancing_scheme", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"ip_protocol", "load_balancing_scheme", "project_id"},
	},
	// Load Balancing — Health Checks
	{
		API: "/api/v1/bronze/gcp/compute/health-checks", Schema: "bronze",
		Table: "gcp_compute_health_checks", Nav: admin.NavMeta{Label: "Health Checks", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "type", "check_interval_sec", "timeout_sec", "healthy_threshold", "unhealthy_threshold", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"type", "project_id"},
	},
	// Load Balancing — NEGs
	{
		API: "/api/v1/bronze/gcp/compute/negs", Schema: "bronze",
		Table: "gcp_compute_negs", Nav: admin.NavMeta{Label: "NEGs", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "network_endpoint_type", "zone", "default_port", "size", "region", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "network_endpoint_type", Kind: lh.Multi}, {Column: "zone", Kind: lh.Multi, Suffix: true}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"network_endpoint_type", "zone", "project_id"},
		ColumnExprs:         map[string]string{"zone": `SUBSTRING("zone" FROM '[^/]*$')`},
	},
	// Load Balancing — SSL Policies
	{
		API: "/api/v1/bronze/gcp/compute/ssl-policies", Schema: "bronze",
		Table: "gcp_compute_ssl_policies", Nav: admin.NavMeta{Label: "SSL Policies", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "profile", "min_tls_version", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "profile", Kind: lh.Multi}, {Column: "min_tls_version", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"profile", "min_tls_version", "project_id"},
	},
	// Load Balancing — Target HTTP Proxies
	{
		API: "/api/v1/bronze/gcp/compute/target-http-proxies", Schema: "bronze",
		Table: "gcp_compute_target_http_proxies", Nav: admin.NavMeta{Label: "Target HTTP Proxies", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "url_map", "region", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},
	// Load Balancing — Target HTTPS Proxies
	{
		API: "/api/v1/bronze/gcp/compute/target-https-proxies", Schema: "bronze",
		Table: "gcp_compute_target_https_proxies", Nav: admin.NavMeta{Label: "Target HTTPS Proxies", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "url_map", "ssl_policy", "quic_override", "region", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "quic_override", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"quic_override", "project_id"},
	},
	// Load Balancing — Target SSL Proxies
	{
		API: "/api/v1/bronze/gcp/compute/target-ssl-proxies", Schema: "bronze",
		Table: "gcp_compute_target_ssl_proxies", Nav: admin.NavMeta{Label: "Target SSL Proxies", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "service", "ssl_policy", "proxy_header", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},
	// Load Balancing — Target TCP Proxies
	{
		API: "/api/v1/bronze/gcp/compute/target-tcp-proxies", Schema: "bronze",
		Table: "gcp_compute_target_tcp_proxies", Nav: admin.NavMeta{Label: "Target TCP Proxies", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "service", "proxy_header", "region", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},
	// Load Balancing — URL Maps
	{
		API: "/api/v1/bronze/gcp/compute/url-maps", Schema: "bronze",
		Table: "gcp_compute_url_maps", Nav: admin.NavMeta{Label: "URL Maps", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		Columns:             []string{"resource_id", "name", "default_service", "region", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},

	// Security & Config — Security Policies
	{
		API: "/api/v1/bronze/gcp/compute/security-policies", Schema: "bronze",
		Table: "gcp_compute_security_policies", Nav: admin.NavMeta{Label: "Security Policies", Group: []string{"Bronze", "GCP", "Compute", "Security & Config"}},
		Columns:             []string{"resource_id", "name", "type", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"type", "project_id"},
	},
	// Security & Config — Project Metadata
	{
		API: "/api/v1/bronze/gcp/compute/project-metadata", Schema: "bronze",
		Table: "gcp_compute_project_metadata", Nav: admin.NavMeta{Label: "Project Metadata", Group: []string{"Bronze", "GCP", "Compute", "Security & Config"}},
		Columns:             []string{"resource_id", "name", "default_service_account", "default_network_tier", "xpn_project_status", "project_id", "creation_timestamp", "first_collected_at", "collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "default_network_tier", Kind: lh.Multi}, {Column: "xpn_project_status", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "creation_timestamp", DefaultDesc: true,
		FilterOptionColumns: []string{"default_network_tier", "xpn_project_status", "project_id"},
	},

	// Instances — Child Tables
	{
		API: "/api/v1/bronze/gcp/compute/instance-disks", Schema: "bronze",
		Table: "gcp_compute_instance_disks", Nav: admin.NavMeta{Label: "Instance Disks", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		From: `SELECT d."id", d."source", d."device_name", d."index", d."boot", d."auto_delete", d."mode", d."type", d."disk_size_gb",
			i."name" AS instance_name, i."zone" AS instance_zone, i."project_id"
			FROM "bronze"."gcp_compute_instance_disks" d
			LEFT JOIN "bronze"."gcp_compute_instances" i ON d."bronze_gcp_compute_instance_disks" = i."resource_id"`,
		Columns:             []string{"id", "source", "device_name", "index", "boot", "auto_delete", "mode", "type", "disk_size_gb", "instance_name", "instance_zone", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "device_name", Kind: lh.Search}, {Column: "boot", Kind: lh.Multi}, {Column: "mode", Kind: lh.Multi}, {Column: "type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"boot", "mode", "type", "project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/compute/instance-nics", Schema: "bronze",
		Table: "gcp_compute_instance_nics", Nav: admin.NavMeta{Label: "Instance NICs", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		From: `SELECT n."id", n."name", n."network", n."subnetwork", n."network_ip", n."stack_type", n."nic_type",
			i."name" AS instance_name, i."zone" AS instance_zone, i."project_id"
			FROM "bronze"."gcp_compute_instance_nics" n
			LEFT JOIN "bronze"."gcp_compute_instances" i ON n."bronze_gcp_compute_instance_nics" = i."resource_id"`,
		Columns:             []string{"id", "name", "network", "subnetwork", "network_ip", "stack_type", "nic_type", "instance_name", "instance_zone", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "stack_type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"stack_type", "project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/compute/instance-service-accounts", Schema: "bronze",
		Table: "gcp_compute_instance_service_accounts", Nav: admin.NavMeta{Label: "Instance Service Accounts", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		From: `SELECT sa."id", sa."email",
			i."name" AS instance_name, i."zone" AS instance_zone, i."project_id"
			FROM "bronze"."gcp_compute_instance_service_accounts" sa
			LEFT JOIN "bronze"."gcp_compute_instances" i ON sa."bronze_gcp_compute_instance_service_accounts" = i."resource_id"`,
		Columns:             []string{"id", "email", "instance_name", "instance_zone", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "email", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/compute/instance-group-members", Schema: "bronze",
		Table: "gcp_compute_instance_group_members", Nav: admin.NavMeta{Label: "Instance Group Members", Group: []string{"Bronze", "GCP", "Compute", "Instances"}},
		From: `SELECT m."id", m."instance_name", m."status",
			g."name" AS group_name, g."zone" AS group_zone, g."project_id"
			FROM "bronze"."gcp_compute_instance_group_members" m
			LEFT JOIN "bronze"."gcp_compute_instance_groups" g ON m."bronze_gcp_compute_instance_group_members" = g."resource_id"`,
		Columns:             []string{"id", "instance_name", "status", "group_name", "group_zone", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "instance_name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "project_id"},
	},

	// Networking — Child Tables
	{
		API: "/api/v1/bronze/gcp/compute/firewall-rules", Schema: "bronze",
		Table: "gcp_compute_firewall_alloweds", Nav: admin.NavMeta{Label: "Firewall Allow Rules", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		From: `SELECT a."id", a."ip_protocol", a."ports_json",
			f."name" AS firewall_name, f."direction", f."priority", f."network", f."project_id"
			FROM "bronze"."gcp_compute_firewall_alloweds" a
			LEFT JOIN "bronze"."gcp_compute_firewalls" f ON a."bronze_gcp_compute_firewall_allowed" = f."resource_id"`,
		Columns:             []string{"id", "ip_protocol", "firewall_name", "direction", "priority", "network", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "firewall_name", Kind: lh.Search}, {Column: "ip_protocol", Kind: lh.Multi}, {Column: "direction", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"ip_protocol", "direction", "project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/compute/firewall-deny-rules", Schema: "bronze",
		Table: "gcp_compute_firewall_denieds", Nav: admin.NavMeta{Label: "Firewall Deny Rules", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		From: `SELECT d."id", d."ip_protocol", d."ports_json",
			f."name" AS firewall_name, f."direction", f."priority", f."network", f."project_id"
			FROM "bronze"."gcp_compute_firewall_denieds" d
			LEFT JOIN "bronze"."gcp_compute_firewalls" f ON d."bronze_gcp_compute_firewall_denied" = f."resource_id"`,
		Columns:             []string{"id", "ip_protocol", "firewall_name", "direction", "priority", "network", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "firewall_name", Kind: lh.Search}, {Column: "ip_protocol", Kind: lh.Multi}, {Column: "direction", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"ip_protocol", "direction", "project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/compute/network-peerings", Schema: "bronze",
		Table: "gcp_compute_network_peerings", Nav: admin.NavMeta{Label: "Network Peerings", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		From: `SELECT p."id", p."name", p."network", p."state", p."state_details",
			p."export_custom_routes", p."import_custom_routes", p."exchange_subnet_routes", p."stack_type", p."peer_mtu",
			n."name" AS local_network, n."project_id"
			FROM "bronze"."gcp_compute_network_peerings" p
			LEFT JOIN "bronze"."gcp_compute_networks" n ON p."bronze_gcp_compute_network_peerings" = n."resource_id"`,
		Columns:             []string{"id", "name", "network", "state", "export_custom_routes", "import_custom_routes", "exchange_subnet_routes", "stack_type", "peer_mtu", "local_network", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "state", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"state", "project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/compute/subnetwork-secondary-ranges", Schema: "bronze",
		Table: "gcp_compute_subnetwork_secondary_ranges", Nav: admin.NavMeta{Label: "Secondary Ranges", Group: []string{"Bronze", "GCP", "Compute", "Networking"}},
		From: `SELECT sr."id", sr."range_name", sr."ip_cidr_range",
			s."name" AS subnetwork_name, s."region", s."project_id"
			FROM "bronze"."gcp_compute_subnetwork_secondary_ranges" sr
			LEFT JOIN "bronze"."gcp_compute_subnetworks" s ON sr."bronze_gcp_compute_subnetwork_secondary_ip_ranges" = s."resource_id"`,
		Columns:             []string{"id", "range_name", "ip_cidr_range", "subnetwork_name", "region", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "range_name", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
		ColumnExprs:         map[string]string{"region": `SUBSTRING("region" FROM '[^/]*$')`},
	},

	// Load Balancing — Child Tables
	{
		API: "/api/v1/bronze/gcp/compute/backend-service-backends", Schema: "bronze",
		Table: "gcp_compute_backend_service_backends", Nav: admin.NavMeta{Label: "Backend Service Backends", Group: []string{"Bronze", "GCP", "Compute", "Load Balancing"}},
		From: `SELECT b."id", b."group", b."balancing_mode", b."capacity_scaler", b."failover", b."max_rate", b."max_utilization",
			bs."name" AS service_name, bs."load_balancing_scheme", bs."protocol", bs."project_id"
			FROM "bronze"."gcp_compute_backend_service_backends" b
			LEFT JOIN "bronze"."gcp_compute_backend_services" bs ON b."bronze_gcp_compute_backend_service_backends" = bs."resource_id"`,
		Columns:             []string{"id", "group", "balancing_mode", "capacity_scaler", "failover", "max_rate", "max_utilization", "service_name", "load_balancing_scheme", "protocol", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "service_name", Kind: lh.Search}, {Column: "balancing_mode", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"balancing_mode", "project_id"},
	},
}
