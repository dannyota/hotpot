package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodomain"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodomainrecord"
)

// Service handles DigitalOcean Domain ingestion.
type Service struct {
	client        *Client
	entClient     *ent.Client
	domainHistory *DomainHistoryService
	recordHistory *RecordHistoryService
}

// NewService creates a new Domain ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:        client,
		entClient:     entClient,
		domainHistory: NewDomainHistoryService(entClient),
		recordHistory: NewRecordHistoryService(entClient),
	}
}

// IngestDomainsResult contains the result of Domain ingestion.
type IngestDomainsResult struct {
	DomainCount    int
	CollectedAt    time.Time
	DurationMillis int64
	DomainNames    []string
}

// IngestDomains fetches all Domains from DigitalOcean and saves them.
func (s *Service) IngestDomains(ctx context.Context, heartbeat func()) (*IngestDomainsResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiDomains, err := s.client.ListAllDomains(ctx)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allDomains []*DomainData
	var domainNames []string
	for _, v := range apiDomains {
		allDomains = append(allDomains, ConvertDomain(v, collectedAt))
		domainNames = append(domainNames, v.Name)
	}

	if err := s.saveDomains(ctx, allDomains); err != nil {
		return nil, fmt.Errorf("save domains: %w", err)
	}

	return &IngestDomainsResult{
		DomainCount:    len(allDomains),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
		DomainNames:    domainNames,
	}, nil
}

func (s *Service) saveDomains(ctx context.Context, domains []*DomainData) error {
	if len(domains) == 0 {
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

	for _, data := range domains {
		existing, err := tx.BronzeDODomain.Query().
			Where(bronzedodomain.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Domain %s: %w", data.ResourceID, err)
		}

		diff := DiffDomainData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODomain.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Domain %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODomain.Create().
				SetID(data.ResourceID).
				SetTTL(data.TTL).
				SetZoneFile(data.ZoneFile).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Domain %s: %w", data.ResourceID, err)
			}

			if err := s.domainHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Domain %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODomain.UpdateOneID(data.ResourceID).
				SetTTL(data.TTL).
				SetZoneFile(data.ZoneFile).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Domain %s: %w", data.ResourceID, err)
			}

			if err := s.domainHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Domain %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleDomains removes Domains that were not collected in the latest run.
func (s *Service) DeleteStaleDomains(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeDODomain.Query().
		Where(bronzedodomain.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, d := range stale {
		if err := s.domainHistory.CloseHistory(ctx, tx, d.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for Domain %s: %w", d.ID, err)
		}

		if err := tx.BronzeDODomain.DeleteOne(d).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete Domain %s: %w", d.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// IngestRecordsResult contains the result of Domain Record ingestion.
type IngestRecordsResult struct {
	RecordCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// IngestRecords fetches all Domain Records for given domains and saves them.
func (s *Service) IngestRecords(ctx context.Context, domainNames []string, heartbeat func()) (*IngestRecordsResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allRecords []*DomainRecordData
	for _, domainName := range domainNames {
		apiRecords, err := s.client.ListAllRecords(ctx, domainName)
		if err != nil {
			return nil, fmt.Errorf("list records for domain %s: %w", domainName, err)
		}

		for _, v := range apiRecords {
			allRecords = append(allRecords, ConvertDomainRecord(v, domainName, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}
	}

	if err := s.saveRecords(ctx, allRecords); err != nil {
		return nil, fmt.Errorf("save domain records: %w", err)
	}

	return &IngestRecordsResult{
		RecordCount:    len(allRecords),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveRecords(ctx context.Context, records []*DomainRecordData) error {
	if len(records) == 0 {
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

	for _, data := range records {
		existing, err := tx.BronzeDODomainRecord.Query().
			Where(bronzedodomainrecord.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing DomainRecord %s: %w", data.ResourceID, err)
		}

		diff := DiffDomainRecordData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODomainRecord.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for DomainRecord %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODomainRecord.Create().
				SetID(data.ResourceID).
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
				SetTag(data.Tag).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create DomainRecord %s: %w", data.ResourceID, err)
			}

			if err := s.recordHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for DomainRecord %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODomainRecord.UpdateOneID(data.ResourceID).
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
				SetTag(data.Tag).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update DomainRecord %s: %w", data.ResourceID, err)
			}

			if err := s.recordHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for DomainRecord %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleRecords removes Domain Records that were not collected in the latest run.
func (s *Service) DeleteStaleRecords(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeDODomainRecord.Query().
		Where(bronzedodomainrecord.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, r := range stale {
		if err := s.recordHistory.CloseHistory(ctx, tx, r.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for DomainRecord %s: %w", r.ID, err)
		}

		if err := tx.BronzeDODomainRecord.DeleteOne(r).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete DomainRecord %s: %w", r.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
