package bucketiam

import (
	"encoding/json"
	"fmt"
	"time"
)

// BucketIamPolicyData holds converted bucket IAM policy data ready for Ent insertion.
type BucketIamPolicyData struct {
	ID          string
	BucketName  string
	Etag        string
	Version     int
	ProjectID   string
	Bindings    []BindingData
	CollectedAt time.Time
}

// BindingData holds converted binding data.
type BindingData struct {
	Role          string
	MembersJSON   json.RawMessage
	ConditionJSON json.RawMessage
}

// ConvertBucketIamPolicy converts a raw GCP API bucket IAM policy to Ent-compatible data.
func ConvertBucketIamPolicy(raw BucketIamPolicyRaw, projectID string, collectedAt time.Time) (*BucketIamPolicyData, error) {
	policy := raw.Policy
	if policy == nil {
		return nil, fmt.Errorf("nil policy for bucket %s", raw.BucketName)
	}

	data := &BucketIamPolicyData{
		ID:          raw.BucketName,
		BucketName:  raw.BucketName,
		Etag:        policy.Etag,
		Version:     int(policy.Version),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	// Convert bindings
	for _, binding := range policy.Bindings {
		bd := BindingData{
			Role: binding.Role,
		}

		if len(binding.Members) > 0 {
			membersJSON, err := json.Marshal(binding.Members)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal members for binding %s: %w", binding.Role, err)
			}
			bd.MembersJSON = membersJSON
		}

		if binding.Condition != nil {
			conditionJSON, err := json.Marshal(binding.Condition)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal condition for binding %s: %w", binding.Role, err)
			}
			bd.ConditionJSON = conditionJSON
		}

		data.Bindings = append(data.Bindings, bd)
	}

	return data, nil
}
