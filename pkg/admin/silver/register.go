package silver

import (
	"database/sql"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/admin/silver/httptraffic"
	"danny.vn/hotpot/pkg/admin/silver/inventory"
)

// Register registers all Silver layer admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	inventory.Register(driver, db)
	httptraffic.Register(db)
}
