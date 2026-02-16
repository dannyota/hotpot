package serviceperimeter

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ServicePerimeterDiff represents changes between old and new service perimeter state.
type ServicePerimeterDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ServicePerimeterDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffServicePerimeterData compares existing Ent entity with new ServicePerimeterData and returns differences.
func DiffServicePerimeterData(old *ent.BronzeGCPAccessContextManagerServicePerimeter, new *ServicePerimeterData) *ServicePerimeterDiff {
	diff := &ServicePerimeterDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Title != new.Title ||
		old.Description != new.Description ||
		old.PerimeterType != new.PerimeterType ||
		old.Etag != new.Etag ||
		old.UseExplicitDryRunSpec != new.UseExplicitDryRunSpec ||
		!bytes.Equal(old.StatusJSON, new.StatusJSON) ||
		!bytes.Equal(old.SpecJSON, new.SpecJSON) {
		diff.IsChanged = true
	}

	return diff
}
