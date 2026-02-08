package iam

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPIAMServiceAccount represents a GCP IAM service account in the bronze layer.
// Fields preserve raw API response data from iam.projects.serviceAccounts.list.
type BronzeGCPIAMServiceAccount struct {
	ent.Schema
}

func (BronzeGCPIAMServiceAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPIAMServiceAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Unique ID of the service account"),
		field.String("name").
			NotEmpty().
			Comment("Resource name of the service account"),
		field.String("email").
			NotEmpty().
			Comment("Email address of the service account"),
		field.String("display_name").
			Optional().
			Comment("User-specified display name"),
		field.String("description").
			Optional().
			Comment("User-specified description"),
		field.String("oauth2_client_id").
			Optional().
			Comment("OAuth2 client ID for the service account"),
		field.Bool("disabled").
			Default(false).
			Comment("Whether the service account is disabled"),
		field.String("etag").
			Optional().
			Comment("Entity tag for the service account"),
		field.String("project_id").
			NotEmpty().
			Comment("Project ID where the service account exists"),
	}
}

func (BronzeGCPIAMServiceAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPIAMServiceAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_iam_service_accounts"},
	}
}
