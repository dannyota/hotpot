package ranger_device

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzes1rangerdevice"
)

// Service handles SentinelOne ranger device ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new ranger device ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of ranger device ingestion.
type IngestResult struct {
	DeviceCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all ranger devices from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allDevices []*RangerDeviceData
	cursor := ""
	batchNum := 0

	for {
		batchNum++
		batch, err := s.client.GetDevicesBatch(cursor)
		if err != nil {
			slog.Error("s1 ranger devices batch failed", "batch", batchNum, "totalSoFar", len(allDevices), "error", err)
			return nil, fmt.Errorf("get ranger devices batch: %w", err)
		}

		for _, apiDevice := range batch.Devices {
			allDevices = append(allDevices, ConvertRangerDevice(apiDevice, collectedAt))
		}

		slog.Info("s1 ranger devices batch fetched", "batch", batchNum, "batchItems", len(batch.Devices), "totalItems", len(allDevices), "hasMore", batch.HasMore)

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveDevices(ctx, allDevices); err != nil {
		return nil, fmt.Errorf("save ranger devices: %w", err)
	}

	return &IngestResult{
		DeviceCount:    len(allDevices),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveDevices(ctx context.Context, devices []*RangerDeviceData) error {
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

	for _, data := range devices {
		existing, err := tx.BronzeS1RangerDevice.Query().
			Where(bronzes1rangerdevice.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing ranger device %s: %w", data.ResourceID, err)
		}

		diff := DiffRangerDeviceData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1RangerDevice.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for ranger device %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1RangerDevice.Create().
				SetID(data.ResourceID).
				SetLocalIP(data.LocalIP).
				SetExternalIP(data.ExternalIP).
				SetMACAddress(data.MACAddress).
				SetOsType(data.OsType).
				SetOsName(data.OsName).
				SetOsVersion(data.OsVersion).
				SetDeviceType(data.DeviceType).
				SetDeviceFunction(data.DeviceFunction).
				SetManufacturer(data.Manufacturer).
				SetManagedState(data.ManagedState).
				SetAgentID(data.AgentID).
				SetSubnetAddress(data.SubnetAddress).
				SetGatewayIPAddress(data.GatewayIPAddress).
				SetGatewayMACAddress(data.GatewayMACAddress).
				SetNetworkName(data.NetworkName).
				SetDomain(data.Domain).
				SetSiteName(data.SiteName).
				SetDeviceReview(data.DeviceReview).
				SetHasIdentity(data.HasIdentity).
				SetHasUserLabel(data.HasUserLabel).
				SetFingerprintScore(data.FingerprintScore).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.FirstSeen != nil {
				create.SetFirstSeen(*data.FirstSeen)
			}
			if data.LastSeen != nil {
				create.SetLastSeen(*data.LastSeen)
			}
			if data.TCPPortsJSON != nil {
				create.SetTCPPortsJSON(data.TCPPortsJSON)
			}
			if data.UDPPortsJSON != nil {
				create.SetUDPPortsJSON(data.UDPPortsJSON)
			}
			if data.HostnamesJSON != nil {
				create.SetHostnamesJSON(data.HostnamesJSON)
			}
			if data.DiscoveryMethodsJSON != nil {
				create.SetDiscoveryMethodsJSON(data.DiscoveryMethodsJSON)
			}
			if data.NetworksJSON != nil {
				create.SetNetworksJSON(data.NetworksJSON)
			}
			if data.TagsJSON != nil {
				create.SetTagsJSON(data.TagsJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create ranger device %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for ranger device %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1RangerDevice.UpdateOneID(data.ResourceID).
				SetLocalIP(data.LocalIP).
				SetExternalIP(data.ExternalIP).
				SetMACAddress(data.MACAddress).
				SetOsType(data.OsType).
				SetOsName(data.OsName).
				SetOsVersion(data.OsVersion).
				SetDeviceType(data.DeviceType).
				SetDeviceFunction(data.DeviceFunction).
				SetManufacturer(data.Manufacturer).
				SetManagedState(data.ManagedState).
				SetAgentID(data.AgentID).
				SetSubnetAddress(data.SubnetAddress).
				SetGatewayIPAddress(data.GatewayIPAddress).
				SetGatewayMACAddress(data.GatewayMACAddress).
				SetNetworkName(data.NetworkName).
				SetDomain(data.Domain).
				SetSiteName(data.SiteName).
				SetDeviceReview(data.DeviceReview).
				SetHasIdentity(data.HasIdentity).
				SetHasUserLabel(data.HasUserLabel).
				SetFingerprintScore(data.FingerprintScore).
				SetCollectedAt(data.CollectedAt)

			if data.FirstSeen != nil {
				update.SetFirstSeen(*data.FirstSeen)
			}
			if data.LastSeen != nil {
				update.SetLastSeen(*data.LastSeen)
			}
			if data.TCPPortsJSON != nil {
				update.SetTCPPortsJSON(data.TCPPortsJSON)
			}
			if data.UDPPortsJSON != nil {
				update.SetUDPPortsJSON(data.UDPPortsJSON)
			}
			if data.HostnamesJSON != nil {
				update.SetHostnamesJSON(data.HostnamesJSON)
			}
			if data.DiscoveryMethodsJSON != nil {
				update.SetDiscoveryMethodsJSON(data.DiscoveryMethodsJSON)
			}
			if data.NetworksJSON != nil {
				update.SetNetworksJSON(data.NetworksJSON)
			}
			if data.TagsJSON != nil {
				update.SetTagsJSON(data.TagsJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update ranger device %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for ranger device %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes ranger devices that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeS1RangerDevice.Query().
		Where(bronzes1rangerdevice.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, d := range stale {
		if err := s.history.CloseHistory(ctx, tx, d.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for ranger device %s: %w", d.ID, err)
		}

		if err := tx.BronzeS1RangerDevice.DeleteOne(d).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete ranger device %s: %w", d.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
