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

// BronzeGCPCloudAssetIAMPolicySearch represents an IAM policy search result in the bronze layer.
// Fields preserve raw API response data from asset.SearchAllIamPolicies.
type BronzeGCPCloudAssetIAMPolicySearch struct {
	ent.Schema
}

func (BronzeGCPCloudAssetIAMPolicySearch) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPCloudAssetIAMPolicySearch) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Full resource name of the resource associated with the IAM policy"),
		field.String("asset_type").
			Optional().
			Comment("Asset type (e.g., compute.googleapis.com/Disk)"),
		field.String("project").
			Optional().
			Comment("Project that the resource belongs to (projects/{PROJECT_NUMBER})"),
		field.String("organization").
			Optional().
			Comment("Organization the IAM policy belongs to (organizations/{ORGANIZATION_NUMBER})"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name (scope of the search)"),

		// JSON fields for nested data
		field.JSON("folders_json", json.RawMessage{}).
			Optional().
			Comment("Folders the IAM policy belongs to as JSON array"),
		field.JSON("policy_json", json.RawMessage{}).
			Optional().
			Comment("IAM policy directly set on the resource as JSON"),
		field.JSON("explanation_json", json.RawMessage{}).
			Optional().
			Comment("Explanation about the IAM policy search result as JSON"),
	}
}

func (BronzeGCPCloudAssetIAMPolicySearch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
		index.Fields("asset_type"),
	}
}

func (BronzeGCPCloudAssetIAMPolicySearch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudasset_iam_policy_searches"},
	}
}
