package service

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
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
func DiffServiceData(old *ent.BronzeGCPRunService, new *ServiceData) *ServiceDiff {
	diff := &ServiceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.UID != new.UID ||
		old.Generation != new.Generation ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime ||
		old.DeleteTime != new.DeleteTime ||
		old.Creator != new.Creator ||
		old.LastModifier != new.LastModifier ||
		old.Ingress != new.Ingress ||
		old.LaunchStage != new.LaunchStage ||
		old.URI != new.URI ||
		old.ObservedGeneration != new.ObservedGeneration ||
		old.LatestReadyRevision != new.LatestReadyRevision ||
		old.LatestCreatedRevision != new.LatestCreatedRevision ||
		old.Reconciling != new.Reconciling ||
		old.Etag != new.Etag ||
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
