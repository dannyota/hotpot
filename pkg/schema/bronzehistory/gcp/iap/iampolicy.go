package iap

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPIAPIAMPolicy stores historical snapshots of IAP IAM policies.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPIAPIAMPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPIAPIAMPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPIAPIAMPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze IAP IAM policy by resource_id"),

		// All IAM policy fields (same as bronze.BronzeGCPIAPIAMPolicy)
		field.String("name").
			NotEmpty(),
		field.String("etag").
			Optional(),
		field.Int("version").
			Optional(),
		field.JSON("bindings_json", json.RawMessage{}).
			Optional(),
		field.JSON("audit_configs_json", json.RawMessage{}).
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPIAPIAMPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPIAPIAMPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_iap_iam_policies_history"},
	}
}
