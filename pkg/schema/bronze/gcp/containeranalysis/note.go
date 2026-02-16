package containeranalysis

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPContainerAnalysisNote represents a Grafeas note in the bronze layer.
// Fields preserve raw API response data from grafeas.ListNotes.
type BronzeGCPContainerAnalysisNote struct {
	ent.Schema
}

func (BronzeGCPContainerAnalysisNote) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPContainerAnalysisNote) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Note resource name (e.g., projects/123/notes/CVE-2021-1234)"),
		field.String("short_description").
			Optional().
			Comment("A one sentence description of this note"),
		field.String("long_description").
			Optional().
			Comment("A detailed description of this note"),
		field.Int("kind").
			Optional().
			Comment("NoteKind enum: 0=UNSPECIFIED, 1=VULNERABILITY, 2=BUILD, 3=IMAGE, 4=PACKAGE, 5=DEPLOYMENT, 6=DISCOVERY, 7=ATTESTATION, 8=UPGRADE"),

		// Timestamps as strings
		field.String("expiration_time").
			Optional().
			Comment("Time of expiration for this note"),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),

		// JSON fields for nested/repeated data
		field.JSON("related_url_json", json.RawMessage{}).
			Optional().
			Comment("URLs associated with this note"),
		field.JSON("related_note_names", json.RawMessage{}).
			Optional().
			Comment("Other notes related to this note"),

		// Note type-specific detail fields (oneof in proto, stored as separate optional JSON)
		field.JSON("vulnerability_json", json.RawMessage{}).
			Optional().
			Comment("VulnerabilityNote details"),
		field.JSON("build_json", json.RawMessage{}).
			Optional().
			Comment("BuildNote details"),
		field.JSON("image_json", json.RawMessage{}).
			Optional().
			Comment("ImageNote details"),
		field.JSON("package_json", json.RawMessage{}).
			Optional().
			Comment("PackageNote details"),
		field.JSON("deployment_json", json.RawMessage{}).
			Optional().
			Comment("DeploymentNote details"),
		field.JSON("discovery_json", json.RawMessage{}).
			Optional().
			Comment("DiscoveryNote details"),
		field.JSON("attestation_json", json.RawMessage{}).
			Optional().
			Comment("AttestationNote details"),
		field.JSON("upgrade_json", json.RawMessage{}).
			Optional().
			Comment("UpgradeNote details"),
		field.JSON("compliance_json", json.RawMessage{}).
			Optional().
			Comment("ComplianceNote details"),
		field.JSON("dsse_attestation_json", json.RawMessage{}).
			Optional().
			Comment("DSSEAttestationNote details"),
		field.JSON("vulnerability_assessment_json", json.RawMessage{}).
			Optional().
			Comment("VulnerabilityAssessmentNote details"),
		field.JSON("sbom_reference_json", json.RawMessage{}).
			Optional().
			Comment("SBOMReferenceNote details"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPContainerAnalysisNote) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("project_id"),
		index.Fields("kind"),
	}
}

func (BronzeGCPContainerAnalysisNote) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_containeranalysis_notes"},
	}
}
