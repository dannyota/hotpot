package orgpolicy

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPOrgPolicyPolicy stores historical snapshots of organization policies.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPOrgPolicyPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPOrgPolicyPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPOrgPolicyPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze org policy by resource_id"),

		// All policy fields (same as bronze.BronzeGCPOrgPolicyPolicy)
		field.String("etag").
			Optional(),
		field.JSON("spec", map[string]any{}).
			Optional(),
		field.JSON("dry_run_spec", map[string]any{}).
			Optional(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPOrgPolicyPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPOrgPolicyPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_orgpolicy_policies_history"},
	}
}
