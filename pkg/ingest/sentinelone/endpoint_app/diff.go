package endpoint_app

import (
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

func timeEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

// EndpointAppDiff represents changes between old and new endpoint app states.
type EndpointAppDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffEndpointAppData compares old Ent entity and new data.
func DiffEndpointAppData(old *ents1.BronzeS1EndpointApp, new *EndpointAppData) *EndpointAppDiff {
	if old == nil {
		return &EndpointAppDiff{IsNew: true}
	}

	changed := old.AgentID != new.AgentID ||
		old.Name != new.Name ||
		old.Version != new.Version ||
		old.Publisher != new.Publisher ||
		old.Size != new.Size ||
		!timeEqual(old.InstalledDate, new.InstalledDate)

	return &EndpointAppDiff{IsChanged: changed}
}
