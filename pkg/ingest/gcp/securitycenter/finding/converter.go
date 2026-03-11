package finding

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// FindingData holds converted SCC finding data ready for Ent insertion.
type FindingData struct {
	ID               string
	Parent           string
	ResourceName     string
	State            string
	Category         string
	ExternalURI      string
	Severity         string
	FindingClass     string
	CanonicalName    string
	Mute             string
	OrganizationID   string
	CreateTime       string
	EventTime        string
	SourceProperties json.RawMessage
	SecurityMarks    json.RawMessage
	Indicator        json.RawMessage
	Vulnerability    json.RawMessage
	Connections      json.RawMessage
	Compliances      json.RawMessage
	Contacts         json.RawMessage
	CollectedAt      time.Time
}

// ConvertFinding converts a raw GCP API SCC finding to Ent-compatible data.
func ConvertFinding(orgName string, sourceName string, result *securitycenterpb.ListFindingsResponse_ListFindingsResult, collectedAt time.Time) *FindingData {
	finding := result.GetFinding()
	if finding == nil {
		return nil
	}

	data := &FindingData{
		ID:             finding.GetName(),
		Parent:         finding.GetParent(),
		ResourceName:   finding.GetResourceName(),
		State:          finding.GetState().String(),
		Category:       finding.GetCategory(),
		ExternalURI:    finding.GetExternalUri(),
		Severity:       finding.GetSeverity().String(),
		FindingClass:   finding.GetFindingClass().String(),
		CanonicalName:  finding.GetCanonicalName(),
		Mute:           finding.GetMute().String(),
		OrganizationID: orgName,
		CollectedAt:    collectedAt,
	}

	// Convert timestamps
	if finding.GetCreateTime() != nil {
		data.CreateTime = finding.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if finding.GetEventTime() != nil {
		data.EventTime = finding.GetEventTime().AsTime().Format(time.RFC3339)
	}

	// Marshal nested proto fields to JSON
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	if props := finding.GetSourceProperties(); len(props) > 0 {
		if b, err := json.Marshal(props); err == nil {
			data.SourceProperties = b
		}
	}
	if marks := finding.GetSecurityMarks(); marks != nil {
		if b, err := marshaler.Marshal(marks); err == nil {
			data.SecurityMarks = b
		}
	}
	if ind := finding.GetIndicator(); ind != nil {
		if b, err := marshaler.Marshal(ind); err == nil {
			data.Indicator = b
		}
	}
	if vuln := finding.GetVulnerability(); vuln != nil {
		if b, err := marshaler.Marshal(vuln); err == nil {
			data.Vulnerability = b
		}
	}
	if conns := finding.GetConnections(); len(conns) > 0 {
		if b, err := json.Marshal(conns); err == nil {
			data.Connections = b
		}
	}
	if comps := finding.GetCompliances(); len(comps) > 0 {
		if b, err := json.Marshal(comps); err == nil {
			data.Compliances = b
		}
	}
	if contacts := finding.GetContacts(); len(contacts) > 0 {
		if b, err := json.Marshal(contacts); err == nil {
			data.Contacts = b
		}
	}

	return data
}
