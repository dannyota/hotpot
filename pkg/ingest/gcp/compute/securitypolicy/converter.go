package securitypolicy

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// SecurityPolicyData holds converted security policy data ready for Ent insertion.
type SecurityPolicyData struct {
	ID                          string
	Name                        string
	Description                 string
	SelfLink                    string
	CreationTimestamp           string
	Type                        string
	Fingerprint                 string
	RulesJSON                   json.RawMessage
	AssociationsJSON            json.RawMessage
	AdaptiveProtectionConfigJSON json.RawMessage
	AdvancedOptionsConfigJSON   json.RawMessage
	DdosProtectionConfigJSON    json.RawMessage
	RecaptchaOptionsConfigJSON  json.RawMessage
	LabelsJSON                  json.RawMessage
	ProjectID                   string
	CollectedAt                 time.Time
}

// ConvertSecurityPolicy converts a GCP API SecurityPolicy to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertSecurityPolicy(p *computepb.SecurityPolicy, projectID string, collectedAt time.Time) (*SecurityPolicyData, error) {
	data := &SecurityPolicyData{
		ID:                fmt.Sprintf("%d", p.GetId()),
		Name:              p.GetName(),
		Description:       p.GetDescription(),
		SelfLink:          p.GetSelfLink(),
		CreationTimestamp: p.GetCreationTimestamp(),
		Type:              p.GetType(),
		Fingerprint:       p.GetFingerprint(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Convert JSONB fields - slices
	if len(p.GetRules()) > 0 {
		rulesBytes, err := json.Marshal(p.GetRules())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rules JSON for %s: %w", p.GetName(), err)
		}
		data.RulesJSON = rulesBytes
	}
	if len(p.GetAssociations()) > 0 {
		assocBytes, err := json.Marshal(p.GetAssociations())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal associations JSON for %s: %w", p.GetName(), err)
		}
		data.AssociationsJSON = assocBytes
	}

	// Convert JSONB fields - config objects (check nil)
	if p.AdaptiveProtectionConfig != nil {
		apBytes, err := json.Marshal(p.AdaptiveProtectionConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal adaptive protection config JSON for %s: %w", p.GetName(), err)
		}
		data.AdaptiveProtectionConfigJSON = apBytes
	}
	if p.AdvancedOptionsConfig != nil {
		aoBytes, err := json.Marshal(p.AdvancedOptionsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal advanced options config JSON for %s: %w", p.GetName(), err)
		}
		data.AdvancedOptionsConfigJSON = aoBytes
	}
	if p.DdosProtectionConfig != nil {
		dpBytes, err := json.Marshal(p.DdosProtectionConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal DDoS protection config JSON for %s: %w", p.GetName(), err)
		}
		data.DdosProtectionConfigJSON = dpBytes
	}
	if p.RecaptchaOptionsConfig != nil {
		roBytes, err := json.Marshal(p.RecaptchaOptionsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal reCAPTCHA options config JSON for %s: %w", p.GetName(), err)
		}
		data.RecaptchaOptionsConfigJSON = roBytes
	}

	// Convert labels map
	if len(p.GetLabels()) > 0 {
		labelsBytes, err := json.Marshal(p.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels JSON for %s: %w", p.GetName(), err)
		}
		data.LabelsJSON = labelsBytes
	}

	return data, nil
}
