package ranger_device

import (
	"encoding/json"
	"time"
)

// RangerDeviceData holds converted ranger device data ready for Ent insertion.
type RangerDeviceData struct {
	ResourceID           string
	LocalIP              string
	ExternalIP           string
	MACAddress           string
	OsType               string
	OsName               string
	OsVersion            string
	DeviceType           string
	DeviceFunction       string
	Manufacturer         string
	ManagedState         string
	AgentID              string
	FirstSeen            *time.Time
	LastSeen             *time.Time
	SubnetAddress        string
	GatewayIPAddress     string
	GatewayMACAddress    string
	NetworkName          string
	Domain               string
	SiteName             string
	DeviceReview         string
	HasIdentity          bool
	HasUserLabel         bool
	FingerprintScore     int
	TCPPortsJSON         json.RawMessage
	UDPPortsJSON         json.RawMessage
	HostnamesJSON        json.RawMessage
	DiscoveryMethodsJSON json.RawMessage
	NetworksJSON         json.RawMessage
	TagsJSON             json.RawMessage
	CollectedAt          time.Time
}

// ConvertRangerDevice converts an API ranger device to RangerDeviceData.
func ConvertRangerDevice(d APIRangerDevice, collectedAt time.Time) *RangerDeviceData {
	return &RangerDeviceData{
		ResourceID:           d.ID,
		LocalIP:              d.LocalIP,
		ExternalIP:           d.ExternalIP,
		MACAddress:           d.MacAddress,
		OsType:               d.OsType,
		OsName:               d.OsName,
		OsVersion:            d.OsVersion,
		DeviceType:           d.DeviceType,
		DeviceFunction:       d.DeviceFunction,
		Manufacturer:         d.Manufacturer,
		ManagedState:         d.ManagedState,
		AgentID:              d.AgentID,
		FirstSeen:            d.FirstSeen,
		LastSeen:             d.LastSeen,
		SubnetAddress:        d.SubnetAddress,
		GatewayIPAddress:     d.GatewayIPAddress,
		GatewayMACAddress:    d.GatewayMacAddress,
		NetworkName:          d.NetworkName,
		Domain:               d.Domain,
		SiteName:             d.SiteName,
		DeviceReview:         d.DeviceReview,
		HasIdentity:          d.HasIdentity,
		HasUserLabel:         d.HasUserLabel,
		FingerprintScore:     d.FingerprintScore,
		TCPPortsJSON:         d.TCPPorts,
		UDPPortsJSON:         d.UDPPorts,
		HostnamesJSON:        d.Hostnames,
		DiscoveryMethodsJSON: d.DiscoveryMethods,
		NetworksJSON:         d.Networks,
		TagsJSON:             d.Tags,
		CollectedAt:          collectedAt,
	}
}
