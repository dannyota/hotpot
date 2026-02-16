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

// BronzeGCPFolderIamPolicy represents a GCP folder IAM policy in the bronze layer.
type BronzeGCPFolderIamPolicy struct {
	ent.Schema
}

func (BronzeGCPFolderIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPFolderIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Folder resource name"),
		field.String("resource_name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
	}
}

func (BronzeGCPFolderIamPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bindings", BronzeGCPFolderIamPolicyBinding.Type),
	}
}

func (BronzeGCPFolderIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
	}
}

func (BronzeGCPFolderIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_folder_iam_policies"},
	}
}

// BronzeGCPFolderIamPolicyBinding represents an IAM policy binding for a GCP folder.
type BronzeGCPFolderIamPolicyBinding struct {
	ent.Schema
}

func (BronzeGCPFolderIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("role").
			NotEmpty(),
		field.JSON("members_json", json.RawMessage{}).
			Optional(),
		field.JSON("condition_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPFolderIamPolicyBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", BronzeGCPFolderIamPolicy.Type).
			Ref("bindings").
			Unique().
			Required(),
	}
}

func (BronzeGCPFolderIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_folder_iam_policy_bindings"},
	}
}
