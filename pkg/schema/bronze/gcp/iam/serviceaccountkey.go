package iam

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPIAMServiceAccountKey represents a GCP IAM service account key in the bronze layer.
// Fields preserve raw API response data from iam.projects.serviceAccounts.keys.list.
type BronzeGCPIAMServiceAccountKey struct {
	ent.Schema
}

func (BronzeGCPIAMServiceAccountKey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPIAMServiceAccountKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Unique ID of the service account key (short form)"),
		field.String("name").
			NotEmpty().
			Comment("Full resource name of the service account key"),
		field.String("service_account_email").
			NotEmpty().
			Comment("Email address of the associated service account"),
		field.String("key_origin").
			Optional().
			Comment("Origin of the key (e.g., GOOGLE_PROVIDED, USER_PROVIDED)"),
		field.String("key_type").
			Optional().
			Comment("Type of the key (e.g., USER_MANAGED, SYSTEM_MANAGED)"),
		field.String("key_algorithm").
			Optional().
			Comment("Algorithm used for the key"),
		field.Time("valid_after_time").
			Optional().
			Comment("Key is valid after this time"),
		field.Time("valid_before_time").
			Optional().
			Comment("Key is valid before this time"),
		field.Bool("disabled").
			Default(false).
			Comment("Whether the key is disabled"),
		field.String("project_id").
			NotEmpty().
			Comment("Project ID where the service account key exists"),
	}
}

func (BronzeGCPIAMServiceAccountKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("service_account_email"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPIAMServiceAccountKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_iam_service_account_keys"},
	}
}
