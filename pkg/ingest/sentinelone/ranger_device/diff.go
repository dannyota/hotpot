package ranger_device

import (
	"bytes"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
)

// RangerDeviceDiff represents changes between old and new ranger device states.
type RangerDeviceDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffRangerDeviceData compares old Ent entity and new data.
func DiffRangerDeviceData(old *ents1.BronzeS1RangerDevice, new *RangerDeviceData) *RangerDeviceDiff {
	if old == nil {
		return &RangerDeviceDiff{IsNew: true}
	}

	return &RangerDeviceDiff{
		IsChanged: old.LocalIP != new.LocalIP ||
			old.ExternalIP != new.ExternalIP ||
			old.MACAddress != new.MACAddress ||
			old.OsType != new.OsType ||
			old.OsName != new.OsName ||
			old.OsVersion != new.OsVersion ||
			old.DeviceType != new.DeviceType ||
			old.DeviceFunction != new.DeviceFunction ||
			old.Manufacturer != new.Manufacturer ||
			old.ManagedState != new.ManagedState ||
			old.AgentID != new.AgentID ||
			!timeEqual(old.FirstSeen, new.FirstSeen) ||
			!timeEqual(old.LastSeen, new.LastSeen) ||
			old.SubnetAddress != new.SubnetAddress ||
			old.GatewayIPAddress != new.GatewayIPAddress ||
			old.GatewayMACAddress != new.GatewayMACAddress ||
			old.NetworkName != new.NetworkName ||
			old.Domain != new.Domain ||
			old.SiteName != new.SiteName ||
			old.DeviceReview != new.DeviceReview ||
			old.HasIdentity != new.HasIdentity ||
			old.HasUserLabel != new.HasUserLabel ||
			old.FingerprintScore != new.FingerprintScore ||
			!bytes.Equal(old.TCPPortsJSON, new.TCPPortsJSON) ||
			!bytes.Equal(old.UDPPortsJSON, new.UDPPortsJSON) ||
			!bytes.Equal(old.HostnamesJSON, new.HostnamesJSON) ||
			!bytes.Equal(old.DiscoveryMethodsJSON, new.DiscoveryMethodsJSON) ||
			!bytes.Equal(old.NetworksJSON, new.NetworksJSON) ||
			!bytes.Equal(old.TagsJSON, new.TagsJSON),
	}
}

func timeEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
