package bronze

import (
	"database/sql"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/admin/bronze/apicatalog"
	"danny.vn/hotpot/pkg/admin/bronze/gcp"
	"danny.vn/hotpot/pkg/admin/bronze/greennode"
	"danny.vn/hotpot/pkg/admin/bronze/meec"
	"danny.vn/hotpot/pkg/admin/bronze/s1"
	"danny.vn/hotpot/pkg/admin/bronze/vault"
)

// Register registers all Bronze layer admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	gcp.Register(driver, db)
	greennode.Register(driver, db)
	s1.Register(driver, db)
	meec.Register(db)
	apicatalog.Register(db)
	vault.Register(db)
}
