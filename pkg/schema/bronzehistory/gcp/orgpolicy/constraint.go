package orgpolicy

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPOrgPolicyConstraint stores historical snapshots of organization policy constraints.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPOrgPolicyConstraint struct {
	ent.Schema
}

func (BronzeHistoryGCPOrgPolicyConstraint) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPOrgPolicyConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze org policy constraint by resource_id"),

		// All constraint fields (same as bronze.BronzeGCPOrgPolicyConstraint)
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.Int("constraint_default").
			Default(0),
		field.Bool("supports_dry_run").
			Default(false),
		field.Bool("supports_simulation").
			Default(false),
		field.JSON("list_constraint", map[string]any{}).
			Optional(),
		field.JSON("boolean_constraint", map[string]any{}).
			Optional(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPOrgPolicyConstraint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPOrgPolicyConstraint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_orgpolicy_constraints_history"},
	}
}
