package occurrence

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// OccurrenceDiff represents changes between old and new Grafeas occurrence state.
type OccurrenceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *OccurrenceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffOccurrenceData compares existing Ent entity with new OccurrenceData and returns differences.
func DiffOccurrenceData(old *ent.BronzeGCPContainerAnalysisOccurrence, new *OccurrenceData) *OccurrenceDiff {
	diff := &OccurrenceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.ResourceURI != new.ResourceURI ||
		old.NoteName != new.NoteName ||
		old.Kind != new.Kind ||
		old.Remediation != new.Remediation ||
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
		!bytes.Equal(old.SbomReferenceJSON, new.SbomReferenceJSON) ||
		!bytes.Equal(old.EnvelopeJSON, new.EnvelopeJSON) {
		diff.IsChanged = true
	}

	return diff
}
