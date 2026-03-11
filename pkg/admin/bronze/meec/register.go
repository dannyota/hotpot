package meec

import (
	"database/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all MEEC admin routes.
func Register(db *sql.DB) {
	lh.RegisterSQL(db, sqlTables)
}

func bronzeMEEC(api, table, label string) lh.SQLTable {
	return lh.SQLTable{
		API:    "/api/v1/bronze/meec/" + api,
		Schema: "bronze",
		Table:  table,
		Nav:    admin.NavMeta{Label: label, Group: []string{"Bronze", "MEEC"}},
	}
}

var sqlTables = []lh.SQLTable{
	// Computers
	{
		API: "/api/v1/bronze/meec/inventory/computers", Schema: "bronze",
		Table: "meec_inventory_computers", Nav: admin.NavMeta{Label: "Computers", Group: []string{"Bronze", "MEEC"}},
		Columns:             []string{"resource_id", "resource_name", "fqdn_name", "domain_netbios_name", "ip_address", "os_name", "os_platform_name", "agent_version", "computer_live_status", "installation_status", "branch_office_name", "owner", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "resource_name", Kind: lh.Search}, {Column: "os_platform_name", Kind: lh.Multi}, {Column: "computer_live_status", Kind: lh.Multi}, {Column: "installation_status", Kind: lh.Multi}, {Column: "branch_office_name", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"os_platform_name", "computer_live_status", "installation_status", "branch_office_name"},
	},
	// Software
	{
		API: "/api/v1/bronze/meec/inventory/software", Schema: "bronze",
		Table: "meec_inventory_software", Nav: admin.NavMeta{Label: "Software", Group: []string{"Bronze", "MEEC"}},
		Columns:             []string{"resource_id", "software_name", "software_version", "display_name", "manufacturer_name", "sw_category_name", "sw_type", "managed_installations", "network_installations", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "software_name", Kind: lh.Search}, {Column: "manufacturer_name", Kind: lh.Multi}, {Column: "sw_category_name", Kind: lh.Multi}, {Column: "sw_type", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"manufacturer_name", "sw_category_name", "sw_type"},
	},
	// Installed Software
	{
		API: "/api/v1/bronze/meec/inventory/installed-software", Schema: "bronze",
		Table: "meec_inventory_installed_software", Nav: admin.NavMeta{Label: "Installed Software", Group: []string{"Bronze", "MEEC"}},
		Columns:             []string{"resource_id", "computer_resource_id", "software_name", "software_version", "display_name", "manufacturer_name", "architecture", "sw_category_name", "sw_type", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "software_name", Kind: lh.Search}, {Column: "manufacturer_name", Kind: lh.Multi}, {Column: "architecture", Kind: lh.Multi}, {Column: "sw_category_name", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"manufacturer_name", "architecture", "sw_category_name"},
	},
}
