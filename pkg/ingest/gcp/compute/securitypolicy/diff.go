package securitypolicy

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SecurityPolicyDiff represents changes between old and new security policy states.
type SecurityPolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSecurityPolicyData compares existing Ent entity with new SecurityPolicyData and returns differences.
func DiffSecurityPolicyData(old *ent.BronzeGCPComputeSecurityPolicy, new *SecurityPolicyData) *SecurityPolicyDiff {
	diff := &SecurityPolicyDiff{}

	// New security policy
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Type != new.Type ||
		old.Fingerprint != new.Fingerprint ||
		!bytes.Equal(old.RulesJSON, new.RulesJSON) ||
		!bytes.Equal(old.AssociationsJSON, new.AssociationsJSON) ||
		!bytes.Equal(old.AdaptiveProtectionConfigJSON, new.AdaptiveProtectionConfigJSON) ||
		!bytes.Equal(old.AdvancedOptionsConfigJSON, new.AdvancedOptionsConfigJSON) ||
		!bytes.Equal(old.DdosProtectionConfigJSON, new.DdosProtectionConfigJSON) ||
		!bytes.Equal(old.RecaptchaOptionsConfigJSON, new.RecaptchaOptionsConfigJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the security policy changed.
func (d *SecurityPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
