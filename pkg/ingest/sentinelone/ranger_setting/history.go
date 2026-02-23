package ranger_setting

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorys1rangersetting"
)

// HistoryService handles history tracking for ranger settings.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new ranger setting.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *RangerSettingData, now time.Time) error {
	create := tx.BronzeHistoryS1RangerSetting.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
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
		SetNewNetworkInHours(data.NewNetworkInHours)

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
		return fmt.Errorf("create ranger setting history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed ranger setting.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1RangerSetting, new *RangerSettingData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1RangerSetting.Query().
		Where(
			bronzehistorys1rangersetting.ResourceID(old.ID),
			bronzehistorys1rangersetting.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current ranger setting history: %w", err)
	}

	if err := tx.BronzeHistoryS1RangerSetting.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ranger setting history: %w", err)
	}

	create := tx.BronzeHistoryS1RangerSetting.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetAccountID(new.AccountID).
		SetScopeID(new.ScopeID).
		SetEnabled(new.Enabled).
		SetUsePeriodicSnapshots(new.UsePeriodicSnapshots).
		SetSnapshotPeriod(new.SnapshotPeriod).
		SetNetworkDecommissionValue(new.NetworkDecommissionValue).
		SetMinAgentsInNetworkToScan(new.MinAgentsInNetworkToScan).
		SetTCPPortScan(new.TCPPortScan).
		SetUDPPortScan(new.UDPPortScan).
		SetIcmpScan(new.ICMPScan).
		SetSmbScan(new.SMBScan).
		SetMdnsScan(new.MDNSScan).
		SetRdnsScan(new.RDNSScan).
		SetSnmpScan(new.SNMPScan).
		SetMultiScanSsdp(new.MultiScanSSDP).
		SetUseFullDNSScan(new.UseFullDNSScan).
		SetScanOnlyLocalSubnets(new.ScanOnlyLocalSubnets).
		SetAutoEnableNetworks(new.AutoEnableNetworks).
		SetCombineDevices(new.CombineDevices).
		SetNewNetworkInHours(new.NewNetworkInHours)

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
		return fmt.Errorf("create new ranger setting history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted ranger setting.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1RangerSetting.Query().
		Where(
			bronzehistorys1rangersetting.ResourceID(resourceID),
			bronzehistorys1rangersetting.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current ranger setting history: %w", err)
	}

	if err := tx.BronzeHistoryS1RangerSetting.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ranger setting history: %w", err)
	}

	return nil
}
