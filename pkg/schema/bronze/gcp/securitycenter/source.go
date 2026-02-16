package securitycenter

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPSecurityCenterSource represents a Security Command Center source in the bronze layer.
// Fields preserve raw API response data from securitycenter.ListSources.
type BronzeGCPSecurityCenterSource struct {
	ent.Schema
}

func (BronzeGCPSecurityCenterSource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSecurityCenterSource) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Source resource name (e.g., organizations/123/sources/456)"),
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("canonical_name").
			Optional(),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name"),
	}
}

func (BronzeGCPSecurityCenterSource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
	}
}

func (BronzeGCPSecurityCenterSource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_securitycenter_sources"},
	}
}
