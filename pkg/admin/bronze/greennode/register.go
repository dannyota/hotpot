package greennode

import (
	"database/sql"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/admin"
	"danny.vn/hotpot/pkg/admin/bronze/greennode/compute"
	"danny.vn/hotpot/pkg/admin/bronze/greennode/loadbalancer"
	"danny.vn/hotpot/pkg/admin/bronze/greennode/network"
	"danny.vn/hotpot/pkg/admin/bronze/greennode/volume"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all GreenNode admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	compute.Register(driver, db)
	loadbalancer.Register(driver, db)
	network.Register(driver, db)
	volume.Register(driver, db)
	lh.RegisterSQL(db, sqlTables)
}

func bronzeGN(api, table, label, group string, columns []string, filters []lh.SQLFilterDef, filterOptionCols []string) lh.SQLTable {
	return lh.SQLTable{
		API:                 "/api/v1/bronze/greennode/" + api,
		Schema:              "bronze",
		Table:               table,
		Nav:                 admin.NavMeta{Label: label, Group: []string{"Bronze", "GreenNode", group}},
		Columns:             columns,
		Filters:             filters,
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: filterOptionCols,
	}
}

var sqlTables = []lh.SQLTable{
	// DNS
	bronzeGN("dns/hosted-zones", "greennode_dns_hosted_zones", "Hosted Zones", "DNS",
		[]string{"resource_id", "domain_name", "status", "type", "description", "count_records", "portal_user_id", "created_at_api", "updated_at_api", "deleted_at_api", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "domain_name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "type", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "type", "project_id"},
	),
	{
		API:    "/api/v1/bronze/greennode/dns/records",
		Schema: "bronze",
		Table:  "greennode_dns_records",
		Nav:    admin.NavMeta{Label: "Records", Group: []string{"Bronze", "GreenNode", "DNS"}},
		Columns: []string{"id", "record_id", "sub_domain", "status", "type", "routing_policy", "ttl", "created_at_api", "updated_at_api", "deleted_at_api"},
		Filters: []lh.SQLFilterDef{
			{Column: "sub_domain", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "type", Kind: lh.Multi},
			{Column: "routing_policy", Kind: lh.Multi},
		},
		DefaultSort:         "id",
		DefaultDesc:         true,
		FilterOptionColumns: []string{"status", "type", "routing_policy"},
	},

	// GLB
	bronzeGN("glb/load-balancers", "greennode_glb_global_load_balancers", "Load Balancers", "GLB",
		[]string{"resource_id", "name", "status", "package", "type", "description", "user_id", "created_at_api", "updated_at_api", "deleted_at_api", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "status", Kind: lh.Multi},
			{Column: "type", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"status", "type", "project_id"},
	),
	bronzeGN("glb/packages", "greennode_glb_global_packages", "Packages", "GLB",
		[]string{"resource_id", "name", "description", "description_en", "enabled", "base_sku", "base_connection_rate", "base_domestic_traffic_total", "base_non_domestic_traffic_total", "connection_sku", "domestic_traffic_sku", "non_domestic_traffic_sku", "created_at_api", "updated_at_api", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"project_id"},
	),
	bronzeGN("glb/regions", "greennode_glb_global_regions", "Regions", "GLB",
		[]string{"resource_id", "name", "vserver_endpoint", "vlb_endpoint", "ui_server_endpoint", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"project_id"},
	),

	// Portal
	bronzeGN("portal/regions", "greennode_portal_regions", "Regions", "Portal",
		[]string{"resource_id", "name", "description", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"project_id"},
	),
	bronzeGN("portal/zones", "greennode_portal_zones", "Zones", "Portal",
		[]string{"resource_id", "name", "openstack_zone", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"project_id"},
	),
	bronzeGN("portal/quotas", "greennode_portal_quotas", "Quotas", "Portal",
		[]string{"resource_id", "name", "description", "type", "limit_value", "used_value", "region", "project_id", "collected_at", "first_collected_at"},
		[]lh.SQLFilterDef{
			{Column: "name", Kind: lh.Search},
			{Column: "type", Kind: lh.Multi},
			{Column: "region", Kind: lh.Multi},
			{Column: "project_id", Kind: lh.Multi},
		},
		[]string{"type", "region", "project_id"},
	),
}
