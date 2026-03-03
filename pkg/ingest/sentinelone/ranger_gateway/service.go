package ranger_gateway

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzes1rangergateway"
)

// Service handles SentinelOne ranger gateway ingestion.
type Service struct {
	client    *Client
	entClient *ents1.Client
	history   *HistoryService
}

// NewService creates a new ranger gateway ingestion service.
func NewService(client *Client, entClient *ents1.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of gateway ingestion.
type IngestResult struct {
	GatewayCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all ranger gateways from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	totalExpected, err := s.client.GetCount()
	if err != nil {
		slog.Warn("s1 ranger gateways: failed to get count, continuing without total", "error", err)
	}

	var allGateways []*RangerGatewayData
	cursor := ""
	batchNum := 0

	for {
		batchNum++
		batch, err := s.client.GetGatewaysBatch(cursor)
		if err != nil {
			slog.Error("s1 ranger gateways batch failed", "batch", batchNum, "totalSoFar", len(allGateways), "error", err)
			return nil, fmt.Errorf("get ranger gateways batch: %w", err)
		}

		for _, apiGateway := range batch.Gateways {
			allGateways = append(allGateways, ConvertRangerGateway(apiGateway, collectedAt))
		}

		slog.Info("s1 ranger gateways batch fetched", "batch", batchNum, "batchItems", len(batch.Gateways), "totalFetched", len(allGateways), "totalExpected", totalExpected, "hasMore", batch.HasMore)

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveGateways(ctx, allGateways); err != nil {
		return nil, fmt.Errorf("save ranger gateways: %w", err)
	}

	return &IngestResult{
		GatewayCount:   len(allGateways),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveGateways(ctx context.Context, gateways []*RangerGatewayData) error {
	if len(gateways) == 0 {
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

	activeIDs := make(map[string]struct{}, len(gateways))

	for _, data := range gateways {
		existing, err := tx.BronzeS1RangerGateway.Query().
			Where(bronzes1rangergateway.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ents1.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing ranger gateway %s: %w", data.ResourceID, err)
		}

		diff := DiffRangerGatewayData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1RangerGateway.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for ranger gateway %s: %w", data.ResourceID, err)
			}
			activeIDs[data.ResourceID] = struct{}{}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1RangerGateway.Create().
				SetID(data.ResourceID).
				SetIP(data.IP).
				SetMACAddress(data.MacAddress).
				SetExternalIP(data.ExternalIP).
				SetManufacturer(data.Manufacturer).
				SetNetworkName(data.NetworkName).
				SetAccountID(data.AccountID).
				SetAccountName(data.AccountName).
				SetSiteID(data.SiteID).
				SetNumberOfAgents(data.NumberOfAgents).
				SetNumberOfRangers(data.NumberOfRangers).
				SetConnectedRangers(data.ConnectedRangers).
				SetTotalAgents(data.TotalAgents).
				SetAgentPercentage(data.AgentPercentage).
				SetAllowScan(data.AllowScan).
				SetArchived(data.Archived).
				SetNewNetwork(data.NewNetwork).
				SetInheritSettings(data.InheritSettings).
				SetTCPPortScan(data.TCPPortScan).
				SetUDPPortScan(data.UDPPortScan).
				SetIcmpScan(data.ICMPScan).
				SetSmbScan(data.SMBScan).
				SetMdnsScan(data.MDNSScan).
				SetRdnsScan(data.RDNSScan).
				SetSnmpScan(data.SNMPScan).
				SetScanOnlyLocalSubnets(data.ScanOnlyLocalSubnets).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.CreatedAtAPI != nil {
				create.SetCreatedAtAPI(*data.CreatedAtAPI)
			}
			if data.ExpiryDate != nil {
				create.SetExpiryDate(*data.ExpiryDate)
			}
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
				return fmt.Errorf("create ranger gateway %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for ranger gateway %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1RangerGateway.UpdateOneID(data.ResourceID).
				SetIP(data.IP).
				SetMACAddress(data.MacAddress).
				SetExternalIP(data.ExternalIP).
				SetManufacturer(data.Manufacturer).
				SetNetworkName(data.NetworkName).
				SetAccountID(data.AccountID).
				SetAccountName(data.AccountName).
				SetSiteID(data.SiteID).
				SetNumberOfAgents(data.NumberOfAgents).
				SetNumberOfRangers(data.NumberOfRangers).
				SetConnectedRangers(data.ConnectedRangers).
				SetTotalAgents(data.TotalAgents).
				SetAgentPercentage(data.AgentPercentage).
				SetAllowScan(data.AllowScan).
				SetArchived(data.Archived).
				SetNewNetwork(data.NewNetwork).
				SetInheritSettings(data.InheritSettings).
				SetTCPPortScan(data.TCPPortScan).
				SetUDPPortScan(data.UDPPortScan).
				SetIcmpScan(data.ICMPScan).
				SetSmbScan(data.SMBScan).
				SetMdnsScan(data.MDNSScan).
				SetRdnsScan(data.RDNSScan).
				SetSnmpScan(data.SNMPScan).
				SetScanOnlyLocalSubnets(data.ScanOnlyLocalSubnets).
				SetCollectedAt(data.CollectedAt)

			if data.CreatedAtAPI != nil {
				update.SetCreatedAtAPI(*data.CreatedAtAPI)
			}
			if data.ExpiryDate != nil {
				update.SetExpiryDate(*data.ExpiryDate)
			}
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
				return fmt.Errorf("update ranger gateway %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for ranger gateway %s: %w", data.ResourceID, err)
			}
		}

		activeIDs[data.ResourceID] = struct{}{}
	}

	// Delete stale ranger gateways not returned by the API.
	allDBIDs, err := tx.BronzeS1RangerGateway.Query().
		Select(bronzes1rangergateway.FieldID).
		Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query all ranger gateway IDs: %w", err)
	}

	staleCount := 0
	for _, id := range allDBIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}

		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale ranger gateway %s: %w", id, err)
		}

		if err := tx.BronzeS1RangerGateway.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale ranger gateway %s: %w", id, err)
		}
		staleCount++
	}

	if staleCount > 0 {
		slog.Info("s1 ranger gateways: deleted stale", "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

