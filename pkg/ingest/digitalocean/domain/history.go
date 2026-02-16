package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodomain"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodomainrecord"
)

// DomainHistoryService handles history tracking for Domains.
type DomainHistoryService struct {
	entClient *ent.Client
}

// NewDomainHistoryService creates a new domain history service.
func NewDomainHistoryService(entClient *ent.Client) *DomainHistoryService {
	return &DomainHistoryService{entClient: entClient}
}

func (h *DomainHistoryService) buildCreate(tx *ent.Tx, data *DomainData) *ent.BronzeHistoryDODomainCreate {
	return tx.BronzeHistoryDODomain.Create().
		SetResourceID(data.ResourceID).
		SetTTL(data.TTL).
		SetZoneFile(data.ZoneFile)
}

// CreateHistory creates a history record for a new Domain.
func (h *DomainHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *DomainData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Domain history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Domain.
func (h *DomainHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODomain, new *DomainData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODomain.Query().
		Where(
			bronzehistorydodomain.ResourceID(old.ID),
			bronzehistorydodomain.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Domain history: %w", err)
	}

	if err := tx.BronzeHistoryDODomain.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Domain history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Domain history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Domain.
func (h *DomainHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODomain.Query().
		Where(
			bronzehistorydodomain.ResourceID(resourceID),
			bronzehistorydodomain.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Domain history: %w", err)
	}

	if err := tx.BronzeHistoryDODomain.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Domain history: %w", err)
	}

	return nil
}

// RecordHistoryService handles history tracking for Domain Records.
type RecordHistoryService struct {
	entClient *ent.Client
}

// NewRecordHistoryService creates a new domain record history service.
func NewRecordHistoryService(entClient *ent.Client) *RecordHistoryService {
	return &RecordHistoryService{entClient: entClient}
}

func (h *RecordHistoryService) buildCreate(tx *ent.Tx, data *DomainRecordData) *ent.BronzeHistoryDODomainRecordCreate {
	return tx.BronzeHistoryDODomainRecord.Create().
		SetResourceID(data.ResourceID).
		SetDomainName(data.DomainName).
		SetRecordID(data.RecordID).
		SetType(data.Type).
		SetName(data.Name).
		SetData(data.Data).
		SetPriority(data.Priority).
		SetPort(data.Port).
		SetTTL(data.TTL).
		SetWeight(data.Weight).
		SetFlags(data.Flags).
		SetTag(data.Tag)
}

// CreateHistory creates a history record for a new Domain Record.
func (h *RecordHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *DomainRecordData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create DomainRecord history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Domain Record.
func (h *RecordHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODomainRecord, new *DomainRecordData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODomainRecord.Query().
		Where(
			bronzehistorydodomainrecord.ResourceID(old.ID),
			bronzehistorydodomainrecord.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current DomainRecord history: %w", err)
	}

	if err := tx.BronzeHistoryDODomainRecord.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close DomainRecord history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new DomainRecord history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Domain Record.
func (h *RecordHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODomainRecord.Query().
		Where(
			bronzehistorydodomainrecord.ResourceID(resourceID),
			bronzehistorydodomainrecord.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current DomainRecord history: %w", err)
	}

	if err := tx.BronzeHistoryDODomainRecord.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close DomainRecord history: %w", err)
	}

	return nil
}
