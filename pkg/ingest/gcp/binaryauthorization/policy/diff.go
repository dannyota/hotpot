package policy

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// PolicyDiff represents changes between old and new Binary Authorization policy state.
type PolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *PolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffPolicyData compares existing Ent entity with new PolicyData and returns differences.
func DiffPolicyData(old *ent.BronzeGCPBinaryAuthorizationPolicy, new *PolicyData) *PolicyDiff {
	diff := &PolicyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Description != new.Description ||
		old.GlobalPolicyEvaluationMode != new.GlobalPolicyEvaluationMode ||
		!bytes.Equal(old.DefaultAdmissionRuleJSON, new.DefaultAdmissionRuleJSON) ||
		!bytes.Equal(old.ClusterAdmissionRulesJSON, new.ClusterAdmissionRulesJSON) ||
		!bytes.Equal(old.KubeNamespaceAdmissionRulesJSON, new.KubeNamespaceAdmissionRulesJSON) ||
		!bytes.Equal(old.IstioServiceIdentityAdmissionRulesJSON, new.IstioServiceIdentityAdmissionRulesJSON) ||
		old.UpdateTime != new.UpdateTime ||
		old.Etag != new.Etag {
		diff.IsChanged = true
	}

	return diff
}
