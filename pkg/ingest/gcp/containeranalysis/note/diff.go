package note

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// NoteDiff represents changes between old and new Grafeas note state.
type NoteDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *NoteDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffNoteData compares existing Ent entity with new NoteData and returns differences.
func DiffNoteData(old *ent.BronzeGCPContainerAnalysisNote, new *NoteData) *NoteDiff {
	diff := &NoteDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.ShortDescription != new.ShortDescription ||
		old.LongDescription != new.LongDescription ||
		old.Kind != new.Kind ||
		old.ExpirationTime != new.ExpirationTime ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime ||
		!bytes.Equal(old.VulnerabilityJSON, new.VulnerabilityJSON) ||
		!bytes.Equal(old.BuildJSON, new.BuildJSON) ||
		!bytes.Equal(old.ImageJSON, new.ImageJSON) ||
		!bytes.Equal(old.PackageJSON, new.PackageJSON) ||
		!bytes.Equal(old.DeploymentJSON, new.DeploymentJSON) ||
		!bytes.Equal(old.DiscoveryJSON, new.DiscoveryJSON) ||
		!bytes.Equal(old.AttestationJSON, new.AttestationJSON) ||
		!bytes.Equal(old.UpgradeJSON, new.UpgradeJSON) ||
		!bytes.Equal(old.ComplianceJSON, new.ComplianceJSON) ||
		!bytes.Equal(old.DsseAttestationJSON, new.DsseAttestationJSON) ||
		!bytes.Equal(old.VulnerabilityAssessmentJSON, new.VulnerabilityAssessmentJSON) ||
		!bytes.Equal(old.SbomReferenceJSON, new.SbomReferenceJSON) ||
		!bytes.Equal(old.RelatedURLJSON, new.RelatedURLJSON) ||
		!bytes.Equal(old.RelatedNoteNames, new.RelatedNoteNames) {
		diff.IsChanged = true
	}

	return diff
}
