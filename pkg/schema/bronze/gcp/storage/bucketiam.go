package storage

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

// BronzeGCPStorageBucketIamPolicy represents a GCP Storage bucket IAM policy in the bronze layer.
type BronzeGCPStorageBucketIamPolicy struct {
	ent.Schema
}

func (BronzeGCPStorageBucketIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPStorageBucketIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Bucket name used as resource ID"),
		field.String("bucket_name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPStorageBucketIamPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("bindings", BronzeGCPStorageBucketIamPolicyBinding.Type),
	}
}

func (BronzeGCPStorageBucketIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPStorageBucketIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_bucket_iam_policies"},
	}
}

// BronzeGCPStorageBucketIamPolicyBinding represents an IAM policy binding for a GCP Storage bucket.
type BronzeGCPStorageBucketIamPolicyBinding struct {
	ent.Schema
}

func (BronzeGCPStorageBucketIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("role").
			NotEmpty(),
		field.JSON("members_json", json.RawMessage{}).
			Optional(),
		field.JSON("condition_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPStorageBucketIamPolicyBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("policy", BronzeGCPStorageBucketIamPolicy.Type).
			Ref("bindings").
			Unique().
			Required(),
	}
}

func (BronzeGCPStorageBucketIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_bucket_iam_policy_bindings"},
	}
}
