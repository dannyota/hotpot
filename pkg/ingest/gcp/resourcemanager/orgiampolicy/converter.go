package orgiampolicy

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
)

// OrgIamPolicyData holds converted organization IAM policy data ready for Ent insertion.
type OrgIamPolicyData struct {
	ID           string
	ResourceName string
	Etag         string
	Version      int
	Bindings     []BindingData
	CollectedAt  time.Time
}

// BindingData holds converted binding data.
type BindingData struct {
	Role          string
	MembersJSON   json.RawMessage
	ConditionJSON json.RawMessage
}

// ConvertOrgIamPolicy converts a raw GCP API organization IAM policy to Ent-compatible data.
func ConvertOrgIamPolicy(orgName string, policy *iampb.Policy, collectedAt time.Time) (*OrgIamPolicyData, error) {
	if policy == nil {
		return nil, fmt.Errorf("nil policy for organization %s", orgName)
	}

	data := &OrgIamPolicyData{
		ID:           orgName,
		ResourceName: orgName,
		Etag:         string(policy.GetEtag()),
		Version:      int(policy.GetVersion()),
		CollectedAt:  collectedAt,
	}

	// Convert bindings
	for _, binding := range policy.GetBindings() {
		bd := BindingData{
			Role: binding.GetRole(),
		}

		if len(binding.GetMembers()) > 0 {
			membersJSON, err := json.Marshal(binding.GetMembers())
			if err != nil {
				return nil, fmt.Errorf("failed to marshal members for binding %s: %w", binding.GetRole(), err)
			}
			bd.MembersJSON = membersJSON
		}

		if binding.GetCondition() != nil {
			conditionJSON, err := json.Marshal(binding.GetCondition())
			if err != nil {
				return nil, fmt.Errorf("failed to marshal condition for binding %s: %w", binding.GetRole(), err)
			}
			bd.ConditionJSON = conditionJSON
		}

		data.Bindings = append(data.Bindings, bd)
	}

	return data, nil
}
