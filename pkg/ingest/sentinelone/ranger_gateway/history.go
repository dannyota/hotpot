package ranger_gateway

import (
	"context"
	"fmt"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzehistorys1rangergateway"
)

// HistoryService handles history tracking for ranger gateways.
type HistoryService struct {
	entClient *ents1.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ents1.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new gateway.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ents1.Tx, data *RangerGatewayData, now time.Time) error {
	create := tx.BronzeHistoryS1RangerGateway.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
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
		SetScanOnlyLocalSubnets(data.ScanOnlyLocalSubnets)

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
		return fmt.Errorf("create ranger gateway history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed gateway.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ents1.Tx, old *ents1.BronzeS1RangerGateway, new *RangerGatewayData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1RangerGateway.Query().
		Where(
			bronzehistorys1rangergateway.ResourceID(old.ID),
			bronzehistorys1rangergateway.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current ranger gateway history: %w", err)
	}

	if err := tx.BronzeHistoryS1RangerGateway.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ranger gateway history: %w", err)
	}

	create := tx.BronzeHistoryS1RangerGateway.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetIP(new.IP).
		SetMACAddress(new.MacAddress).
		SetExternalIP(new.ExternalIP).
		SetManufacturer(new.Manufacturer).
		SetNetworkName(new.NetworkName).
		SetAccountID(new.AccountID).
		SetAccountName(new.AccountName).
		SetSiteID(new.SiteID).
		SetNumberOfAgents(new.NumberOfAgents).
		SetNumberOfRangers(new.NumberOfRangers).
		SetConnectedRangers(new.ConnectedRangers).
		SetTotalAgents(new.TotalAgents).
		SetAgentPercentage(new.AgentPercentage).
		SetAllowScan(new.AllowScan).
		SetArchived(new.Archived).
		SetNewNetwork(new.NewNetwork).
		SetInheritSettings(new.InheritSettings).
		SetTCPPortScan(new.TCPPortScan).
		SetUDPPortScan(new.UDPPortScan).
		SetIcmpScan(new.ICMPScan).
		SetSmbScan(new.SMBScan).
		SetMdnsScan(new.MDNSScan).
		SetRdnsScan(new.RDNSScan).
		SetSnmpScan(new.SNMPScan).
		SetScanOnlyLocalSubnets(new.ScanOnlyLocalSubnets)

	if new.CreatedAtAPI != nil {
		create.SetCreatedAtAPI(*new.CreatedAtAPI)
	}
	if new.ExpiryDate != nil {
		create.SetExpiryDate(*new.ExpiryDate)
	}
	if new.TCPPortsJSON != nil {
		create.SetTCPPortsJSON(new.TCPPortsJSON)
	}
	if new.UDPPortsJSON != nil {
		create.SetUDPPortsJSON(new.UDPPortsJSON)
	}
	if new.RestrictionsJSON != nil {
		create.SetRestrictionsJSON(new.RestrictionsJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new ranger gateway history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted gateway.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ents1.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1RangerGateway.Query().
		Where(
			bronzehistorys1rangergateway.ResourceID(resourceID),
			bronzehistorys1rangergateway.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ents1.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current ranger gateway history: %w", err)
	}

	if err := tx.BronzeHistoryS1RangerGateway.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ranger gateway history: %w", err)
	}

	return nil
}
