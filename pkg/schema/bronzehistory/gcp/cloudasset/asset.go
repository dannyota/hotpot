package cloudasset

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPCloudAssetAsset stores historical snapshots of Cloud Asset Inventory assets.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPCloudAssetAsset struct {
	ent.Schema
}

func (BronzeHistoryGCPCloudAssetAsset) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPCloudAssetAsset) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Cloud Asset asset by resource_id"),

		// All asset fields (same as bronze.BronzeGCPCloudAssetAsset)
		field.String("asset_type").
			NotEmpty(),
		field.String("organization_id").
			NotEmpty(),

		// Timestamp as string
		field.String("update_time").
			Optional(),

		// JSON fields for nested data
		field.JSON("resource_json", json.RawMessage{}).
			Optional(),
		field.JSON("iam_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("org_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("access_policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("os_inventory_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPCloudAssetAsset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPCloudAssetAsset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudasset_assets_history"},
	}
}
