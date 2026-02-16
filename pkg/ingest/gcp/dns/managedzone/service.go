package managedzone

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpdnsmanagedzone"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpdnsmanagedzonelabel"
)

// Service handles GCP DNS managed zone ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new managed zone ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for managed zone ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of managed zone ingestion.
type IngestResult struct {
	ProjectID        string
	ManagedZoneCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches managed zones from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch managed zones from GCP
	zones, err := s.client.ListManagedZones(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list managed zones: %w", err)
	}

	// Convert to data structs
	zoneDataList := make([]*ManagedZoneData, 0, len(zones))
	for _, zone := range zones {
		data, err := ConvertManagedZone(zone, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert managed zone: %w", err)
		}
		zoneDataList = append(zoneDataList, data)
	}

	// Save to database
	if err := s.saveManagedZones(ctx, zoneDataList); err != nil {
		return nil, fmt.Errorf("failed to save managed zones: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		ManagedZoneCount: len(zoneDataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveManagedZones saves managed zones to the database with history tracking.
func (s *Service) saveManagedZones(ctx context.Context, zones []*ManagedZoneData) error {
	if len(zones) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, zoneData := range zones {
		// Load existing managed zone with labels
		existing, err := tx.BronzeGCPDNSManagedZone.Query().
			Where(bronzegcpdnsmanagedzone.ID(zoneData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing managed zone %s: %w", zoneData.Name, err)
		}

		// Compute diff
		diff := DiffManagedZoneData(existing, zoneData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPDNSManagedZone.UpdateOneID(zoneData.ID).
				SetCollectedAt(zoneData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for managed zone %s: %w", zoneData.Name, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := deleteManagedZoneChildren(ctx, tx, zoneData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for managed zone %s: %w", zoneData.Name, err)
			}
		}

		// Create or update managed zone
		var savedZone *ent.BronzeGCPDNSManagedZone
		if existing == nil {
			// Create new managed zone
			create := tx.BronzeGCPDNSManagedZone.Create().
				SetID(zoneData.ID).
				SetName(zoneData.Name).
				SetProjectID(zoneData.ProjectID).
				SetCollectedAt(zoneData.CollectedAt).
				SetFirstCollectedAt(zoneData.CollectedAt)

			if zoneData.DnsName != "" {
				create.SetDNSName(zoneData.DnsName)
			}
			if zoneData.Description != "" {
				create.SetDescription(zoneData.Description)
			}
			if zoneData.Visibility != "" {
				create.SetVisibility(zoneData.Visibility)
			}
			if zoneData.CreationTime != "" {
				create.SetCreationTime(zoneData.CreationTime)
			}
			if zoneData.DnssecConfigJSON != nil {
				create.SetDnssecConfigJSON(zoneData.DnssecConfigJSON)
			}
			if zoneData.PrivateVisibilityConfigJSON != nil {
				create.SetPrivateVisibilityConfigJSON(zoneData.PrivateVisibilityConfigJSON)
			}
			if zoneData.ForwardingConfigJSON != nil {
				create.SetForwardingConfigJSON(zoneData.ForwardingConfigJSON)
			}
			if zoneData.PeeringConfigJSON != nil {
				create.SetPeeringConfigJSON(zoneData.PeeringConfigJSON)
			}
			if zoneData.CloudLoggingConfigJSON != nil {
				create.SetCloudLoggingConfigJSON(zoneData.CloudLoggingConfigJSON)
			}

			savedZone, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create managed zone %s: %w", zoneData.Name, err)
			}
		} else {
			// Update existing managed zone
			update := tx.BronzeGCPDNSManagedZone.UpdateOneID(zoneData.ID).
				SetName(zoneData.Name).
				SetProjectID(zoneData.ProjectID).
				SetCollectedAt(zoneData.CollectedAt)

			if zoneData.DnsName != "" {
				update.SetDNSName(zoneData.DnsName)
			}
			if zoneData.Description != "" {
				update.SetDescription(zoneData.Description)
			}
			if zoneData.Visibility != "" {
				update.SetVisibility(zoneData.Visibility)
			}
			if zoneData.CreationTime != "" {
				update.SetCreationTime(zoneData.CreationTime)
			}
			if zoneData.DnssecConfigJSON != nil {
				update.SetDnssecConfigJSON(zoneData.DnssecConfigJSON)
			}
			if zoneData.PrivateVisibilityConfigJSON != nil {
				update.SetPrivateVisibilityConfigJSON(zoneData.PrivateVisibilityConfigJSON)
			}
			if zoneData.ForwardingConfigJSON != nil {
				update.SetForwardingConfigJSON(zoneData.ForwardingConfigJSON)
			}
			if zoneData.PeeringConfigJSON != nil {
				update.SetPeeringConfigJSON(zoneData.PeeringConfigJSON)
			}
			if zoneData.CloudLoggingConfigJSON != nil {
				update.SetCloudLoggingConfigJSON(zoneData.CloudLoggingConfigJSON)
			}

			savedZone, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update managed zone %s: %w", zoneData.Name, err)
			}
		}

		// Create new children
		if err := createManagedZoneChildren(ctx, tx, savedZone, zoneData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for managed zone %s: %w", zoneData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, zoneData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for managed zone %s: %w", zoneData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, zoneData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for managed zone %s: %w", zoneData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteManagedZoneChildren deletes labels for a managed zone.
func deleteManagedZoneChildren(ctx context.Context, tx *ent.Tx, managedZoneID string) error {
	_, err := tx.BronzeGCPDNSManagedZoneLabel.Delete().
		Where(bronzegcpdnsmanagedzonelabel.HasManagedZoneWith(bronzegcpdnsmanagedzone.ID(managedZoneID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	return nil
}

// createManagedZoneChildren creates labels for a managed zone.
func createManagedZoneChildren(ctx context.Context, tx *ent.Tx, savedZone *ent.BronzeGCPDNSManagedZone, zoneData *ManagedZoneData) error {
	for _, label := range zoneData.Labels {
		_, err := tx.BronzeGCPDNSManagedZoneLabel.Create().
			SetKey(label.Key).
			SetValue(label.Value).
			SetManagedZone(savedZone).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}

	return nil
}

// DeleteStaleManagedZones removes managed zones that were not collected in the latest run.
// Also closes history records for deleted managed zones.
func (s *Service) DeleteStaleManagedZones(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale managed zones
	staleZones, err := tx.BronzeGCPDNSManagedZone.Query().
		Where(
			bronzegcpdnsmanagedzone.ProjectID(projectID),
			bronzegcpdnsmanagedzone.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale managed zone
	for _, zone := range staleZones {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, zone.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for managed zone %s: %w", zone.ID, err)
		}

		// Delete managed zone (labels will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPDNSManagedZone.DeleteOne(zone).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete managed zone %s: %w", zone.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
