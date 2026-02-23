package ranger_setting

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// RangerSettingDiff represents changes between old and new ranger setting states.
type RangerSettingDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffRangerSettingData compares old Ent entity and new data.
func DiffRangerSettingData(old *ent.BronzeS1RangerSetting, new *RangerSettingData) *RangerSettingDiff {
	if old == nil {
		return &RangerSettingDiff{IsNew: true}
	}

	return &RangerSettingDiff{
		IsChanged: old.AccountID != new.AccountID ||
			old.ScopeID != new.ScopeID ||
			old.Enabled != new.Enabled ||
			old.UsePeriodicSnapshots != new.UsePeriodicSnapshots ||
			old.SnapshotPeriod != new.SnapshotPeriod ||
			old.NetworkDecommissionValue != new.NetworkDecommissionValue ||
			old.MinAgentsInNetworkToScan != new.MinAgentsInNetworkToScan ||
			old.TCPPortScan != new.TCPPortScan ||
			old.UDPPortScan != new.UDPPortScan ||
			old.IcmpScan != new.ICMPScan ||
			old.SmbScan != new.SMBScan ||
			old.MdnsScan != new.MDNSScan ||
			old.RdnsScan != new.RDNSScan ||
			old.SnmpScan != new.SNMPScan ||
			old.MultiScanSsdp != new.MultiScanSSDP ||
			old.UseFullDNSScan != new.UseFullDNSScan ||
			old.ScanOnlyLocalSubnets != new.ScanOnlyLocalSubnets ||
			old.AutoEnableNetworks != new.AutoEnableNetworks ||
			old.CombineDevices != new.CombineDevices ||
			old.NewNetworkInHours != new.NewNetworkInHours ||
			!bytes.Equal(old.TCPPortsJSON, new.TCPPortsJSON) ||
			!bytes.Equal(old.UDPPortsJSON, new.UDPPortsJSON) ||
			!bytes.Equal(old.RestrictionsJSON, new.RestrictionsJSON),
	}
}
