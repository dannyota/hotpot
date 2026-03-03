package service

import (
	"bytes"

	entrun "github.com/dannyota/hotpot/pkg/storage/ent/gcp/run"
)

// ServiceDiff represents changes between old and new Cloud Run service state.
type ServiceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ServiceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffServiceData compares existing Ent entity with new ServiceData and returns differences.
func DiffServiceData(old *entrun.BronzeGCPRunService, new *ServiceData) *ServiceDiff {
	diff := &ServiceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	// NOTE: UpdateTime, ObservedGeneration, Reconciling, Etag excluded — volatile
	// GCP fields. They are still updated on the bronze record (see service.go no-change path).
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.UID != new.UID ||
		old.Generation != new.Generation ||
		old.CreateTime != new.CreateTime ||
		old.DeleteTime != new.DeleteTime ||
		old.Creator != new.Creator ||
		old.LastModifier != new.LastModifier ||
		old.Ingress != new.Ingress ||
		old.LaunchStage != new.LaunchStage ||
		old.URI != new.URI ||
		old.LatestReadyRevision != new.LatestReadyRevision ||
		old.LatestCreatedRevision != new.LatestCreatedRevision ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.AnnotationsJSON, new.AnnotationsJSON) ||
		!bytes.Equal(old.TemplateJSON, new.TemplateJSON) ||
		!bytes.Equal(old.TrafficJSON, new.TrafficJSON) ||
		!bytes.Equal(old.TerminalConditionJSON, new.TerminalConditionJSON) ||
		!bytes.Equal(old.ConditionsJSON, new.ConditionsJSON) ||
		!bytes.Equal(old.TrafficStatusesJSON, new.TrafficStatusesJSON) {
		diff.IsChanged = true
	}

	return diff
}
