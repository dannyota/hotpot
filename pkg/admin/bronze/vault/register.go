package vault

import (
	"database/sql"

	"danny.vn/hotpot/pkg/admin"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all Vault admin routes.
func Register(db *sql.DB) {
	lh.RegisterSQL(db, sqlTables)
}

var sqlTables = []lh.SQLTable{
	{
		API:    "/api/v1/bronze/vault/pki/certificates",
		Schema: "bronze",
		Table:  "vault_pki_certificates",
		Nav:    admin.NavMeta{Label: "Certificates", Group: []string{"Bronze", "Vault", "PKI"}},
		Columns: []string{
			"resource_id", "vault_name", "mount_path", "common_name",
			"serial_number", "key_type", "key_bits", "not_before", "not_after",
			"is_revoked", "collected_at", "first_collected_at",
		},
		Filters: []lh.SQLFilterDef{
			{Column: "common_name", Kind: lh.Search},
			{Column: "vault_name", Kind: lh.Multi},
			{Column: "mount_path", Kind: lh.Multi},
			{Column: "key_type", Kind: lh.Multi},
			{Column: "is_revoked", Kind: lh.Multi},
		},
		DefaultSort:         "collected_at",
		DefaultDesc:         true,
		FilterOptionColumns: []string{"vault_name", "mount_path", "key_type", "is_revoked"},
	},
}
