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

// BronzeGCPContainerAnalysisOccurrence represents a Grafeas occurrence in the bronze layer.
// Fields preserve raw API response data from grafeas.ListOccurrences.
type BronzeGCPContainerAnalysisOccurrence struct {
	ent.Schema
}

func (BronzeGCPContainerAnalysisOccurrence) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPContainerAnalysisOccurrence) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Occurrence resource name (e.g., projects/123/occurrences/456)"),
		field.String("resource_uri").
			Optional().
			Comment("URI of the affected resource (e.g., https://gcr.io/project/image@sha256:abc)"),
		field.String("note_name").
			Optional().
			Comment("Analysis note associated with this occurrence (e.g., projects/123/notes/CVE-2021-1234)"),
		field.Int("kind").
			Optional().
			Comment("NoteKind enum: 0=UNSPECIFIED, 1=VULNERABILITY, 2=BUILD, 3=IMAGE, 4=PACKAGE, 5=DEPLOYMENT, 6=DISCOVERY, 7=ATTESTATION, 8=UPGRADE"),
		field.String("remediation").
			Optional().
			Comment("Actions that can be taken to remedy the note"),

		// Timestamps as strings
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),

		// Occurrence detail fields (oneof in proto, stored as separate optional JSON)
		field.JSON("vulnerability_json", json.RawMessage{}).
			Optional().
			Comment("VulnerabilityOccurrence details"),
		field.JSON("build_json", json.RawMessage{}).
			Optional().
			Comment("BuildOccurrence details"),
		field.JSON("image_json", json.RawMessage{}).
			Optional().
			Comment("ImageOccurrence details"),
		field.JSON("package_json", json.RawMessage{}).
			Optional().
			Comment("PackageOccurrence details"),
		field.JSON("deployment_json", json.RawMessage{}).
			Optional().
			Comment("DeploymentOccurrence details"),
		field.JSON("discovery_json", json.RawMessage{}).
			Optional().
			Comment("DiscoveryOccurrence details"),
		field.JSON("attestation_json", json.RawMessage{}).
			Optional().
			Comment("AttestationOccurrence details"),
		field.JSON("upgrade_json", json.RawMessage{}).
			Optional().
			Comment("UpgradeOccurrence details"),
		field.JSON("compliance_json", json.RawMessage{}).
			Optional().
			Comment("ComplianceOccurrence details"),
		field.JSON("dsse_attestation_json", json.RawMessage{}).
			Optional().
			Comment("DSSEAttestationOccurrence details"),
		field.JSON("sbom_reference_json", json.RawMessage{}).
			Optional().
			Comment("SBOMReferenceOccurrence details"),
		field.JSON("envelope_json", json.RawMessage{}).
			Optional().
			Comment("DSSE envelope"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPContainerAnalysisOccurrence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("project_id"),
		index.Fields("kind"),
		index.Fields("note_name"),
		index.Fields("resource_uri"),
	}
}

func (BronzeGCPContainerAnalysisOccurrence) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_containeranalysis_occurrences"},
	}
}
