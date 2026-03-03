package normalize

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/normalize/machine"
	"github.com/dannyota/hotpot/pkg/normalize/machine/gcp"
	"github.com/dannyota/hotpot/pkg/normalize/machine/greennode"
	"github.com/dannyota/hotpot/pkg/normalize/machine/meec"
	"github.com/dannyota/hotpot/pkg/normalize/machine/s1"
)

// Register wires all normalize domains to the worker.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, db *sql.DB) {
	// Provider order determines field priority (first non-empty wins).
	providers := []machine.Provider{
		s1.Provider{},
		meec.Provider{},
		greennode.Provider{},
		gcp.Provider{},
	}
	machine.Register(w, configService, driver, db, providers)
}
