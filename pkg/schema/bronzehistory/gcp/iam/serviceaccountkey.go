package iam

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPIAMServiceAccountKey stores historical snapshots of GCP IAM service account keys.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPIAMServiceAccountKey struct {
	ent.Schema
}

func (BronzeHistoryGCPIAMServiceAccountKey) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPIAMServiceAccountKey) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze service account key by resource_id"),

		// All service account key fields (same as bronze.BronzeGCPIAMServiceAccountKey)
		field.String("name").
			NotEmpty(),
		field.String("service_account_email").
			NotEmpty(),
		field.String("key_origin").
			Optional(),
		field.String("key_type").
			Optional(),
		field.String("key_algorithm").
			Optional(),
		field.Time("valid_after_time").
			Optional(),
		field.Time("valid_before_time").
			Optional(),
		field.Bool("disabled").
			Default(false),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPIAMServiceAccountKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPIAMServiceAccountKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_iam_service_account_keys_history"},
	}
}
