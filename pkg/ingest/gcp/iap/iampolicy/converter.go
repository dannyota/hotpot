package iampolicy

import (
	"encoding/json"
	"fmt"
	"time"

	iamv1 "google.golang.org/genproto/googleapis/iam/v1"
)

// IAMPolicyData holds converted IAP IAM policy data ready for Ent insertion.
type IAMPolicyData struct {
	ID               string
	Name             string
	Etag             string
	Version          int
	BindingsJSON     json.RawMessage
	AuditConfigsJSON json.RawMessage
	ProjectID        string
	CollectedAt      time.Time
}

// ConvertIAMPolicy converts a raw GCP API IAM policy to Ent-compatible data.
func ConvertIAMPolicy(policy *iamv1.Policy, projectID string, collectedAt time.Time) (*IAMPolicyData, error) {
	resourceName := "projects/" + projectID + "/iap_web"

	data := &IAMPolicyData{
		ID:          resourceName,
		Name:        resourceName,
		Etag:        string(policy.GetEtag()),
		Version:     int(policy.GetVersion()),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if len(policy.GetBindings()) > 0 {
		bindingsJSON, err := json.Marshal(policy.GetBindings())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal bindings for %s: %w", resourceName, err)
		}
		data.BindingsJSON = bindingsJSON
	}

	if len(policy.GetAuditConfigs()) > 0 {
		auditJSON, err := json.Marshal(policy.GetAuditConfigs())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal audit_configs for %s: %w", resourceName, err)
		}
		data.AuditConfigsJSON = auditJSON
	}

	return data, nil
}
