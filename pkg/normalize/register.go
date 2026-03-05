package normalize

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/normalize/installedsoftware"
	ismeec "github.com/dannyota/hotpot/pkg/normalize/installedsoftware/meec"
	iss1 "github.com/dannyota/hotpot/pkg/normalize/installedsoftware/s1"
	"github.com/dannyota/hotpot/pkg/normalize/k8snode"
	k8snodegcp "github.com/dannyota/hotpot/pkg/normalize/k8snode/gcp"
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

	// K8s node providers.
	k8sProviders := []k8snode.Provider{
		k8snodegcp.Provider{},
	}
	k8snode.Register(w, configService, driver, db, k8sProviders)

	// Installed software providers.
	isProviders := []installedsoftware.Provider{
		iss1.Provider{},
		ismeec.Provider{},
	}
	installedsoftware.Register(w, configService, driver, db, isProviders)
}
