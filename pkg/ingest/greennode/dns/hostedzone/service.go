package hostedzone

import (
	"context"
	"fmt"
	"time"

	entdns "github.com/dannyota/hotpot/pkg/storage/ent/greennode/dns"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/dns/bronzegreennodednshostedzone"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/dns/bronzegreennodednsrecord"
)

// Service handles GreenNode DNS hosted zone ingestion.
type Service struct {
	client    *Client
	entClient *entdns.Client
	history   *HistoryService
}

// NewService creates a new hosted zone ingestion service.
func NewService(client *Client, entClient *entdns.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of hosted zone ingestion.
type IngestResult struct {
	HostedZoneCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches hosted zones and records from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	hostedZones, err := s.client.ListHostedZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list hosted zones: %w", err)
	}

	zoneDataList := make([]*HostedZoneData, 0, len(hostedZones))
	for _, hz := range hostedZones {
		data, err := ConvertHostedZone(hz, projectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert hosted zone: %w", err)
		}

		// Fetch records for this hosted zone
		records, err := s.client.ListRecordsByHostedZoneID(ctx, hz.HostedZoneID)
		if err != nil {
			return nil, fmt.Errorf("list records for zone %s: %w", hz.HostedZoneID, err)
		}

		recordData, err := ConvertRecords(records)
		if err != nil {
			return nil, fmt.Errorf("convert records for zone %s: %w", hz.HostedZoneID, err)
		}
		data.Records = recordData

		zoneDataList = append(zoneDataList, data)
	}

	if err := s.saveHostedZones(ctx, zoneDataList); err != nil {
		return nil, fmt.Errorf("save hosted zones: %w", err)
	}

	return &IngestResult{
		HostedZoneCount: len(zoneDataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveHostedZones(ctx context.Context, zones []*HostedZoneData) error {
	if len(zones) == 0 {
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

	for _, data := range zones {
		existing, err := tx.BronzeGreenNodeDNSHostedZone.Query().
			Where(bronzegreennodednshostedzone.ID(data.ID)).
			WithRecords().
			First(ctx)
		if err != nil && !entdns.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing hosted zone %s: %w", data.ID, err)
		}

		diff := DiffHostedZoneData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeDNSHostedZone.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for hosted zone %s: %w", data.ID, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteHostedZoneChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for hosted zone %s: %w", data.ID, err)
			}
		}

		var savedZone *entdns.BronzeGreenNodeDNSHostedZone
		if existing == nil {
			create := tx.BronzeGreenNodeDNSHostedZone.Create().
				SetID(data.ID).
				SetDomainName(data.DomainName).
				SetStatus(data.Status).
				SetDescription(data.Description).
				SetType(data.Type).
				SetCountRecords(data.CountRecords).
				SetPortalUserID(data.PortalUserID).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetNillableDeletedAtAPI(data.DeletedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.AssocVpcIdsJSON != nil {
				create.SetAssocVpcIdsJSON(data.AssocVpcIdsJSON)
			}
			if data.AssocVpcMapRegionJSON != nil {
				create.SetAssocVpcMapRegionJSON(data.AssocVpcMapRegionJSON)
			}

			savedZone, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create hosted zone %s: %w", data.ID, err)
			}
		} else {
			update := tx.BronzeGreenNodeDNSHostedZone.UpdateOneID(data.ID).
				SetDomainName(data.DomainName).
				SetStatus(data.Status).
				SetDescription(data.Description).
				SetType(data.Type).
				SetCountRecords(data.CountRecords).
				SetPortalUserID(data.PortalUserID).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetNillableDeletedAtAPI(data.DeletedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.AssocVpcIdsJSON != nil {
				update.SetAssocVpcIdsJSON(data.AssocVpcIdsJSON)
			}
			if data.AssocVpcMapRegionJSON != nil {
				update.SetAssocVpcMapRegionJSON(data.AssocVpcMapRegionJSON)
			}

			savedZone, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update hosted zone %s: %w", data.ID, err)
			}
		}

		if err := s.createHostedZoneChildren(ctx, tx, savedZone, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for hosted zone %s: %w", data.ID, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for hosted zone %s: %w", data.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for hosted zone %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteHostedZoneChildren(ctx context.Context, tx *entdns.Tx, hostedZoneID string) error {
	_, err := tx.BronzeGreenNodeDNSRecord.Delete().
		Where(bronzegreennodednsrecord.HasHostedZoneWith(bronzegreennodednshostedzone.ID(hostedZoneID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete records: %w", err)
	}
	return nil
}

func (s *Service) createHostedZoneChildren(ctx context.Context, tx *entdns.Tx, zone *entdns.BronzeGreenNodeDNSHostedZone, data *HostedZoneData) error {
	for _, r := range data.Records {
		create := tx.BronzeGreenNodeDNSRecord.Create().
			SetHostedZone(zone).
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
			return fmt.Errorf("create record %s: %w", r.RecordID, err)
		}
	}
	return nil
}

// DeleteStaleHostedZones removes hosted zones not collected in the latest run for the given project.
func (s *Service) DeleteStaleHostedZones(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeDNSHostedZone.Query().
		Where(
			bronzegreennodednshostedzone.ProjectID(projectID),
			bronzegreennodednshostedzone.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale hosted zones: %w", err)
	}

	for _, hz := range stale {
		if err := s.history.CloseHistory(ctx, tx, hz.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for hosted zone %s: %w", hz.ID, err)
		}
		if err := s.deleteHostedZoneChildren(ctx, tx, hz.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for hosted zone %s: %w", hz.ID, err)
		}
		if err := tx.BronzeGreenNodeDNSHostedZone.DeleteOneID(hz.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete hosted zone %s: %w", hz.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
