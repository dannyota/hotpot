package ranger_setting

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzes1rangersetting"
)

// Service handles SentinelOne Ranger setting ingestion.
type Service struct {
	client    *Client
	entClient *ents1.Client
	history   *HistoryService
}

// NewService creates a new ranger setting ingestion service.
func NewService(client *Client, entClient *ents1.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of ranger setting ingestion.
type IngestResult struct {
	SettingCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Ranger settings for all accounts from SentinelOne.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	accounts, err := s.entClient.BronzeS1Account.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}

	var allSettings []*RangerSettingData

	for _, account := range accounts {
		apiSetting, err := s.client.GetSettings(account.ID)
		if err != nil {
			return nil, fmt.Errorf("get ranger settings for account %s: %w", account.ID, err)
		}

		allSettings = append(allSettings, ConvertRangerSetting(*apiSetting, account.ID, collectedAt))

		if heartbeat != nil {
			heartbeat()
		}
	}

	if err := s.saveSettings(ctx, allSettings); err != nil {
		return nil, fmt.Errorf("save ranger settings: %w", err)
	}

	return &IngestResult{
		SettingCount:   len(allSettings),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSettings(ctx context.Context, settings []*RangerSettingData) error {
	if len(settings) == 0 {
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

	activeIDs := make(map[string]struct{}, len(settings))

	for _, data := range settings {
		existing, err := tx.BronzeS1RangerSetting.Query().
			Where(bronzes1rangersetting.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ents1.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing ranger setting %s: %w", data.ResourceID, err)
		}

		diff := DiffRangerSettingData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1RangerSetting.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for ranger setting %s: %w", data.ResourceID, err)
			}
			activeIDs[data.ResourceID] = struct{}{}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1RangerSetting.Create().
				SetID(data.ResourceID).
				SetAccountID(data.AccountID).
				SetScopeID(data.ScopeID).
				SetEnabled(data.Enabled).
				SetUsePeriodicSnapshots(data.UsePeriodicSnapshots).
				SetSnapshotPeriod(data.SnapshotPeriod).
				SetNetworkDecommissionValue(data.NetworkDecommissionValue).
				SetMinAgentsInNetworkToScan(data.MinAgentsInNetworkToScan).
				SetTCPPortScan(data.TCPPortScan).
				SetUDPPortScan(data.UDPPortScan).
				SetIcmpScan(data.ICMPScan).
				SetSmbScan(data.SMBScan).
				SetMdnsScan(data.MDNSScan).
				SetRdnsScan(data.RDNSScan).
				SetSnmpScan(data.SNMPScan).
				SetMultiScanSsdp(data.MultiScanSSDP).
				SetUseFullDNSScan(data.UseFullDNSScan).
				SetScanOnlyLocalSubnets(data.ScanOnlyLocalSubnets).
				SetAutoEnableNetworks(data.AutoEnableNetworks).
				SetCombineDevices(data.CombineDevices).
				SetNewNetworkInHours(data.NewNetworkInHours).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.TCPPortsJSON != nil {
				create.SetTCPPortsJSON(data.TCPPortsJSON)
			}
			if data.UDPPortsJSON != nil {
				create.SetUDPPortsJSON(data.UDPPortsJSON)
			}
			if data.RestrictionsJSON != nil {
				create.SetRestrictionsJSON(data.RestrictionsJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create ranger setting %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for ranger setting %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1RangerSetting.UpdateOneID(data.ResourceID).
				SetAccountID(data.AccountID).
				SetScopeID(data.ScopeID).
				SetEnabled(data.Enabled).
				SetUsePeriodicSnapshots(data.UsePeriodicSnapshots).
				SetSnapshotPeriod(data.SnapshotPeriod).
				SetNetworkDecommissionValue(data.NetworkDecommissionValue).
				SetMinAgentsInNetworkToScan(data.MinAgentsInNetworkToScan).
				SetTCPPortScan(data.TCPPortScan).
				SetUDPPortScan(data.UDPPortScan).
				SetIcmpScan(data.ICMPScan).
				SetSmbScan(data.SMBScan).
				SetMdnsScan(data.MDNSScan).
				SetRdnsScan(data.RDNSScan).
				SetSnmpScan(data.SNMPScan).
				SetMultiScanSsdp(data.MultiScanSSDP).
				SetUseFullDNSScan(data.UseFullDNSScan).
				SetScanOnlyLocalSubnets(data.ScanOnlyLocalSubnets).
				SetAutoEnableNetworks(data.AutoEnableNetworks).
				SetCombineDevices(data.CombineDevices).
				SetNewNetworkInHours(data.NewNetworkInHours).
				SetCollectedAt(data.CollectedAt)

			if data.TCPPortsJSON != nil {
				update.SetTCPPortsJSON(data.TCPPortsJSON)
			}
			if data.UDPPortsJSON != nil {
				update.SetUDPPortsJSON(data.UDPPortsJSON)
			}
			if data.RestrictionsJSON != nil {
				update.SetRestrictionsJSON(data.RestrictionsJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update ranger setting %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for ranger setting %s: %w", data.ResourceID, err)
			}
		}

		activeIDs[data.ResourceID] = struct{}{}
	}

	// Delete stale ranger settings not returned by the API.
	allDBIDs, err := tx.BronzeS1RangerSetting.Query().
		Select(bronzes1rangersetting.FieldID).
		Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query all ranger setting IDs: %w", err)
	}

	staleCount := 0
	for _, id := range allDBIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}

		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale ranger setting %s: %w", id, err)
		}

		if err := tx.BronzeS1RangerSetting.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale ranger setting %s: %w", id, err)
		}
		staleCount++
	}

	if staleCount > 0 {
		slog.Info("s1 ranger settings: deleted stale", "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

