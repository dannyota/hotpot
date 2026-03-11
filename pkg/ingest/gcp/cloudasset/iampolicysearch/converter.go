package iampolicysearch

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// IAMPolicySearchData holds converted IAM policy search data ready for Ent insertion.
type IAMPolicySearchData struct {
	ID              string
	AssetType       string
	Project         string
	Organization    string
	OrganizationID  string
	FoldersJSON     json.RawMessage
	PolicyJSON      json.RawMessage
	ExplanationJSON json.RawMessage
	CollectedAt     time.Time
}

// ConvertIAMPolicySearch converts a raw GCP API IAM policy search result to Ent-compatible data.
func ConvertIAMPolicySearch(orgName string, result *assetpb.IamPolicySearchResult, collectedAt time.Time) *IAMPolicySearchData {
	data := &IAMPolicySearchData{
		ID:             result.GetResource(),
		AssetType:      result.GetAssetType(),
		Project:        result.GetProject(),
		Organization:   result.GetOrganization(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}

	// Marshal nested proto fields to JSON
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if folders := result.GetFolders(); len(folders) > 0 {
		if b, err := json.Marshal(folders); err == nil {
			data.FoldersJSON = b
		}
	}
	if policy := result.GetPolicy(); policy != nil {
		if b, err := marshaler.Marshal(policy); err == nil {
			data.PolicyJSON = b
		}
	}
	if explanation := result.GetExplanation(); explanation != nil {
		if b, err := marshaler.Marshal(explanation); err == nil {
			data.ExplanationJSON = b
		}
	}

	return data
}
