package securitycenter

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPSecurityCenterFinding stores historical snapshots of SCC findings.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPSecurityCenterFinding struct {
	ent.Schema
}

func (BronzeHistoryGCPSecurityCenterFinding) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPSecurityCenterFinding) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze SCC finding by resource_id"),

		// All finding fields (same as bronze.BronzeGCPSecurityCenterFinding)
		field.String("parent").
			NotEmpty(),
		field.String("resource_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("category").
			Optional(),
		field.String("external_uri").
			Optional(),
		field.String("severity").
			Optional(),
		field.String("finding_class").
			Optional(),
		field.String("canonical_name").
			Optional(),
		field.String("mute").
			Optional(),
		field.String("organization_id").
			NotEmpty(),

		// Timestamps as strings
		field.String("create_time").
			Optional(),
		field.String("event_time").
			Optional(),

		// JSON fields for nested data
		field.JSON("source_properties", json.RawMessage{}).
			Optional(),
		field.JSON("security_marks", json.RawMessage{}).
			Optional(),
		field.JSON("indicator", json.RawMessage{}).
			Optional(),
		field.JSON("vulnerability", json.RawMessage{}).
			Optional(),
		field.JSON("connections", json.RawMessage{}).
			Optional(),
		field.JSON("compliances", json.RawMessage{}).
			Optional(),
		field.JSON("contacts", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPSecurityCenterFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPSecurityCenterFinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_securitycenter_findings_history"},
	}
}
