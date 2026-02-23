package ranger_gateway

import (
	"bytes"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// RangerGatewayDiff represents changes between old and new gateway states.
type RangerGatewayDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffRangerGatewayData compares old Ent entity and new data.
func DiffRangerGatewayData(old *ent.BronzeS1RangerGateway, new *RangerGatewayData) *RangerGatewayDiff {
	if old == nil {
		return &RangerGatewayDiff{IsNew: true}
	}

	return &RangerGatewayDiff{
		IsChanged: old.IP != new.IP ||
			old.MACAddress != new.MacAddress ||
			old.ExternalIP != new.ExternalIP ||
			old.Manufacturer != new.Manufacturer ||
			old.NetworkName != new.NetworkName ||
			old.AccountID != new.AccountID ||
			old.AccountName != new.AccountName ||
			old.SiteID != new.SiteID ||
			old.NumberOfAgents != new.NumberOfAgents ||
			old.NumberOfRangers != new.NumberOfRangers ||
			old.ConnectedRangers != new.ConnectedRangers ||
			old.TotalAgents != new.TotalAgents ||
			old.AgentPercentage != new.AgentPercentage ||
			old.AllowScan != new.AllowScan ||
			old.Archived != new.Archived ||
			old.NewNetwork != new.NewNetwork ||
			old.InheritSettings != new.InheritSettings ||
			old.TCPPortScan != new.TCPPortScan ||
			old.UDPPortScan != new.UDPPortScan ||
			old.IcmpScan != new.ICMPScan ||
			old.SmbScan != new.SMBScan ||
			old.MdnsScan != new.MDNSScan ||
			old.RdnsScan != new.RDNSScan ||
			old.SnmpScan != new.SNMPScan ||
			old.ScanOnlyLocalSubnets != new.ScanOnlyLocalSubnets ||
			!timeEqual(old.CreatedAtAPI, new.CreatedAtAPI) ||
			!timeEqual(old.ExpiryDate, new.ExpiryDate) ||
			!bytes.Equal(old.TCPPortsJSON, new.TCPPortsJSON) ||
			!bytes.Equal(old.UDPPortsJSON, new.UDPPortsJSON) ||
			!bytes.Equal(old.RestrictionsJSON, new.RestrictionsJSON),
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
