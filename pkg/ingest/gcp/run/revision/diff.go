package revision

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// RevisionDiff represents changes between old and new Cloud Run revision state.
type RevisionDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *RevisionDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffRevisionData compares existing Ent entity with new RevisionData and returns differences.
func DiffRevisionData(old *ent.BronzeGCPRunRevision, new *RevisionData) *RevisionDiff {
	diff := &RevisionDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.UID != new.UID ||
		old.Generation != new.Generation ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime ||
		old.DeleteTime != new.DeleteTime ||
		old.LaunchStage != new.LaunchStage ||
		old.ServiceName != new.ServiceName ||
		old.ExecutionEnvironment != new.ExecutionEnvironment ||
		old.EncryptionKey != new.EncryptionKey ||
		old.MaxInstanceRequestConcurrency != new.MaxInstanceRequestConcurrency ||
		old.Timeout != new.Timeout ||
		old.ServiceAccount != new.ServiceAccount ||
		old.Reconciling != new.Reconciling ||
		old.ObservedGeneration != new.ObservedGeneration ||
		old.LogURI != new.LogURI ||
		old.Etag != new.Etag ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.AnnotationsJSON, new.AnnotationsJSON) ||
		!bytes.Equal(old.ScalingJSON, new.ScalingJSON) ||
		!bytes.Equal(old.ContainersJSON, new.ContainersJSON) ||
		!bytes.Equal(old.VolumesJSON, new.VolumesJSON) ||
		!bytes.Equal(old.ConditionsJSON, new.ConditionsJSON) {
		diff.IsChanged = true
	}

	return diff
}
