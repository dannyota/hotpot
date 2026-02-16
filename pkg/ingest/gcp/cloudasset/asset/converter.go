package asset

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// AssetData holds converted Cloud Asset data ready for Ent insertion.
type AssetData struct {
	ID               string
	AssetType        string
	OrganizationID   string
	UpdateTime       string
	ResourceJSON     json.RawMessage
	IamPolicyJSON    json.RawMessage
	OrgPolicyJSON    json.RawMessage
	AccessPolicyJSON json.RawMessage
	OsInventoryJSON  json.RawMessage
	CollectedAt      time.Time
}

// ConvertAsset converts a raw GCP API Cloud Asset to Ent-compatible data.
func ConvertAsset(orgName string, a *assetpb.Asset, collectedAt time.Time) *AssetData {
	data := &AssetData{
		ID:             a.GetName(),
		AssetType:      a.GetAssetType(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}

	// Convert update time
	if a.GetUpdateTime() != nil {
		data.UpdateTime = a.GetUpdateTime().AsTime().Format(time.RFC3339)
	}

	// Marshal nested proto fields to JSON
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if res := a.GetResource(); res != nil {
		if b, err := marshaler.Marshal(res); err == nil {
			data.ResourceJSON = b
		}
	}
	if policy := a.GetIamPolicy(); policy != nil {
		if b, err := marshaler.Marshal(policy); err == nil {
			data.IamPolicyJSON = b
		}
	}
	if orgPolicies := a.GetOrgPolicy(); len(orgPolicies) > 0 {
		if b, err := json.Marshal(orgPolicies); err == nil {
			data.OrgPolicyJSON = b
		}
	}
	if ap := a.GetAccessPolicy(); ap != nil {
		if b, err := marshaler.Marshal(ap); err == nil {
			data.AccessPolicyJSON = b
		}
	}
	if inv := a.GetOsInventory(); inv != nil {
		if b, err := marshaler.Marshal(inv); err == nil {
			data.OsInventoryJSON = b
		}
	}

	return data
}
