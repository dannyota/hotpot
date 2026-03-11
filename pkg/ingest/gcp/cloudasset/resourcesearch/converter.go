package resourcesearch

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// ResourceSearchData holds converted resource search data ready for Ent insertion.
type ResourceSearchData struct {
	ID                       string
	AssetType                string
	Project                  string
	DisplayName              string
	Description              string
	Location                 string
	OrganizationID           string
	LabelsJSON               json.RawMessage
	NetworkTagsJSON          json.RawMessage
	AdditionalAttributesJSON json.RawMessage
	CollectedAt              time.Time
}

// ConvertResourceSearch converts a raw GCP API resource search result to Ent-compatible data.
func ConvertResourceSearch(orgName string, result *assetpb.ResourceSearchResult, collectedAt time.Time) *ResourceSearchData {
	data := &ResourceSearchData{
		ID:             result.GetName(),
		AssetType:      result.GetAssetType(),
		Project:        result.GetProject(),
		DisplayName:    result.GetDisplayName(),
		Description:    result.GetDescription(),
		Location:       result.GetLocation(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}

	// Marshal nested proto fields to JSON
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if labels := result.GetLabels(); len(labels) > 0 {
		if b, err := json.Marshal(labels); err == nil {
			data.LabelsJSON = b
		}
	}
	if tags := result.GetNetworkTags(); len(tags) > 0 {
		if b, err := json.Marshal(tags); err == nil {
			data.NetworkTagsJSON = b
		}
	}
	if attrs := result.GetAdditionalAttributes(); attrs != nil {
		if b, err := marshaler.Marshal(attrs); err == nil {
			data.AdditionalAttributesJSON = b
		}
	}

	return data
}
