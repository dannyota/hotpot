package network_discovery

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzes1networkdiscovery"
)

// Service handles SentinelOne network discovery device ingestion.
type Service struct {
	client    *Client
	entClient *ents1.Client
	history   *HistoryService
}

// NewService creates a new network discovery ingestion service.
func NewService(client *Client, entClient *ents1.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of network discovery device ingestion.
type IngestResult struct {
	DeviceCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all network discovery devices from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	totalExpected, err := s.client.GetCount()
	if err != nil {
		slog.Warn("s1 network discovery: failed to get count, continuing without total", "error", err)
	}

	var allDevices []*NetworkDiscoveryData
	cursor := ""
	batchNum := 0

	for {
		batchNum++
		batch, err := s.client.GetDevicesBatch(cursor)
		if err != nil {
			slog.Error("s1 network discovery batch failed", "batch", batchNum, "totalSoFar", len(allDevices), "error", err)
			return nil, fmt.Errorf("get network discovery batch: %w", err)
		}

		for _, apiDevice := range batch.Devices {
			allDevices = append(allDevices, ConvertNetworkDiscovery(apiDevice, collectedAt))
		}

		slog.Info("s1 network discovery batch fetched", "batch", batchNum, "batchItems", len(batch.Devices), "totalFetched", len(allDevices), "totalExpected", totalExpected, "hasMore", batch.HasMore)

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveDevices(ctx, allDevices); err != nil {
		return nil, fmt.Errorf("save network discovery devices: %w", err)
	}

	return &IngestResult{
		DeviceCount:    len(allDevices),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveDevices(ctx context.Context, devices []*NetworkDiscoveryData) error {
	if len(devices) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	activeIDs := make(map[string]struct{}, len(devices))

	for _, data := range devices {
		existing, err := tx.BronzeS1NetworkDiscovery.Query().
			Where(bronzes1networkdiscovery.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ents1.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing network discovery device %s: %w", data.ResourceID, err)
		}

		diff := DiffNetworkDiscoveryData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1NetworkDiscovery.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for network discovery device %s: %w", data.ResourceID, err)
			}
			activeIDs[data.ResourceID] = struct{}{}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1NetworkDiscovery.Create().
				SetID(data.ResourceID).
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
				SetAdsEnabled(data.AdsEnabled).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

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
				tx.Rollback()
				return fmt.Errorf("create network discovery device %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for network discovery device %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1NetworkDiscovery.UpdateOneID(data.ResourceID).
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
				SetAdsEnabled(data.AdsEnabled).
				SetCollectedAt(data.CollectedAt)

			if data.FirstSeenDt != nil {
				update.SetFirstSeenDt(*data.FirstSeenDt)
			}
			if data.LastUpdateDt != nil {
				update.SetLastUpdateDt(*data.LastUpdateDt)
			}
			if data.LastActiveDt != nil {
				update.SetLastActiveDt(*data.LastActiveDt)
			}
			if data.LastRebootDt != nil {
				update.SetLastRebootDt(*data.LastRebootDt)
			}
			if data.S1UpdatedAt != nil {
				update.SetS1UpdatedAt(*data.S1UpdatedAt)
			}
			if data.AgentJSON != nil {
				update.SetAgentJSON(data.AgentJSON)
			}
			if data.NetworkInterfacesJSON != nil {
				update.SetNetworkInterfacesJSON(data.NetworkInterfacesJSON)
			}
			if data.AlertsJSON != nil {
				update.SetAlertsJSON(data.AlertsJSON)
			}
			if data.AlertsCountJSON != nil {
				update.SetAlertsCountJSON(data.AlertsCountJSON)
			}
			if data.DeviceReviewLogJSON != nil {
				update.SetDeviceReviewLogJSON(data.DeviceReviewLogJSON)
			}
			if data.IdentityJSON != nil {
				update.SetIdentityJSON(data.IdentityJSON)
			}
			if data.NotesJSON != nil {
				update.SetNotesJSON(data.NotesJSON)
			}
			if data.TagsJSON != nil {
				update.SetTagsJSON(data.TagsJSON)
			}
			if data.MissingCoverageJSON != nil {
				update.SetMissingCoverageJSON(data.MissingCoverageJSON)
			}
			if data.SubnetsJSON != nil {
				update.SetSubnetsJSON(data.SubnetsJSON)
			}
			if data.SurfacesJSON != nil {
				update.SetSurfacesJSON(data.SurfacesJSON)
			}
			if data.NetworkNamesJSON != nil {
				update.SetNetworkNamesJSON(data.NetworkNamesJSON)
			}
			if data.RiskFactorsJSON != nil {
				update.SetRiskFactorsJSON(data.RiskFactorsJSON)
			}
			if data.ActiveCoverageJSON != nil {
				update.SetActiveCoverageJSON(data.ActiveCoverageJSON)
			}
			if data.DiscoveryMethodsJSON != nil {
				update.SetDiscoveryMethodsJSON(data.DiscoveryMethodsJSON)
			}
			if data.HostnamesJSON != nil {
				update.SetHostnamesJSON(data.HostnamesJSON)
			}
			if data.InternalIPsJSON != nil {
				update.SetInternalIpsJSON(data.InternalIPsJSON)
			}
			if data.InternalIPsV6JSON != nil {
				update.SetInternalIpsV6JSON(data.InternalIPsV6JSON)
			}
			if data.MACAddressesJSON != nil {
				update.SetMACAddressesJSON(data.MACAddressesJSON)
			}
			if data.GatewayIPsJSON != nil {
				update.SetGatewayIpsJSON(data.GatewayIPsJSON)
			}
			if data.GatewayMacsJSON != nil {
				update.SetGatewayMacsJSON(data.GatewayMacsJSON)
			}
			if data.TCPPortsJSON != nil {
				update.SetTCPPortsJSON(data.TCPPortsJSON)
			}
			if data.UDPPortsJSON != nil {
				update.SetUDPPortsJSON(data.UDPPortsJSON)
			}
			if data.RangerTagsJSON != nil {
				update.SetRangerTagsJSON(data.RangerTagsJSON)
			}
			if data.IDSecondaryJSON != nil {
				update.SetIDSecondaryJSON(data.IDSecondaryJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update network discovery device %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for network discovery device %s: %w", data.ResourceID, err)
			}
		}

		activeIDs[data.ResourceID] = struct{}{}
	}

	allDBIDs, err := tx.BronzeS1NetworkDiscovery.Query().Select(bronzes1networkdiscovery.FieldID).Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query all network discovery device IDs: %w", err)
	}

	staleCount := 0
	for _, id := range allDBIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}
		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale network discovery device %s: %w", id, err)
		}
		if err := tx.BronzeS1NetworkDiscovery.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale network discovery device %s: %w", id, err)
		}
		staleCount++
	}
	if staleCount > 0 {
		slog.Info("s1 network discovery devices: deleted stale", "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
