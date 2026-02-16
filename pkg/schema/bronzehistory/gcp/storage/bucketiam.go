package storage

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPStorageBucketIamPolicy stores historical snapshots of GCP Storage bucket IAM policies.
type BronzeHistoryGCPStorageBucketIamPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPStorageBucketIamPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPStorageBucketIamPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze bucket IAM policy by resource_id"),

		// IAM policy fields
		field.String("bucket_name").
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

func (BronzeHistoryGCPStorageBucketIamPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPStorageBucketIamPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_bucket_iam_policies_history"},
	}
}

// BronzeHistoryGCPStorageBucketIamPolicyBinding stores historical snapshots of bucket IAM policy bindings.
type BronzeHistoryGCPStorageBucketIamPolicyBinding struct {
	ent.Schema
}

func (BronzeHistoryGCPStorageBucketIamPolicyBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("policy_history_id").
			Comment("Links to parent BronzeHistoryGCPStorageBucketIamPolicy"),
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

func (BronzeHistoryGCPStorageBucketIamPolicyBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("policy_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPStorageBucketIamPolicyBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_storage_bucket_iam_policy_bindings_history"},
	}
}
