package orgpolicy

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPOrgPolicyConstraint represents an Organization Policy constraint in the bronze layer.
// Fields preserve raw API response data from orgpolicy.ListConstraints.
type BronzeGCPOrgPolicyConstraint struct {
	ent.Schema
}

func (BronzeGCPOrgPolicyConstraint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPOrgPolicyConstraint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Constraint resource name (e.g., organizations/123/constraints/compute.disableSerialPortAccess)"),
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.Int("constraint_default").
			Default(0).
			Comment("Default enforcement behavior (0=UNSPECIFIED, 1=ALLOW, 2=DENY)"),
		field.Bool("supports_dry_run").
			Default(false),
		field.Bool("supports_simulation").
			Default(false),
		field.JSON("list_constraint", map[string]any{}).
			Optional().
			Comment("List constraint details as JSON"),
		field.JSON("boolean_constraint", map[string]any{}).
			Optional().
			Comment("Boolean constraint details as JSON"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPOrgPolicyConstraint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
	}
}

func (BronzeGCPOrgPolicyConstraint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_orgpolicy_constraints"},
	}
}
