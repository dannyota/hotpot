package network_discovery

import (
	"context"
	"fmt"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzehistorys1networkdiscovery"
)

// HistoryService handles history tracking for network discovery devices.
type HistoryService struct {
	entClient *ents1.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ents1.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new network discovery device.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ents1.Tx, data *NetworkDiscoveryData, now time.Time) error {
	create := tx.BronzeHistoryS1NetworkDiscovery.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetIPAddress(data.IPAddress).
		SetDomain(data.Domain).
		SetSerialNumber(data.SerialNumber).
		SetCategory(data.Category).
		SetSubCategory(data.SubCategory).
		SetResourceType(data.ResourceType).
		SetOs(data.OS).
		SetOsFamily(data.OSFamily).
		SetOsVersion(data.OSVersion).
		SetOsNameVersion(data.OSNameVersion).
		SetArchitecture(data.Architecture).
		SetManufacturer(data.Manufacturer).
		SetCPU(data.CPU).
		SetMemoryReadable(data.MemoryReadable).
		SetNetworkName(data.NetworkName).
		SetAssetStatus(data.AssetStatus).
		SetAssetCriticality(data.AssetCriticality).
		SetAssetEnvironment(data.AssetEnvironment).
		SetInfectionStatus(data.InfectionStatus).
		SetDeviceReview(data.DeviceReview).
		SetEppUnsupportedUnknown(data.EppUnsupportedUnknown).
		SetAssetContactEmail(data.AssetContactEmail).
		SetLegacyIdentityPolicyName(data.LegacyIdentityPolicyName).
		SetPreviousOsType(data.PreviousOSType).
		SetPreviousOsVersion(data.PreviousOSVersion).
		SetPreviousDeviceFunction(data.PreviousDeviceFunction).
		SetDetectedFromSite(data.DetectedFromSite).
		SetS1AccountID(data.S1AccountID).
		SetS1AccountName(data.S1AccountName).
		SetS1SiteID(data.S1SiteID).
		SetS1SiteName(data.S1SiteName).
		SetS1GroupID(data.S1GroupID).
		SetS1GroupName(data.S1GroupName).
		SetS1ScopeID(data.S1ScopeID).
		SetS1ScopeLevel(data.S1ScopeLevel).
		SetS1ScopePath(data.S1ScopePath).
		SetS1OnboardedAccountName(data.S1OnboardedAccountName).
		SetS1OnboardedGroupName(data.S1OnboardedGroupName).
		SetS1OnboardedSiteName(data.S1OnboardedSiteName).
		SetS1OnboardedScopeLevel(data.S1OnboardedScopeLevel).
		SetS1OnboardedScopePath(data.S1OnboardedScopePath).
		SetMemory(data.Memory).
		SetCoreCount(data.CoreCount).
		SetS1ManagementID(data.S1ManagementID).
		SetS1ScopeType(data.S1ScopeType).
		SetS1OnboardedAccountID(data.S1OnboardedAccountID).
		SetS1OnboardedGroupID(data.S1OnboardedGroupID).
		SetS1OnboardedScopeID(data.S1OnboardedScopeID).
		SetS1OnboardedSiteID(data.S1OnboardedSiteID).
		SetIsAdConnector(data.IsAdConnector).
		SetIsDcServer(data.IsDcServer).
		SetAdsEnabled(data.AdsEnabled)

	if data.FirstSeenDt != nil {
		create.SetFirstSeenDt(*data.FirstSeenDt)
	}
	if data.LastUpdateDt != nil {
		create.SetLastUpdateDt(*data.LastUpdateDt)
	}
	if data.LastActiveDt != nil {
		create.SetLastActiveDt(*data.LastActiveDt)
	}
	if data.LastRebootDt != nil {
		create.SetLastRebootDt(*data.LastRebootDt)
	}
	if data.S1UpdatedAt != nil {
		create.SetS1UpdatedAt(*data.S1UpdatedAt)
	}
	if data.AgentJSON != nil {
		create.SetAgentJSON(data.AgentJSON)
	}
	if data.NetworkInterfacesJSON != nil {
		create.SetNetworkInterfacesJSON(data.NetworkInterfacesJSON)
	}
	if data.AlertsJSON != nil {
		create.SetAlertsJSON(data.AlertsJSON)
	}
	if data.AlertsCountJSON != nil {
		create.SetAlertsCountJSON(data.AlertsCountJSON)
	}
	if data.DeviceReviewLogJSON != nil {
		create.SetDeviceReviewLogJSON(data.DeviceReviewLogJSON)
	}
	if data.IdentityJSON != nil {
		create.SetIdentityJSON(data.IdentityJSON)
	}
	if data.NotesJSON != nil {
		create.SetNotesJSON(data.NotesJSON)
	}
	if data.TagsJSON != nil {
		create.SetTagsJSON(data.TagsJSON)
	}
	if data.MissingCoverageJSON != nil {
		create.SetMissingCoverageJSON(data.MissingCoverageJSON)
	}
	if data.SubnetsJSON != nil {
		create.SetSubnetsJSON(data.SubnetsJSON)
	}
	if data.SurfacesJSON != nil {
		create.SetSurfacesJSON(data.SurfacesJSON)
	}
	if data.NetworkNamesJSON != nil {
		create.SetNetworkNamesJSON(data.NetworkNamesJSON)
	}
	if data.RiskFactorsJSON != nil {
		create.SetRiskFactorsJSON(data.RiskFactorsJSON)
	}
	if data.ActiveCoverageJSON != nil {
		create.SetActiveCoverageJSON(data.ActiveCoverageJSON)
	}
	if data.DiscoveryMethodsJSON != nil {
		create.SetDiscoveryMethodsJSON(data.DiscoveryMethodsJSON)
	}
	if data.HostnamesJSON != nil {
		create.SetHostnamesJSON(data.HostnamesJSON)
	}
	if data.InternalIPsJSON != nil {
		create.SetInternalIpsJSON(data.InternalIPsJSON)
	}
	if data.InternalIPsV6JSON != nil {
		create.SetInternalIpsV6JSON(data.InternalIPsV6JSON)
	}
	if data.MACAddressesJSON != nil {
		create.SetMACAddressesJSON(data.MACAddressesJSON)
	}
	if data.GatewayIPsJSON != nil {
		create.SetGatewayIpsJSON(data.GatewayIPsJSON)
	}
	if data.GatewayMacsJSON != nil {
		create.SetGatewayMacsJSON(data.GatewayMacsJSON)
	}
	if data.TCPPortsJSON != nil {
		create.SetTCPPortsJSON(data.TCPPortsJSON)
	}
	if data.UDPPortsJSON != nil {
		create.SetUDPPortsJSON(data.UDPPortsJSON)
	}
	if data.RangerTagsJSON != nil {
		create.SetRangerTagsJSON(data.RangerTagsJSON)
	}
	if data.IDSecondaryJSON != nil {
		create.SetIDSecondaryJSON(data.IDSecondaryJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create network discovery history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed network discovery device.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ents1.Tx, old *ents1.BronzeS1NetworkDiscovery, new *NetworkDiscoveryData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1NetworkDiscovery.Query().
		Where(
			bronzehistorys1networkdiscovery.ResourceID(old.ID),
			bronzehistorys1networkdiscovery.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current network discovery history: %w", err)
	}

	if err := tx.BronzeHistoryS1NetworkDiscovery.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close network discovery history: %w", err)
	}

	create := tx.BronzeHistoryS1NetworkDiscovery.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetIPAddress(new.IPAddress).
		SetDomain(new.Domain).
		SetSerialNumber(new.SerialNumber).
		SetCategory(new.Category).
		SetSubCategory(new.SubCategory).
		SetResourceType(new.ResourceType).
		SetOs(new.OS).
		SetOsFamily(new.OSFamily).
		SetOsVersion(new.OSVersion).
		SetOsNameVersion(new.OSNameVersion).
		SetArchitecture(new.Architecture).
		SetManufacturer(new.Manufacturer).
		SetCPU(new.CPU).
		SetMemoryReadable(new.MemoryReadable).
		SetNetworkName(new.NetworkName).
		SetAssetStatus(new.AssetStatus).
		SetAssetCriticality(new.AssetCriticality).
		SetAssetEnvironment(new.AssetEnvironment).
		SetInfectionStatus(new.InfectionStatus).
		SetDeviceReview(new.DeviceReview).
		SetEppUnsupportedUnknown(new.EppUnsupportedUnknown).
		SetAssetContactEmail(new.AssetContactEmail).
		SetLegacyIdentityPolicyName(new.LegacyIdentityPolicyName).
		SetPreviousOsType(new.PreviousOSType).
		SetPreviousOsVersion(new.PreviousOSVersion).
		SetPreviousDeviceFunction(new.PreviousDeviceFunction).
		SetDetectedFromSite(new.DetectedFromSite).
		SetS1AccountID(new.S1AccountID).
		SetS1AccountName(new.S1AccountName).
		SetS1SiteID(new.S1SiteID).
		SetS1SiteName(new.S1SiteName).
		SetS1GroupID(new.S1GroupID).
		SetS1GroupName(new.S1GroupName).
		SetS1ScopeID(new.S1ScopeID).
		SetS1ScopeLevel(new.S1ScopeLevel).
		SetS1ScopePath(new.S1ScopePath).
		SetS1OnboardedAccountName(new.S1OnboardedAccountName).
		SetS1OnboardedGroupName(new.S1OnboardedGroupName).
		SetS1OnboardedSiteName(new.S1OnboardedSiteName).
		SetS1OnboardedScopeLevel(new.S1OnboardedScopeLevel).
		SetS1OnboardedScopePath(new.S1OnboardedScopePath).
		SetMemory(new.Memory).
		SetCoreCount(new.CoreCount).
		SetS1ManagementID(new.S1ManagementID).
		SetS1ScopeType(new.S1ScopeType).
		SetS1OnboardedAccountID(new.S1OnboardedAccountID).
		SetS1OnboardedGroupID(new.S1OnboardedGroupID).
		SetS1OnboardedScopeID(new.S1OnboardedScopeID).
		SetS1OnboardedSiteID(new.S1OnboardedSiteID).
		SetIsAdConnector(new.IsAdConnector).
		SetIsDcServer(new.IsDcServer).
		SetAdsEnabled(new.AdsEnabled)

	if new.FirstSeenDt != nil {
		create.SetFirstSeenDt(*new.FirstSeenDt)
	}
	if new.LastUpdateDt != nil {
		create.SetLastUpdateDt(*new.LastUpdateDt)
	}
	if new.LastActiveDt != nil {
		create.SetLastActiveDt(*new.LastActiveDt)
	}
	if new.LastRebootDt != nil {
		create.SetLastRebootDt(*new.LastRebootDt)
	}
	if new.S1UpdatedAt != nil {
		create.SetS1UpdatedAt(*new.S1UpdatedAt)
	}
	if new.AgentJSON != nil {
		create.SetAgentJSON(new.AgentJSON)
	}
	if new.NetworkInterfacesJSON != nil {
		create.SetNetworkInterfacesJSON(new.NetworkInterfacesJSON)
	}
	if new.AlertsJSON != nil {
		create.SetAlertsJSON(new.AlertsJSON)
	}
	if new.AlertsCountJSON != nil {
		create.SetAlertsCountJSON(new.AlertsCountJSON)
	}
	if new.DeviceReviewLogJSON != nil {
		create.SetDeviceReviewLogJSON(new.DeviceReviewLogJSON)
	}
	if new.IdentityJSON != nil {
		create.SetIdentityJSON(new.IdentityJSON)
	}
	if new.NotesJSON != nil {
		create.SetNotesJSON(new.NotesJSON)
	}
	if new.TagsJSON != nil {
		create.SetTagsJSON(new.TagsJSON)
	}
	if new.MissingCoverageJSON != nil {
		create.SetMissingCoverageJSON(new.MissingCoverageJSON)
	}
	if new.SubnetsJSON != nil {
		create.SetSubnetsJSON(new.SubnetsJSON)
	}
	if new.SurfacesJSON != nil {
		create.SetSurfacesJSON(new.SurfacesJSON)
	}
	if new.NetworkNamesJSON != nil {
		create.SetNetworkNamesJSON(new.NetworkNamesJSON)
	}
	if new.RiskFactorsJSON != nil {
		create.SetRiskFactorsJSON(new.RiskFactorsJSON)
	}
	if new.ActiveCoverageJSON != nil {
		create.SetActiveCoverageJSON(new.ActiveCoverageJSON)
	}
	if new.DiscoveryMethodsJSON != nil {
		create.SetDiscoveryMethodsJSON(new.DiscoveryMethodsJSON)
	}
	if new.HostnamesJSON != nil {
		create.SetHostnamesJSON(new.HostnamesJSON)
	}
	if new.InternalIPsJSON != nil {
		create.SetInternalIpsJSON(new.InternalIPsJSON)
	}
	if new.InternalIPsV6JSON != nil {
		create.SetInternalIpsV6JSON(new.InternalIPsV6JSON)
	}
	if new.MACAddressesJSON != nil {
		create.SetMACAddressesJSON(new.MACAddressesJSON)
	}
	if new.GatewayIPsJSON != nil {
		create.SetGatewayIpsJSON(new.GatewayIPsJSON)
	}
	if new.GatewayMacsJSON != nil {
		create.SetGatewayMacsJSON(new.GatewayMacsJSON)
	}
	if new.TCPPortsJSON != nil {
		create.SetTCPPortsJSON(new.TCPPortsJSON)
	}
	if new.UDPPortsJSON != nil {
		create.SetUDPPortsJSON(new.UDPPortsJSON)
	}
	if new.RangerTagsJSON != nil {
		create.SetRangerTagsJSON(new.RangerTagsJSON)
	}
	if new.IDSecondaryJSON != nil {
		create.SetIDSecondaryJSON(new.IDSecondaryJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new network discovery history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted network discovery device.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ents1.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1NetworkDiscovery.Query().
		Where(
			bronzehistorys1networkdiscovery.ResourceID(resourceID),
			bronzehistorys1networkdiscovery.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ents1.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current network discovery history: %w", err)
	}

	if err := tx.BronzeHistoryS1NetworkDiscovery.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close network discovery history: %w", err)
	}

	return nil
}
