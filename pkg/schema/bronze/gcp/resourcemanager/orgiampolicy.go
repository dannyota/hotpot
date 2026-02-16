package resourcemanager

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPOrgIamPolicy represents a GCP organization IAM policy in the bronze layer.
type BronzeGCPOrgIamPolicy struct {
	ent.Schema
}

func (BronzeGCPOrgIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPOrgIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Organization resource name"),
		field.String("resource_name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
	}
}

func (BronzeGCPOrgIamPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bindings", BronzeGCPOrgIamPolicyBinding.Type),
	}
}

func (BronzeGCPOrgIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
	}
}

func (BronzeGCPOrgIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_org_iam_policies"},
	}
}

// BronzeGCPOrgIamPolicyBinding represents an IAM policy binding for a GCP organization.
type BronzeGCPOrgIamPolicyBinding struct {
	ent.Schema
}

func (BronzeGCPOrgIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("role").
			NotEmpty(),
		field.JSON("members_json", json.RawMessage{}).
			Optional(),
		field.JSON("condition_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPOrgIamPolicyBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", BronzeGCPOrgIamPolicy.Type).
			Ref("bindings").
			Unique().
			Required(),
	}
}

func (BronzeGCPOrgIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_org_iam_policy_bindings"},
	}
}
