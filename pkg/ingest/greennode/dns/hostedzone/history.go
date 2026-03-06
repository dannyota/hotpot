package hostedzone

import (
	"context"
	"fmt"
	"time"

	entdns "danny.vn/hotpot/pkg/storage/ent/greennode/dns"
	"danny.vn/hotpot/pkg/storage/ent/greennode/dns/bronzehistorygreennodednshostedzone"
	"danny.vn/hotpot/pkg/storage/ent/greennode/dns/bronzehistorygreennodednsrecord"
)

// HistoryService handles history tracking for DNS hosted zones.
type HistoryService struct {
	entClient *entdns.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entdns.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new hosted zone and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entdns.Tx, data *HostedZoneData, now time.Time) error {
	zoneHist, err := h.createHostedZoneHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	return h.createRecordsHistory(ctx, tx, zoneHist.ID, data.Records, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entdns.Tx, old *entdns.BronzeGreenNodeDNSHostedZone, new *HostedZoneData, diff *HostedZoneDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeDNSHostedZone.Query().
		Where(
			bronzehistorygreennodednshostedzone.ResourceID(old.ID),
			bronzehistorygreennodednshostedzone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current hosted zone history: %w", err)
	}

	if diff.IsChanged {
		// Close old history
		if err := tx.BronzeHistoryGreenNodeDNSHostedZone.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close hosted zone history: %w", err)
		}

		// Create new history
		zoneHist, err := h.createHostedZoneHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeRecordsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createRecordsHistory(ctx, tx, zoneHist.ID, new.Records, now)
	}

	// Hosted zone unchanged, check children
	if diff.RecordsDiff.Changed {
		if err := h.closeRecordsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createRecordsHistory(ctx, tx, currentHist.ID, new.Records, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted hosted zone.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entdns.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeDNSHostedZone.Query().
		Where(
			bronzehistorygreennodednshostedzone.ResourceID(resourceID),
			bronzehistorygreennodednshostedzone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entdns.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current hosted zone history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeDNSHostedZone.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close hosted zone history: %w", err)
	}

	return h.closeRecordsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createHostedZoneHistory(ctx context.Context, tx *entdns.Tx, data *HostedZoneData, now time.Time, firstCollectedAt time.Time) (*entdns.BronzeHistoryGreenNodeDNSHostedZone, error) {
	create := tx.BronzeHistoryGreenNodeDNSHostedZone.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetDomainName(data.DomainName).
		SetStatus(data.Status).
		SetDescription(data.Description).
		SetType(data.Type).
		SetCountRecords(data.CountRecords).
		SetPortalUserID(data.PortalUserID).
		SetCreatedAtAPI(data.CreatedAtAPI).
		SetNillableDeletedAtAPI(data.DeletedAtAPI).
		SetUpdatedAtAPI(data.UpdatedAtAPI).
		SetProjectID(data.ProjectID)

	if data.AssocVpcIdsJSON != nil {
		create.SetAssocVpcIdsJSON(data.AssocVpcIdsJSON)
	}

	if data.AssocVpcMapRegionJSON != nil {
		create.SetAssocVpcMapRegionJSON(data.AssocVpcMapRegionJSON)
	}

	hist, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create hosted zone history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createRecordsHistory(ctx context.Context, tx *entdns.Tx, zoneHistoryID uint, records []RecordData, now time.Time) error {
	for _, r := range records {
		create := tx.BronzeHistoryGreenNodeDNSRecord.Create().
			SetHostedZoneHistoryID(zoneHistoryID).
			SetValidFrom(now).
			SetRecordID(r.RecordID).
			SetSubDomain(r.SubDomain).
			SetStatus(r.Status).
			SetType(r.Type).
			SetRoutingPolicy(r.RoutingPolicy).
			SetTTL(r.TTL).
			SetNillableEnableStickySession(r.EnableStickySession).
			SetCreatedAtAPI(r.CreatedAtAPI).
			SetNillableDeletedAtAPI(r.DeletedAtAPI).
			SetUpdatedAtAPI(r.UpdatedAtAPI)

		if r.ValueJSON != nil {
			create.SetValueJSON(r.ValueJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("create record history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeRecordsHistory(ctx context.Context, tx *entdns.Tx, zoneHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeDNSRecord.Update().
		Where(
			bronzehistorygreennodednsrecord.HostedZoneHistoryID(zoneHistoryID),
			bronzehistorygreennodednsrecord.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close records history: %w", err)
	}
	return nil
}
