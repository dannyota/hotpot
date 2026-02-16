package cloudasset

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPCloudAssetAsset represents a Cloud Asset Inventory asset in the bronze layer.
// Fields preserve raw API response data from asset.ListAssets.
type BronzeGCPCloudAssetAsset struct {
	ent.Schema
}

func (BronzeGCPCloudAssetAsset) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPCloudAssetAsset) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Asset name (full resource URL, e.g., //compute.googleapis.com/projects/123/zones/zone1/instances/instance1)"),
		field.String("asset_type").
			NotEmpty().
			Comment("Asset type (e.g., compute.googleapis.com/Disk)"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),

		// Timestamp as string
		field.String("update_time").
			Optional().
			Comment("Last update timestamp of the asset"),

		// JSON fields for nested data
		field.JSON("resource_json", json.RawMessage{}).
			Optional().
			Comment("Resource representation as JSON"),
		field.JSON("iam_policy_json", json.RawMessage{}).
			Optional().
			Comment("IAM policy set on the resource as JSON"),
		field.JSON("org_policy_json", json.RawMessage{}).
			Optional().
			Comment("Organization policies as JSON array"),
		field.JSON("access_policy_json", json.RawMessage{}).
			Optional().
			Comment("Access context policy (access policy, level, or service perimeter) as JSON"),
		field.JSON("os_inventory_json", json.RawMessage{}).
			Optional().
			Comment("Runtime OS inventory information as JSON"),
	}
}

func (BronzeGCPCloudAssetAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
		index.Fields("asset_type"),
	}
}

func (BronzeGCPCloudAssetAsset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudasset_assets"},
	}
}
