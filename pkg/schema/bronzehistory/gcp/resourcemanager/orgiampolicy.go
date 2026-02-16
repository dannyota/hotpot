package resourcemanager

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPOrgIamPolicy stores historical snapshots of GCP organization IAM policies.
type BronzeHistoryGCPOrgIamPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPOrgIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPOrgIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze org IAM policy by resource_id"),

		// IAM policy fields
		field.String("resource_name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
	}
}

func (BronzeHistoryGCPOrgIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPOrgIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_org_iam_policies_history"},
	}
}

// BronzeHistoryGCPOrgIamPolicyBinding stores historical snapshots of organization IAM policy bindings.
type BronzeHistoryGCPOrgIamPolicyBinding struct {
	ent.Schema
}

func (BronzeHistoryGCPOrgIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("policy_history_id").
			Comment("Links to parent BronzeHistoryGCPOrgIamPolicy"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Binding fields
		field.String("role").
			NotEmpty(),
		field.JSON("members_json", json.RawMessage{}).
			Optional(),
		field.JSON("condition_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPOrgIamPolicyBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("policy_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPOrgIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_org_iam_policy_bindings_history"},
	}
}
