package packetmirroring

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputepacketmirroring"
)

// Service handles GCP Compute packet mirroring ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new packet mirroring ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for packet mirroring ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of packet mirroring ingestion.
type IngestResult struct {
	ProjectID            string
	PacketMirroringCount int
	CollectedAt          time.Time
	DurationMillis       int64
}

// Ingest fetches packet mirrorings from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch packet mirrorings from GCP
	packetMirrorings, err := s.client.ListPacketMirrorings(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list packet mirrorings: %w", err)
	}

	// Convert to data structs
	pmDataList := make([]*PacketMirroringData, 0, len(packetMirrorings))
	for _, pm := range packetMirrorings {
		data, err := ConvertPacketMirroring(pm, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert packet mirroring: %w", err)
		}
		pmDataList = append(pmDataList, data)
	}

	// Save to database
	if err := s.savePacketMirrorings(ctx, pmDataList); err != nil {
		return nil, fmt.Errorf("failed to save packet mirrorings: %w", err)
	}

	return &IngestResult{
		ProjectID:            params.ProjectID,
		PacketMirroringCount: len(pmDataList),
		CollectedAt:          collectedAt,
		DurationMillis:       time.Since(startTime).Milliseconds(),
	}, nil
}

// savePacketMirrorings saves packet mirrorings to the database with history tracking.
func (s *Service) savePacketMirrorings(ctx context.Context, packetMirrorings []*PacketMirroringData) error {
	if len(packetMirrorings) == 0 {
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

	for _, pmData := range packetMirrorings {
		// Load existing packet mirroring
		existing, err := tx.BronzeGCPComputePacketMirroring.Query().
			Where(bronzegcpcomputepacketmirroring.ID(pmData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing packet mirroring %s: %w", pmData.Name, err)
		}

		// Compute diff
		diff := DiffPacketMirroringData(existing, pmData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputePacketMirroring.UpdateOneID(pmData.ID).
				SetCollectedAt(pmData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for packet mirroring %s: %w", pmData.Name, err)
			}
			continue
		}

		// Create or update packet mirroring
		if existing == nil {
			// Create new packet mirroring
			create := tx.BronzeGCPComputePacketMirroring.Create().
				SetID(pmData.ID).
				SetName(pmData.Name).
				SetProjectID(pmData.ProjectID).
				SetCollectedAt(pmData.CollectedAt).
				SetFirstCollectedAt(pmData.CollectedAt)

			if pmData.Description != "" {
				create.SetDescription(pmData.Description)
			}
			if pmData.SelfLink != "" {
				create.SetSelfLink(pmData.SelfLink)
			}
			if pmData.Region != "" {
				create.SetRegion(pmData.Region)
			}
			if pmData.Network != "" {
				create.SetNetwork(pmData.Network)
			}
			if pmData.Priority != 0 {
				create.SetPriority(pmData.Priority)
			}
			if pmData.Enable != "" {
				create.SetEnable(pmData.Enable)
			}
			if pmData.CollectorIlbJSON != nil {
				create.SetCollectorIlbJSON(pmData.CollectorIlbJSON)
			}
			if pmData.MirroredResourcesJSON != nil {
				create.SetMirroredResourcesJSON(pmData.MirroredResourcesJSON)
			}
			if pmData.FilterJSON != nil {
				create.SetFilterJSON(pmData.FilterJSON)
			}
			if pmData.CreationTimestamp != "" {
				create.SetCreationTimestamp(pmData.CreationTimestamp)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create packet mirroring %s: %w", pmData.Name, err)
			}
		} else {
			// Update existing packet mirroring
			update := tx.BronzeGCPComputePacketMirroring.UpdateOneID(pmData.ID).
				SetName(pmData.Name).
				SetProjectID(pmData.ProjectID).
				SetCollectedAt(pmData.CollectedAt)

			if pmData.Description != "" {
				update.SetDescription(pmData.Description)
			}
			if pmData.SelfLink != "" {
				update.SetSelfLink(pmData.SelfLink)
			}
			if pmData.Region != "" {
				update.SetRegion(pmData.Region)
			}
			if pmData.Network != "" {
				update.SetNetwork(pmData.Network)
			}
			if pmData.Priority != 0 {
				update.SetPriority(pmData.Priority)
			}
			if pmData.Enable != "" {
				update.SetEnable(pmData.Enable)
			}
			if pmData.CollectorIlbJSON != nil {
				update.SetCollectorIlbJSON(pmData.CollectorIlbJSON)
			}
			if pmData.MirroredResourcesJSON != nil {
				update.SetMirroredResourcesJSON(pmData.MirroredResourcesJSON)
			}
			if pmData.FilterJSON != nil {
				update.SetFilterJSON(pmData.FilterJSON)
			}
			if pmData.CreationTimestamp != "" {
				update.SetCreationTimestamp(pmData.CreationTimestamp)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update packet mirroring %s: %w", pmData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, pmData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for packet mirroring %s: %w", pmData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, pmData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for packet mirroring %s: %w", pmData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePacketMirrorings removes packet mirrorings that were not collected in the latest run.
// Also closes history records for deleted packet mirrorings.
func (s *Service) DeleteStalePacketMirrorings(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale packet mirrorings
	stalePacketMirrorings, err := tx.BronzeGCPComputePacketMirroring.Query().
		Where(
			bronzegcpcomputepacketmirroring.ProjectID(projectID),
			bronzegcpcomputepacketmirroring.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale packet mirroring
	for _, pm := range stalePacketMirrorings {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, pm.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for packet mirroring %s: %w", pm.ID, err)
		}

		// Delete packet mirroring
		if err := tx.BronzeGCPComputePacketMirroring.DeleteOne(pm).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete packet mirroring %s: %w", pm.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
