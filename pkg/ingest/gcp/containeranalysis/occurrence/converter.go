package occurrence

import (
	"encoding/json"
	"time"

	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// OccurrenceData holds converted Grafeas occurrence data ready for Ent insertion.
type OccurrenceData struct {
	ID          string
	ResourceURI string
	NoteName    string
	Kind        int
	Remediation string
	CreateTime  string
	UpdateTime  string
	ProjectID   string
	CollectedAt time.Time

	// JSON fields
	VulnerabilityJSON   json.RawMessage
	BuildJSON           json.RawMessage
	ImageJSON           json.RawMessage
	PackageJSON         json.RawMessage
	DeploymentJSON      json.RawMessage
	DiscoveryJSON       json.RawMessage
	AttestationJSON     json.RawMessage
	UpgradeJSON         json.RawMessage
	ComplianceJSON      json.RawMessage
	DsseAttestationJSON json.RawMessage
	SbomReferenceJSON   json.RawMessage
	EnvelopeJSON        json.RawMessage
}

// ConvertOccurrence converts a raw Grafeas API occurrence to Ent-compatible data.
func ConvertOccurrence(occ *grafeaspb.Occurrence, projectID string, collectedAt time.Time) *OccurrenceData {
	data := &OccurrenceData{
		ID:          occ.GetName(),
		ResourceURI: occ.GetResourceUri(),
		NoteName:    occ.GetNoteName(),
		Kind:        int(occ.GetKind()),
		Remediation: occ.GetRemediation(),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if occ.GetCreateTime() != nil {
		data.CreateTime = occ.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if occ.GetUpdateTime() != nil {
		data.UpdateTime = occ.GetUpdateTime().AsTime().Format(time.RFC3339)
	}

	// Type-specific detail fields (oneof)
	marshaler := protojson.MarshalOptions{UseProtoNames: true}
	if v := occ.GetVulnerability(); v != nil {
		data.VulnerabilityJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetBuild(); v != nil {
		data.BuildJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetImage(); v != nil {
		data.ImageJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetPackage(); v != nil {
		data.PackageJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetDeployment(); v != nil {
		data.DeploymentJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetDiscovery(); v != nil {
		data.DiscoveryJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetAttestation(); v != nil {
		data.AttestationJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetUpgrade(); v != nil {
		data.UpgradeJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetCompliance(); v != nil {
		data.ComplianceJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetDsseAttestation(); v != nil {
		data.DsseAttestationJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetSbomReference(); v != nil {
		data.SbomReferenceJSON, _ = marshaler.Marshal(v)
	}
	if v := occ.GetEnvelope(); v != nil {
		data.EnvelopeJSON, _ = marshaler.Marshal(v)
	}

	return data
}
