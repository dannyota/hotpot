package sslpolicy

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SslPolicyDiff represents changes between old and new SSL policy states.
type SslPolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSslPolicyData compares existing Ent entity with new SslPolicyData and returns differences.
func DiffSslPolicyData(old *ent.BronzeGCPComputeSslPolicy, new *SslPolicyData) *SslPolicyDiff {
	diff := &SslPolicyDiff{}

	// New SSL policy
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.Profile != new.Profile ||
		old.MinTLSVersion != new.MinTlsVersion ||
		old.Fingerprint != new.Fingerprint ||
		!bytes.Equal(old.CustomFeaturesJSON, new.CustomFeaturesJSON) ||
		!bytes.Equal(old.EnabledFeaturesJSON, new.EnabledFeaturesJSON) ||
		!bytes.Equal(old.WarningsJSON, new.WarningsJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the SSL policy changed.
func (d *SslPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
