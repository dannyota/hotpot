package interconnect

import (
	"time"

	networkv2 "danny.vn/gnode/services/network/v2"
)

// InterconnectData represents a converted interconnect ready for Ent insertion.
type InterconnectData struct {
	UUID        string
	Name        string
	Description string
	Status      string
	EnableGw2   bool
	CircuitID   int
	Gw01IP      string
	Gw02IP      string
	GwVIP       string
	RemoteGw01IP string
	RemoteGw02IP string
	PackageID   string
	TypeID      string
	TypeName    string
	CreatedAt   string
	Region      string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertInterconnect converts a GreenNode SDK Interconnect to InterconnectData.
func ConvertInterconnect(ic *networkv2.Interconnect, projectID, region string, collectedAt time.Time) *InterconnectData {
	return &InterconnectData{
		UUID:         ic.UUID,
		Name:         ic.Name,
		Description:  ic.Description,
		Status:       ic.Status,
		EnableGw2:    ic.EnableGw2,
		CircuitID:    ic.CircuitID,
		Gw01IP:       ic.Gw01IP,
		Gw02IP:       ic.Gw02IP,
		GwVIP:        ic.GwVIP,
		RemoteGw01IP: ic.RemoteGw01IP,
		RemoteGw02IP: ic.RemoteGw02IP,
		PackageID:    ic.PackageID,
		TypeID:       ic.TypeID,
		TypeName:     ic.TypeName,
		CreatedAt:    ic.CreatedAt,
		Region:       region,
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}
}
