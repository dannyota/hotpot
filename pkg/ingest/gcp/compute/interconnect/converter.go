package interconnect

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// InterconnectData holds converted interconnect data ready for Ent insertion.
type InterconnectData struct {
	ID                   string
	Name                 string
	Description          string
	SelfLink             string
	Location             string
	InterconnectType     string
	LinkType             string
	AdminEnabled         bool
	OperationalStatus    string
	ProvisionedLinkCount int
	RequestedLinkCount   int
	PeerIPAddress        string
	GoogleIPAddress      string
	GoogleReferenceID    string
	NocContactEmail      string
	CustomerName         string
	State                string
	CreationTimestamp    string
	ExpectedOutagesJSON  json.RawMessage
	CircuitInfosJSON     json.RawMessage
	ProjectID            string
	CollectedAt          time.Time
}

// ConvertInterconnect converts a GCP API Interconnect to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertInterconnect(ic *computepb.Interconnect, projectID string, collectedAt time.Time) (*InterconnectData, error) {
	data := &InterconnectData{
		ID:                   fmt.Sprintf("%d", ic.GetId()),
		Name:                 ic.GetName(),
		Description:          ic.GetDescription(),
		SelfLink:             ic.GetSelfLink(),
		Location:             ic.GetLocation(),
		InterconnectType:     ic.GetInterconnectType(),
		LinkType:             ic.GetLinkType(),
		AdminEnabled:         ic.GetAdminEnabled(),
		OperationalStatus:    ic.GetOperationalStatus(),
		ProvisionedLinkCount: int(ic.GetProvisionedLinkCount()),
		RequestedLinkCount:   int(ic.GetRequestedLinkCount()),
		PeerIPAddress:        ic.GetPeerIpAddress(),
		GoogleIPAddress:      ic.GetGoogleIpAddress(),
		GoogleReferenceID:    ic.GetGoogleReferenceId(),
		NocContactEmail:      ic.GetNocContactEmail(),
		CustomerName:         ic.GetCustomerName(),
		State:                ic.GetState(),
		CreationTimestamp:    ic.GetCreationTimestamp(),
		ProjectID:            projectID,
		CollectedAt:          collectedAt,
	}

	// Marshal expected outages
	if outages := ic.GetExpectedOutages(); len(outages) > 0 {
		outagesBytes, err := json.Marshal(outages)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal expected outages for %s: %w", ic.GetName(), err)
		}
		data.ExpectedOutagesJSON = outagesBytes
	}

	// Marshal circuit infos
	if circuits := ic.GetCircuitInfos(); len(circuits) > 0 {
		circuitsBytes, err := json.Marshal(circuits)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal circuit infos for %s: %w", ic.GetName(), err)
		}
		data.CircuitInfosJSON = circuitsBytes
	}

	return data, nil
}
