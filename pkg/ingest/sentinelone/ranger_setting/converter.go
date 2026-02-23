package ranger_setting

import (
	"encoding/json"
	"time"
)

// RangerSettingData holds converted ranger setting data ready for Ent insertion.
type RangerSettingData struct {
	ResourceID               string
	AccountID                string
	ScopeID                  string
	Enabled                  bool
	UsePeriodicSnapshots     bool
	SnapshotPeriod           int
	NetworkDecommissionValue int
	MinAgentsInNetworkToScan int
	TCPPortScan              bool
	UDPPortScan              bool
	ICMPScan                 bool
	SMBScan                  bool
	MDNSScan                 bool
	RDNSScan                 bool
	SNMPScan                 bool
	MultiScanSSDP            bool
	UseFullDNSScan           bool
	ScanOnlyLocalSubnets     bool
	AutoEnableNetworks       bool
	CombineDevices           bool
	NewNetworkInHours        int
	TCPPortsJSON             json.RawMessage
	UDPPortsJSON             json.RawMessage
	RestrictionsJSON         json.RawMessage
	CollectedAt              time.Time
}

// ConvertRangerSetting converts an API ranger setting to RangerSettingData.
func ConvertRangerSetting(s APIRangerSetting, accountID string, collectedAt time.Time) *RangerSettingData {
	return &RangerSettingData{
		ResourceID:               accountID,
		AccountID:                s.AccountID.String(),
		ScopeID:                  s.ScopeID,
		Enabled:                  s.Enabled,
		UsePeriodicSnapshots:     s.UsePeriodicSnapshots,
		SnapshotPeriod:           s.SnapshotPeriod,
		NetworkDecommissionValue: s.NetworkDecommissionValue,
		MinAgentsInNetworkToScan: s.MinAgentsInNetworkToScan,
		TCPPortScan:              s.TCPPortScan,
		UDPPortScan:              s.UDPPortScan,
		ICMPScan:                 s.ICMPScan,
		SMBScan:                  s.SMBScan,
		MDNSScan:                 s.MDNSScan,
		RDNSScan:                 s.RDNSScan,
		SNMPScan:                 s.SNMPScan,
		MultiScanSSDP:            s.MultiScanSSDP,
		UseFullDNSScan:           s.UseFullDNSScan,
		ScanOnlyLocalSubnets:     s.ScanOnlyLocalSubnets,
		AutoEnableNetworks:       s.AutoEnableNetworks,
		CombineDevices:           s.CombineDevices,
		NewNetworkInHours:        s.NewNetworkInHours,
		TCPPortsJSON:             s.TCPPorts,
		UDPPortsJSON:             s.UDPPorts,
		RestrictionsJSON:         s.Restrictions,
		CollectedAt:              collectedAt,
	}
}
