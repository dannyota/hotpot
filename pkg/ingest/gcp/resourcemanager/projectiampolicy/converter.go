package projectiampolicy

import (
	"encoding/json"
	"fmt"
	"time"
)

// ProjectIamPolicyData holds converted project IAM policy data ready for Ent insertion.
type ProjectIamPolicyData struct {
	ID           string
	ResourceName string
	Etag         string
	Version      int
	ProjectID    string
	Bindings     []BindingData
	CollectedAt  time.Time
}

// BindingData holds converted binding data.
type BindingData struct {
	Role          string
	MembersJSON   json.RawMessage
	ConditionJSON json.RawMessage
}

// ConvertProjectIamPolicy converts a raw GCP API project IAM policy to Ent-compatible data.
func ConvertProjectIamPolicy(raw *ProjectIamPolicyRaw, collectedAt time.Time) (*ProjectIamPolicyData, error) {
	policy := raw.Policy
	if policy == nil {
		return nil, fmt.Errorf("nil policy for project %s", raw.ProjectID)
	}

	data := &ProjectIamPolicyData{
		ID:           raw.ProjectID,
		ResourceName: "projects/" + raw.ProjectID,
		Etag:         string(policy.GetEtag()),
		Version:      int(policy.GetVersion()),
		ProjectID:    raw.ProjectID,
		CollectedAt:  collectedAt,
	}

	// Convert bindings
	for _, binding := range policy.GetBindings() {
		bd := BindingData{
			Role: binding.GetRole(),
		}

		if members := binding.GetMembers(); len(members) > 0 {
			membersJSON, err := json.Marshal(members)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal members for binding %s: %w", binding.GetRole(), err)
			}
			bd.MembersJSON = membersJSON
		}

		if condition := binding.GetCondition(); condition != nil {
			conditionJSON, err := json.Marshal(condition)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal condition for binding %s: %w", binding.GetRole(), err)
			}
			bd.ConditionJSON = conditionJSON
		}

		data.Bindings = append(data.Bindings, bd)
	}

	return data, nil
}
