package urlmap

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// UrlMapData holds converted URL map data ready for Ent insertion.
type UrlMapData struct {
	ID                     string
	Name                   string
	Description            string
	CreationTimestamp      string
	SelfLink               string
	Fingerprint            string
	DefaultService         string
	Region                 string
	HostRulesJSON          []interface{}
	PathMatchersJSON       []interface{}
	TestsJSON              []interface{}
	DefaultRouteActionJSON map[string]interface{}
	DefaultUrlRedirectJSON map[string]interface{}
	HeaderActionJSON       map[string]interface{}
	ProjectID              string
	CollectedAt            time.Time
}

// ConvertUrlMap converts a GCP API UrlMap to Ent-compatible data.
func ConvertUrlMap(um *computepb.UrlMap, projectID string, collectedAt time.Time) (*UrlMapData, error) {
	data := &UrlMapData{
		ID:                fmt.Sprintf("%d", um.GetId()),
		Name:              um.GetName(),
		Description:       um.GetDescription(),
		CreationTimestamp: um.GetCreationTimestamp(),
		SelfLink:          um.GetSelfLink(),
		Fingerprint:       um.GetFingerprint(),
		DefaultService:    um.GetDefaultService(),
		Region:            um.GetRegion(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	if um.HostRules != nil {
		b, err := json.Marshal(um.HostRules)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal host rules for URL map %s: %w", um.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.HostRulesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal host rules: %w", err)
		}
	}

	if um.PathMatchers != nil {
		b, err := json.Marshal(um.PathMatchers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal path matchers for URL map %s: %w", um.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.PathMatchersJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal path matchers: %w", err)
		}
	}

	if um.Tests != nil {
		b, err := json.Marshal(um.Tests)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tests for URL map %s: %w", um.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.TestsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tests: %w", err)
		}
	}

	if um.DefaultRouteAction != nil {
		b, err := json.Marshal(um.DefaultRouteAction)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default route action for URL map %s: %w", um.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.DefaultRouteActionJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default route action: %w", err)
		}
	}

	if um.DefaultUrlRedirect != nil {
		b, err := json.Marshal(um.DefaultUrlRedirect)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default URL redirect for URL map %s: %w", um.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.DefaultUrlRedirectJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default URL redirect: %w", err)
		}
	}

	if um.HeaderAction != nil {
		b, err := json.Marshal(um.HeaderAction)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal header action for URL map %s: %w", um.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.HeaderActionJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal header action: %w", err)
		}
	}

	return data, nil
}
