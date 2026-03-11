package detect

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	detecthttpmon "danny.vn/hotpot/pkg/detect/httpmonitor"
	"danny.vn/hotpot/pkg/detect/lifecycle"
)

// Register wires all detect domains to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB) {
	lifecycle.Register(w, configService, db)
	detecthttpmon.Register(w, configService, driver, db)
}
