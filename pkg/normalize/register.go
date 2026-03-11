package normalize

import (
	"database/sql"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	normhttptraffic "danny.vn/hotpot/pkg/normalize/httptraffic"
	"danny.vn/hotpot/pkg/normalize/inventory/apiendpoint"
	"danny.vn/hotpot/pkg/normalize/inventory/apiendpoint/manual"
	"danny.vn/hotpot/pkg/normalize/inventory/k8snode"
	k8snodegcp "danny.vn/hotpot/pkg/normalize/inventory/k8snode/gcp"
	"danny.vn/hotpot/pkg/normalize/inventory/machine"
	"danny.vn/hotpot/pkg/normalize/inventory/machine/gcp"
	"danny.vn/hotpot/pkg/normalize/inventory/machine/greennode"
	"danny.vn/hotpot/pkg/normalize/inventory/machine/meec"
	"danny.vn/hotpot/pkg/normalize/inventory/machine/s1"
	"danny.vn/hotpot/pkg/normalize/inventory/software"
	swmeec "danny.vn/hotpot/pkg/normalize/inventory/software/meec"
	sws1 "danny.vn/hotpot/pkg/normalize/inventory/software/s1"
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

	// Software providers.
	swProviders := []software.Provider{
		sws1.Provider{},
		swmeec.Provider{},
	}
	software.Register(w, configService, driver, db, swProviders)

	// API endpoint providers.
	apiProviders := []apiendpoint.Provider{
		manual.Provider{},
	}
	apiendpoint.Register(w, configService, driver, db, apiProviders)

	// HTTP traffic normalization.
	normhttptraffic.Register(w, configService, driver, db)
}
