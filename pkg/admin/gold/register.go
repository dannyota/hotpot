package gold

import (
	"database/sql"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/admin/gold/httpmonitor"
	"danny.vn/hotpot/pkg/admin/gold/lifecycle"
)

// Register registers all Gold layer admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	lifecycle.Register(driver, db)
	httpmonitor.Register(db)
}
