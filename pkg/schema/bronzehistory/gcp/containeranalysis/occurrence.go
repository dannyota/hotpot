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

// BronzeHistoryGCPContainerAnalysisOccurrence stores historical snapshots of Grafeas occurrences.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPContainerAnalysisOccurrence struct {
	ent.Schema
}

func (BronzeHistoryGCPContainerAnalysisOccurrence) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPContainerAnalysisOccurrence) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze occurrence by resource_id"),

		// All occurrence fields (same as bronze.BronzeGCPContainerAnalysisOccurrence)
		field.String("resource_uri").
			Optional(),
		field.String("note_name").
			Optional(),
		field.Int("kind").
			Optional(),
		field.String("remediation").
			Optional(),

		// Timestamps as strings
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),

		// Occurrence detail fields
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
		field.JSON("sbom_reference_json", json.RawMessage{}).
			Optional(),
		field.JSON("envelope_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPContainerAnalysisOccurrence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
		index.Fields("note_name"),
	}
}

func (BronzeHistoryGCPContainerAnalysisOccurrence) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_containeranalysis_occurrences_history"},
	}
}
