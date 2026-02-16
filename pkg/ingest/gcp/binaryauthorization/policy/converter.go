package policy

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/binaryauthorization/apiv1/binaryauthorizationpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// PolicyData holds converted Binary Authorization policy data ready for Ent insertion.
type PolicyData struct {
	ID                                     string
	Description                            string
	GlobalPolicyEvaluationMode             int
	DefaultAdmissionRuleJSON               json.RawMessage
	ClusterAdmissionRulesJSON              json.RawMessage
	KubeNamespaceAdmissionRulesJSON        json.RawMessage
	IstioServiceIdentityAdmissionRulesJSON json.RawMessage
	UpdateTime                             string
	Etag                                   string
	ProjectID                              string
	CollectedAt                            time.Time
}

// ConvertPolicy converts a raw GCP API Binary Authorization policy to Ent-compatible data.
func ConvertPolicy(p *binaryauthorizationpb.Policy, projectID string, collectedAt time.Time) *PolicyData {
	data := &PolicyData{
		ID:                         p.GetName(),
		Description:                p.GetDescription(),
		GlobalPolicyEvaluationMode: int(p.GetGlobalPolicyEvaluationMode()),
		ProjectID:                  projectID,
		CollectedAt:                collectedAt,
	}

	if ts := p.GetUpdateTime(); ts != nil {
		data.UpdateTime = ts.AsTime().Format(time.RFC3339)
	}

	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if rule := p.GetDefaultAdmissionRule(); rule != nil {
		if b, err := marshaler.Marshal(rule); err == nil {
			data.DefaultAdmissionRuleJSON = b
		}
	}

	if rules := p.GetClusterAdmissionRules(); len(rules) > 0 {
		if b, err := marshalAdmissionRulesMap(marshaler, rules); err == nil {
			data.ClusterAdmissionRulesJSON = b
		}
	}

	if rules := p.GetKubernetesNamespaceAdmissionRules(); len(rules) > 0 {
		if b, err := marshalAdmissionRulesMap(marshaler, rules); err == nil {
			data.KubeNamespaceAdmissionRulesJSON = b
		}
	}

	if rules := p.GetIstioServiceIdentityAdmissionRules(); len(rules) > 0 {
		if b, err := marshalAdmissionRulesMap(marshaler, rules); err == nil {
			data.IstioServiceIdentityAdmissionRulesJSON = b
		}
	}

	return data
}

// marshalAdmissionRulesMap marshals a map of admission rules to JSON.
func marshalAdmissionRulesMap(marshaler protojson.MarshalOptions, rules map[string]*binaryauthorizationpb.AdmissionRule) (json.RawMessage, error) {
	result := make(map[string]json.RawMessage, len(rules))
	for key, rule := range rules {
		b, err := marshaler.Marshal(rule)
		if err != nil {
			return nil, err
		}
		result[key] = b
	}
	return json.Marshal(result)
}
