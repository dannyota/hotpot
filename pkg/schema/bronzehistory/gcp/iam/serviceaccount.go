package iam

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BronzeHistoryGCPIAMServiceAccount stores historical snapshots of GCP IAM service accounts.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPIAMServiceAccount struct {
	ent.Schema
}

func (BronzeHistoryGCPIAMServiceAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze service account by resource_id"),
		field.Time("valid_from").
			Immutable().
			Comment("Start of validity period"),
		field.Time("valid_to").
			Optional().
			Nillable().
			Comment("End of validity period (null = current)"),
		field.Time("collected_at").
			Comment("Timestamp when this snapshot was collected"),

		// All service account fields (same as bronze.BronzeGCPIAMServiceAccount)
		field.String("name").
			NotEmpty(),
		field.String("email").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("oauth2_client_id").
			Optional(),
		field.Bool("disabled").
			Default(false),
		field.String("etag").
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPIAMServiceAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPIAMServiceAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_iam_service_accounts_history"},
	}
}
