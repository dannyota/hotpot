package orgpolicy

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPOrgPolicyCustomConstraint represents an Organization Policy custom constraint in the bronze layer.
// Fields preserve raw API response data from orgpolicy.ListCustomConstraints.
type BronzeGCPOrgPolicyCustomConstraint struct {
	ent.Schema
}

func (BronzeGCPOrgPolicyCustomConstraint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPOrgPolicyCustomConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Custom constraint resource name (e.g., organizations/123/customConstraints/custom.createOnlyE2TypeVms)"),
		field.JSON("resource_types", []string{}).
			Optional().
			Comment("Resource instance types on which this custom constraint applies"),
		field.JSON("method_types", []int{}).
			Optional().
			Comment("Operations being applied (0=UNSPECIFIED, 1=CREATE, 2=UPDATE, 3=DELETE, 4=REMOVE_GRANT, 5=GOVERN_TAGS)"),
		field.String("condition").
			Optional().
			Comment("CEL condition used in evaluation of the constraint"),
		field.Int("action_type").
			Default(0).
			Comment("Allow or deny type (0=UNSPECIFIED, 1=ALLOW, 2=DENY)"),
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.Time("update_time").
			Optional().
			Comment("Last time this custom constraint was updated"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPOrgPolicyCustomConstraint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
	}
}

func (BronzeGCPOrgPolicyCustomConstraint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_orgpolicy_custom_constraints"},
	}
}
