package securitycenter

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPSecurityCenterFinding represents a Security Command Center finding in the bronze layer.
// Fields preserve raw API response data from securitycenter.ListFindings.
type BronzeGCPSecurityCenterFinding struct {
	ent.Schema
}

func (BronzeGCPSecurityCenterFinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPSecurityCenterFinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Finding resource name (e.g., organizations/123/sources/456/findings/789)"),
		field.String("parent").
			NotEmpty().
			Comment("Source resource name"),
		field.String("resource_name").
			Optional().
			Comment("Full resource name of the affected resource"),
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
			NotEmpty().
			Comment("Organization resource name"),

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

func (BronzeGCPSecurityCenterFinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
		index.Fields("state"),
		index.Fields("severity"),
		index.Fields("category"),
	}
}

func (BronzeGCPSecurityCenterFinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_securitycenter_findings"},
	}
}
