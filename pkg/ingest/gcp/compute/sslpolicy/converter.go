package sslpolicy

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// SslPolicyData holds converted SSL policy data ready for Ent insertion.
type SslPolicyData struct {
	ID                  string
	Name                string
	Description         string
	SelfLink            string
	CreationTimestamp   string
	Profile             string
	MinTlsVersion       string
	Fingerprint         string
	CustomFeaturesJSON  json.RawMessage
	EnabledFeaturesJSON json.RawMessage
	WarningsJSON        json.RawMessage
	ProjectID           string
	CollectedAt         time.Time
}

// ConvertSslPolicy converts a GCP API SslPolicy to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertSslPolicy(p *computepb.SslPolicy, projectID string, collectedAt time.Time) (*SslPolicyData, error) {
	data := &SslPolicyData{
		ID:                fmt.Sprintf("%d", p.GetId()),
		Name:              p.GetName(),
		Description:       p.GetDescription(),
		SelfLink:          p.GetSelfLink(),
		CreationTimestamp: p.GetCreationTimestamp(),
		Profile:           p.GetProfile(),
		MinTlsVersion:     p.GetMinTlsVersion(),
		Fingerprint:       p.GetFingerprint(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Convert JSONB fields
	if features := p.GetCustomFeatures(); len(features) > 0 {
		b, err := json.Marshal(features)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal custom features JSON for %s: %w", p.GetName(), err)
		}
		data.CustomFeaturesJSON = b
	}

	if features := p.GetEnabledFeatures(); len(features) > 0 {
		b, err := json.Marshal(features)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal enabled features JSON for %s: %w", p.GetName(), err)
		}
		data.EnabledFeaturesJSON = b
	}

	if warnings := p.GetWarnings(); len(warnings) > 0 {
		b, err := json.Marshal(warnings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal warnings JSON for %s: %w", p.GetName(), err)
		}
		data.WarningsJSON = b
	}

	return data, nil
}
