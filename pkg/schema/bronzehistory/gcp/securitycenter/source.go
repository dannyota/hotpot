package securitycenter

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPSecurityCenterSource stores historical snapshots of SCC sources.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPSecurityCenterSource struct {
	ent.Schema
}

func (BronzeHistoryGCPSecurityCenterSource) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSecurityCenterSource) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze SCC source by resource_id"),

		// All source fields (same as bronze.BronzeGCPSecurityCenterSource)
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("canonical_name").
			Optional(),
		field.String("organization_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPSecurityCenterSource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPSecurityCenterSource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_securitycenter_sources_history"},
	}
}
