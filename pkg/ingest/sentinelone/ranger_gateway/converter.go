package ranger_gateway

import (
	"encoding/json"
	"time"
)

// RangerGatewayData holds converted gateway data ready for Ent insertion.
type RangerGatewayData struct {
	ResourceID           string
	IP                   string
	MacAddress           string
	ExternalIP           string
	Manufacturer         string
	NetworkName          string
	AccountID            string
	AccountName          string
	SiteID               string
	NumberOfAgents       int
	NumberOfRangers      int
	ConnectedRangers     int
	TotalAgents          int
	AgentPercentage      float64
	AllowScan            bool
	Archived             bool
	NewNetwork           bool
	InheritSettings      bool
	TCPPortScan          bool
	UDPPortScan          bool
	ICMPScan             bool
	SMBScan              bool
	MDNSScan             bool
	RDNSScan             bool
	SNMPScan             bool
	ScanOnlyLocalSubnets bool
	CreatedAtAPI         *time.Time
	ExpiryDate           *time.Time
	TCPPortsJSON         json.RawMessage
	UDPPortsJSON         json.RawMessage
	RestrictionsJSON     json.RawMessage
	CollectedAt          time.Time
}

// ConvertRangerGateway converts an API gateway to RangerGatewayData.
func ConvertRangerGateway(g APIRangerGateway, collectedAt time.Time) *RangerGatewayData {
	return &RangerGatewayData{
		ResourceID:           g.ID,
		IP:                   g.IP,
		MacAddress:           g.MacAddress,
		ExternalIP:           g.ExternalIP,
		Manufacturer:         g.Manufacturer,
		NetworkName:          g.NetworkName,
		AccountID:            g.AccountID.String(),
		AccountName:          g.AccountName,
		SiteID:               g.SiteID.String(),
		NumberOfAgents:       g.NumberOfAgents,
		NumberOfRangers:      g.NumberOfRangers,
		ConnectedRangers:     g.ConnectedRangers,
		TotalAgents:          g.TotalAgents,
		AgentPercentage:      g.AgentPercentage,
		AllowScan:            g.AllowScan,
		Archived:             g.Archived,
		NewNetwork:           g.New,
		InheritSettings:      g.InheritSettings,
		TCPPortScan:          g.TCPPortScan,
		UDPPortScan:          g.UDPPortScan,
		ICMPScan:             g.ICMPScan,
		SMBScan:              g.SMBScan,
		MDNSScan:             g.MDNSScan,
		RDNSScan:             g.RDNSScan,
		SNMPScan:             g.SNMPScan,
		ScanOnlyLocalSubnets: g.ScanOnlyLocalSubnets,
		CreatedAtAPI:         g.CreatedAt,
		ExpiryDate:           g.ExpiryDate,
		TCPPortsJSON:         g.TCPPorts,
		UDPPortsJSON:         g.UDPPorts,
		RestrictionsJSON:     g.Restrictions,
		CollectedAt:          collectedAt,
	}
}
