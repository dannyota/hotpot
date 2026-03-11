package attestor

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/binaryauthorization/apiv1/binaryauthorizationpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// AttestorData holds converted Binary Authorization attestor data ready for Ent insertion.
type AttestorData struct {
	ID                       string
	Description              string
	UserOwnedGrafeasNoteJSON json.RawMessage
	UpdateTime               string
	Etag                     string
	ProjectID                string
	CollectedAt              time.Time
}

// ConvertAttestor converts a raw GCP API Binary Authorization attestor to Ent-compatible data.
func ConvertAttestor(a *binaryauthorizationpb.Attestor, projectID string, collectedAt time.Time) *AttestorData {
	data := &AttestorData{
		ID:          a.GetName(),
		Description: a.GetDescription(),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if ts := a.GetUpdateTime(); ts != nil {
		data.UpdateTime = ts.AsTime().Format(time.RFC3339)
	}

	if note := a.GetUserOwnedGrafeasNote(); note != nil {
		marshaler := protojson.MarshalOptions{UseProtoNames: true}
		if b, err := marshaler.Marshal(note); err == nil {
			data.UserOwnedGrafeasNoteJSON = b
		}
	}

	return data
}
