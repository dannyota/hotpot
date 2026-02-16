package orgpolicy

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPOrgPolicyCustomConstraint stores historical snapshots of organization policy custom constraints.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPOrgPolicyCustomConstraint struct {
	ent.Schema
}

func (BronzeHistoryGCPOrgPolicyCustomConstraint) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPOrgPolicyCustomConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze org policy custom constraint by resource_id"),

		// All custom constraint fields (same as bronze.BronzeGCPOrgPolicyCustomConstraint)
		field.JSON("resource_types", []string{}).
			Optional().
			Comment("Resource instance types on which this custom constraint applies"),
		field.JSON("method_types", []int{}).
			Optional().
			Comment("Operations being applied (0=UNSPECIFIED, 1=CREATE, 2=UPDATE, 3=DELETE, 4=REMOVE_GRANT, 5=GOVERN_TAGS)"),
		field.String("condition").
			Optional(),
		field.Int("action_type").
			Default(0),
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.Time("update_time").
			Optional(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPOrgPolicyCustomConstraint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPOrgPolicyCustomConstraint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_orgpolicy_custom_constraints_history"},
	}
}
