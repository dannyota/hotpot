package containeranalysis

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPContainerAnalysisNote stores historical snapshots of Grafeas notes.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPContainerAnalysisNote struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerAnalysisNote) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPContainerAnalysisNote) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze note by resource_id"),

		// All note fields (same as bronze.BronzeGCPContainerAnalysisNote)
		field.String("short_description").
			Optional(),
		field.String("long_description").
			Optional(),
		field.Int("kind").
			Optional(),

		// Timestamps as strings
		field.String("expiration_time").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),

		// JSON fields for nested/repeated data
		field.JSON("related_url_json", json.RawMessage{}).
			Optional(),
		field.JSON("related_note_names", json.RawMessage{}).
			Optional(),

		// Note type-specific detail fields
		field.JSON("vulnerability_json", json.RawMessage{}).
			Optional(),
		field.JSON("build_json", json.RawMessage{}).
			Optional(),
		field.JSON("image_json", json.RawMessage{}).
			Optional(),
		field.JSON("package_json", json.RawMessage{}).
			Optional(),
		field.JSON("deployment_json", json.RawMessage{}).
			Optional(),
		field.JSON("discovery_json", json.RawMessage{}).
			Optional(),
		field.JSON("attestation_json", json.RawMessage{}).
			Optional(),
		field.JSON("upgrade_json", json.RawMessage{}).
			Optional(),
		field.JSON("compliance_json", json.RawMessage{}).
			Optional(),
		field.JSON("dsse_attestation_json", json.RawMessage{}).
			Optional(),
		field.JSON("vulnerability_assessment_json", json.RawMessage{}).
			Optional(),
		field.JSON("sbom_reference_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPContainerAnalysisNote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPContainerAnalysisNote) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_containeranalysis_notes_history"},
	}
}
