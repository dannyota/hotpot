package note

import (
	"encoding/json"
	"time"

	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// NoteData holds converted Grafeas note data ready for Ent insertion.
type NoteData struct {
	ID               string
	ShortDescription string
	LongDescription  string
	Kind             int
	ExpirationTime   string
	CreateTime       string
	UpdateTime       string
	ProjectID        string
	CollectedAt      time.Time

	// JSON fields
	RelatedURLJSON              json.RawMessage
	RelatedNoteNames            json.RawMessage
	VulnerabilityJSON           json.RawMessage
	BuildJSON                   json.RawMessage
	ImageJSON                   json.RawMessage
	PackageJSON                 json.RawMessage
	DeploymentJSON              json.RawMessage
	DiscoveryJSON               json.RawMessage
	AttestationJSON             json.RawMessage
	UpgradeJSON                 json.RawMessage
	ComplianceJSON              json.RawMessage
	DsseAttestationJSON         json.RawMessage
	VulnerabilityAssessmentJSON json.RawMessage
	SbomReferenceJSON           json.RawMessage
}

// ConvertNote converts a raw Grafeas API note to Ent-compatible data.
func ConvertNote(note *grafeaspb.Note, projectID string, collectedAt time.Time) *NoteData {
	data := &NoteData{
		ID:               note.GetName(),
		ShortDescription: note.GetShortDescription(),
		LongDescription:  note.GetLongDescription(),
		Kind:             int(note.GetKind()),
		ProjectID:        projectID,
		CollectedAt:      collectedAt,
	}

	if note.GetExpirationTime() != nil {
		data.ExpirationTime = note.GetExpirationTime().AsTime().Format(time.RFC3339)
	}
	if note.GetCreateTime() != nil {
		data.CreateTime = note.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if note.GetUpdateTime() != nil {
		data.UpdateTime = note.GetUpdateTime().AsTime().Format(time.RFC3339)
	}

	// Related URLs
	if urls := note.GetRelatedUrl(); len(urls) > 0 {
		data.RelatedURLJSON, _ = json.Marshal(urls)
	}

	// Related note names
	if names := note.GetRelatedNoteNames(); len(names) > 0 {
		data.RelatedNoteNames, _ = json.Marshal(names)
	}

	// Type-specific detail fields (oneof)
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	if v := note.GetVulnerability(); v != nil {
		data.VulnerabilityJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetBuild(); v != nil {
		data.BuildJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetImage(); v != nil {
		data.ImageJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetPackage(); v != nil {
		data.PackageJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetDeployment(); v != nil {
		data.DeploymentJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetDiscovery(); v != nil {
		data.DiscoveryJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetAttestation(); v != nil {
		data.AttestationJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetUpgrade(); v != nil {
		data.UpgradeJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetCompliance(); v != nil {
		data.ComplianceJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetDsseAttestation(); v != nil {
		data.DsseAttestationJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetVulnerabilityAssessment(); v != nil {
		data.VulnerabilityAssessmentJSON, _ = marshaler.Marshal(v)
	}
	if v := note.GetSbomReference(); v != nil {
		data.SbomReferenceJSON, _ = marshaler.Marshal(v)
	}

	return data
}
