package folderiampolicy

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
)

// FolderIamPolicyData holds converted folder IAM policy data ready for Ent insertion.
type FolderIamPolicyData struct {
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

// ConvertFolderIamPolicy converts a raw GCP API folder IAM policy to Ent-compatible data.
func ConvertFolderIamPolicy(folderName string, policy *iampb.Policy, collectedAt time.Time) (*FolderIamPolicyData, error) {
	if policy == nil {
		return nil, fmt.Errorf("nil policy for folder %s", folderName)
	}

	data := &FolderIamPolicyData{
		ID:           folderName,
		ResourceName: folderName,
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
