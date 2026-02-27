package network_discovery

import (
	"bytes"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

// NetworkDiscoveryDiff represents changes between old and new network discovery device states.
type NetworkDiscoveryDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffNetworkDiscoveryData compares old Ent entity and new data.
func DiffNetworkDiscoveryData(old *ents1.BronzeS1NetworkDiscovery, new *NetworkDiscoveryData) *NetworkDiscoveryDiff {
	if old == nil {
		return &NetworkDiscoveryDiff{IsNew: true}
	}

	return &NetworkDiscoveryDiff{
		IsChanged: old.Name != new.Name ||
			old.IPAddress != new.IPAddress ||
			old.Domain != new.Domain ||
			old.SerialNumber != new.SerialNumber ||
			old.Category != new.Category ||
			old.SubCategory != new.SubCategory ||
			old.ResourceType != new.ResourceType ||
			old.Os != new.OS ||
			old.OsFamily != new.OSFamily ||
			old.OsVersion != new.OSVersion ||
			old.OsNameVersion != new.OSNameVersion ||
			old.Architecture != new.Architecture ||
			old.Manufacturer != new.Manufacturer ||
			old.CPU != new.CPU ||
			old.MemoryReadable != new.MemoryReadable ||
			old.NetworkName != new.NetworkName ||
			old.AssetStatus != new.AssetStatus ||
			old.AssetCriticality != new.AssetCriticality ||
			old.AssetEnvironment != new.AssetEnvironment ||
			old.InfectionStatus != new.InfectionStatus ||
			old.DeviceReview != new.DeviceReview ||
			old.EppUnsupportedUnknown != new.EppUnsupportedUnknown ||
			old.AssetContactEmail != new.AssetContactEmail ||
			old.LegacyIdentityPolicyName != new.LegacyIdentityPolicyName ||
			old.PreviousOsType != new.PreviousOSType ||
			old.PreviousOsVersion != new.PreviousOSVersion ||
			old.PreviousDeviceFunction != new.PreviousDeviceFunction ||
			old.DetectedFromSite != new.DetectedFromSite ||
			old.S1AccountID != new.S1AccountID ||
			old.S1AccountName != new.S1AccountName ||
			old.S1SiteID != new.S1SiteID ||
			old.S1SiteName != new.S1SiteName ||
			old.S1GroupID != new.S1GroupID ||
			old.S1GroupName != new.S1GroupName ||
			old.S1ScopeID != new.S1ScopeID ||
			old.S1ScopeLevel != new.S1ScopeLevel ||
			old.S1ScopePath != new.S1ScopePath ||
			old.S1OnboardedAccountName != new.S1OnboardedAccountName ||
			old.S1OnboardedGroupName != new.S1OnboardedGroupName ||
			old.S1OnboardedSiteName != new.S1OnboardedSiteName ||
			old.S1OnboardedScopeLevel != new.S1OnboardedScopeLevel ||
			old.S1OnboardedScopePath != new.S1OnboardedScopePath ||
			old.Memory != new.Memory ||
			old.CoreCount != new.CoreCount ||
			old.S1ManagementID != new.S1ManagementID ||
			old.S1ScopeType != new.S1ScopeType ||
			old.S1OnboardedAccountID != new.S1OnboardedAccountID ||
			old.S1OnboardedGroupID != new.S1OnboardedGroupID ||
			old.S1OnboardedScopeID != new.S1OnboardedScopeID ||
			old.S1OnboardedSiteID != new.S1OnboardedSiteID ||
			old.IsAdConnector != new.IsAdConnector ||
			old.IsDcServer != new.IsDcServer ||
			old.AdsEnabled != new.AdsEnabled ||
			!timeEqual(old.FirstSeenDt, new.FirstSeenDt) ||
			!timeEqual(old.LastUpdateDt, new.LastUpdateDt) ||
			!timeEqual(old.LastActiveDt, new.LastActiveDt) ||
			!timeEqual(old.LastRebootDt, new.LastRebootDt) ||
			!timeEqual(old.S1UpdatedAt, new.S1UpdatedAt) ||
			!bytes.Equal(old.AgentJSON, new.AgentJSON) ||
			!bytes.Equal(old.NetworkInterfacesJSON, new.NetworkInterfacesJSON) ||
			!bytes.Equal(old.AlertsJSON, new.AlertsJSON) ||
			!bytes.Equal(old.AlertsCountJSON, new.AlertsCountJSON) ||
			!bytes.Equal(old.DeviceReviewLogJSON, new.DeviceReviewLogJSON) ||
			!bytes.Equal(old.IdentityJSON, new.IdentityJSON) ||
			!bytes.Equal(old.NotesJSON, new.NotesJSON) ||
			!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
			!bytes.Equal(old.MissingCoverageJSON, new.MissingCoverageJSON) ||
			!bytes.Equal(old.SubnetsJSON, new.SubnetsJSON) ||
			!bytes.Equal(old.SurfacesJSON, new.SurfacesJSON) ||
			!bytes.Equal(old.NetworkNamesJSON, new.NetworkNamesJSON) ||
			!bytes.Equal(old.RiskFactorsJSON, new.RiskFactorsJSON) ||
			!bytes.Equal(old.ActiveCoverageJSON, new.ActiveCoverageJSON) ||
			!bytes.Equal(old.DiscoveryMethodsJSON, new.DiscoveryMethodsJSON) ||
			!bytes.Equal(old.HostnamesJSON, new.HostnamesJSON) ||
			!bytes.Equal(old.InternalIpsJSON, new.InternalIPsJSON) ||
			!bytes.Equal(old.InternalIpsV6JSON, new.InternalIPsV6JSON) ||
			!bytes.Equal(old.MACAddressesJSON, new.MACAddressesJSON) ||
			!bytes.Equal(old.GatewayIpsJSON, new.GatewayIPsJSON) ||
			!bytes.Equal(old.GatewayMacsJSON, new.GatewayMacsJSON) ||
			!bytes.Equal(old.TCPPortsJSON, new.TCPPortsJSON) ||
			!bytes.Equal(old.UDPPortsJSON, new.UDPPortsJSON) ||
			!bytes.Equal(old.RangerTagsJSON, new.RangerTagsJSON) ||
			!bytes.Equal(old.IDSecondaryJSON, new.IDSecondaryJSON),
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
