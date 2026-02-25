package ranger_device

import (
	"context"
	"fmt"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzehistorys1rangerdevice"
)

// HistoryService handles history tracking for ranger devices.
type HistoryService struct {
	entClient *ents1.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ents1.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new ranger device.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ents1.Tx, data *RangerDeviceData, now time.Time) error {
	create := tx.BronzeHistoryS1RangerDevice.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
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
		SetFingerprintScore(data.FingerprintScore)

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
		return fmt.Errorf("create ranger device history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed ranger device.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ents1.Tx, old *ents1.BronzeS1RangerDevice, new *RangerDeviceData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1RangerDevice.Query().
		Where(
			bronzehistorys1rangerdevice.ResourceID(old.ID),
			bronzehistorys1rangerdevice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current ranger device history: %w", err)
	}

	if err := tx.BronzeHistoryS1RangerDevice.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ranger device history: %w", err)
	}

	create := tx.BronzeHistoryS1RangerDevice.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetLocalIP(new.LocalIP).
		SetExternalIP(new.ExternalIP).
		SetMACAddress(new.MACAddress).
		SetOsType(new.OsType).
		SetOsName(new.OsName).
		SetOsVersion(new.OsVersion).
		SetDeviceType(new.DeviceType).
		SetDeviceFunction(new.DeviceFunction).
		SetManufacturer(new.Manufacturer).
		SetManagedState(new.ManagedState).
		SetAgentID(new.AgentID).
		SetSubnetAddress(new.SubnetAddress).
		SetGatewayIPAddress(new.GatewayIPAddress).
		SetGatewayMACAddress(new.GatewayMACAddress).
		SetNetworkName(new.NetworkName).
		SetDomain(new.Domain).
		SetSiteName(new.SiteName).
		SetDeviceReview(new.DeviceReview).
		SetHasIdentity(new.HasIdentity).
		SetHasUserLabel(new.HasUserLabel).
		SetFingerprintScore(new.FingerprintScore)

	if new.FirstSeen != nil {
		create.SetFirstSeen(*new.FirstSeen)
	}
	if new.LastSeen != nil {
		create.SetLastSeen(*new.LastSeen)
	}
	if new.TCPPortsJSON != nil {
		create.SetTCPPortsJSON(new.TCPPortsJSON)
	}
	if new.UDPPortsJSON != nil {
		create.SetUDPPortsJSON(new.UDPPortsJSON)
	}
	if new.HostnamesJSON != nil {
		create.SetHostnamesJSON(new.HostnamesJSON)
	}
	if new.DiscoveryMethodsJSON != nil {
		create.SetDiscoveryMethodsJSON(new.DiscoveryMethodsJSON)
	}
	if new.NetworksJSON != nil {
		create.SetNetworksJSON(new.NetworksJSON)
	}
	if new.TagsJSON != nil {
		create.SetTagsJSON(new.TagsJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new ranger device history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted ranger device.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ents1.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1RangerDevice.Query().
		Where(
			bronzehistorys1rangerdevice.ResourceID(resourceID),
			bronzehistorys1rangerdevice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ents1.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current ranger device history: %w", err)
	}

	if err := tx.BronzeHistoryS1RangerDevice.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ranger device history: %w", err)
	}

	return nil
}
