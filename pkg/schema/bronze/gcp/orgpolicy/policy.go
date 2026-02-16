package orgpolicy

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPOrgPolicyPolicy represents an Organization Policy in the bronze layer.
// Fields preserve raw API response data from orgpolicy.ListPolicies.
type BronzeGCPOrgPolicyPolicy struct {
	ent.Schema
}

func (BronzeGCPOrgPolicyPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPOrgPolicyPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Policy resource name (e.g., organizations/123/policies/compute.disableSerialPortAccess)"),
		field.String("etag").
			Optional(),
		field.JSON("spec", map[string]any{}).
			Optional().
			Comment("Policy spec as JSON (contains rules, inherit_from_parent, reset)"),
		field.JSON("dry_run_spec", map[string]any{}).
			Optional().
			Comment("Dry run policy spec as JSON"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPOrgPolicyPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
	}
}

func (BronzeGCPOrgPolicyPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_orgpolicy_policies"},
	}
}
