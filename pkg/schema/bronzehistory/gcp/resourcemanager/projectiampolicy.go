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

// BronzeHistoryGCPProjectIamPolicy stores historical snapshots of GCP project IAM policies.
type BronzeHistoryGCPProjectIamPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPProjectIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPProjectIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze project IAM policy by resource_id"),

		// IAM policy fields
		field.String("resource_name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPProjectIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPProjectIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_project_iam_policies_history"},
	}
}

// BronzeHistoryGCPProjectIamPolicyBinding stores historical snapshots of project IAM policy bindings.
type BronzeHistoryGCPProjectIamPolicyBinding struct {
	ent.Schema
}

func (BronzeHistoryGCPProjectIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("policy_history_id").
			Comment("Links to parent BronzeHistoryGCPProjectIamPolicy"),
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

func (BronzeHistoryGCPProjectIamPolicyBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("policy_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPProjectIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_project_iam_policy_bindings_history"},
	}
}
