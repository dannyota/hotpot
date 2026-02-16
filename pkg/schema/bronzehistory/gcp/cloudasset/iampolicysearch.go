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

// BronzeHistoryGCPCloudAssetIAMPolicySearch stores historical snapshots of IAM policy search results.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPCloudAssetIAMPolicySearch struct {
	ent.Schema
}

func (BronzeHistoryGCPCloudAssetIAMPolicySearch) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPCloudAssetIAMPolicySearch) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze IAM policy search by resource_id"),

		// All IAM policy search fields (same as bronze.BronzeGCPCloudAssetIAMPolicySearch)
		field.String("asset_type").
			Optional(),
		field.String("project").
			Optional(),
		field.String("organization").
			Optional(),
		field.String("organization_id").
			NotEmpty(),

		// JSON fields for nested data
		field.JSON("folders_json", json.RawMessage{}).
			Optional(),
		field.JSON("policy_json", json.RawMessage{}).
			Optional(),
		field.JSON("explanation_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPCloudAssetIAMPolicySearch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPCloudAssetIAMPolicySearch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudasset_iam_policy_searches_history"},
	}
}
