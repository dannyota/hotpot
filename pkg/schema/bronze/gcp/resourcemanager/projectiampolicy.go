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

// BronzeGCPProjectIamPolicy represents a GCP project IAM policy in the bronze layer.
type BronzeGCPProjectIamPolicy struct {
	ent.Schema
}

func (BronzeGCPProjectIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPProjectIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Project ID"),
		field.String("resource_name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPProjectIamPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bindings", BronzeGCPProjectIamPolicyBinding.Type),
	}
}

func (BronzeGCPProjectIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPProjectIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_project_iam_policies"},
	}
}

// BronzeGCPProjectIamPolicyBinding represents an IAM policy binding for a GCP project.
type BronzeGCPProjectIamPolicyBinding struct {
	ent.Schema
}

func (BronzeGCPProjectIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("role").
			NotEmpty(),
		field.JSON("members_json", json.RawMessage{}).
			Optional(),
		field.JSON("condition_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPProjectIamPolicyBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", BronzeGCPProjectIamPolicy.Type).
			Ref("bindings").
			Unique().
			Required(),
	}
}

func (BronzeGCPProjectIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_project_iam_policy_bindings"},
	}
}
