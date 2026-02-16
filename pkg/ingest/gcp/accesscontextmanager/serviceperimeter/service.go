package serviceperimeter

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpaccesscontextmanagerserviceperimeter"
)

// Service handles service perimeter ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new service perimeter ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of service perimeter ingestion.
type IngestResult struct {
	PerimeterCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches service perimeters from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch service perimeters from GCP
	rawPerimeters, err := s.client.ListServicePerimeters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list service perimeters: %w", err)
	}

	// Convert to perimeter data
	perimeterDataList := make([]*ServicePerimeterData, 0, len(rawPerimeters))
	for _, raw := range rawPerimeters {
		data := ConvertServicePerimeter(raw.OrgName, raw.AccessPolicyName, raw.ServicePerimeter, collectedAt)
		perimeterDataList = append(perimeterDataList, data)
	}

	// Save to database
	if err := s.savePerimeters(ctx, perimeterDataList); err != nil {
		return nil, fmt.Errorf("failed to save service perimeters: %w", err)
	}

	return &IngestResult{
		PerimeterCount: len(perimeterDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePerimeters saves service perimeters to the database with history tracking.
func (s *Service) savePerimeters(ctx context.Context, perimeters []*ServicePerimeterData) error {
	if len(perimeters) == 0 {
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

	for _, perimeterData := range perimeters {
		// Load existing perimeter
		existing, err := tx.BronzeGCPAccessContextManagerServicePerimeter.Query().
			Where(bronzegcpaccesscontextmanagerserviceperimeter.ID(perimeterData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing service perimeter %s: %w", perimeterData.ID, err)
		}

		// Compute diff
		diff := DiffServicePerimeterData(existing, perimeterData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPAccessContextManagerServicePerimeter.UpdateOneID(perimeterData.ID).
				SetCollectedAt(perimeterData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for service perimeter %s: %w", perimeterData.ID, err)
			}
			continue
		}

		// Create or update perimeter
		if existing == nil {
			create := tx.BronzeGCPAccessContextManagerServicePerimeter.Create().
				SetID(perimeterData.ID).
				SetPerimeterType(perimeterData.PerimeterType).
				SetUseExplicitDryRunSpec(perimeterData.UseExplicitDryRunSpec).
				SetAccessPolicyName(perimeterData.AccessPolicyName).
				SetAccessPolicyID(perimeterData.AccessPolicyName).
				SetOrganizationID(perimeterData.OrganizationID).
				SetCollectedAt(perimeterData.CollectedAt).
				SetFirstCollectedAt(perimeterData.CollectedAt)

			if perimeterData.Title != "" {
				create.SetTitle(perimeterData.Title)
			}
			if perimeterData.Description != "" {
				create.SetDescription(perimeterData.Description)
			}
			if perimeterData.Etag != "" {
				create.SetEtag(perimeterData.Etag)
			}
			if perimeterData.StatusJSON != nil {
				create.SetStatusJSON(perimeterData.StatusJSON)
			}
			if perimeterData.SpecJSON != nil {
				create.SetSpecJSON(perimeterData.SpecJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create service perimeter %s: %w", perimeterData.ID, err)
			}
		} else {
			update := tx.BronzeGCPAccessContextManagerServicePerimeter.UpdateOneID(perimeterData.ID).
				SetPerimeterType(perimeterData.PerimeterType).
				SetUseExplicitDryRunSpec(perimeterData.UseExplicitDryRunSpec).
				SetAccessPolicyName(perimeterData.AccessPolicyName).
				SetAccessPolicyID(perimeterData.AccessPolicyName).
				SetOrganizationID(perimeterData.OrganizationID).
				SetCollectedAt(perimeterData.CollectedAt)

			if perimeterData.Title != "" {
				update.SetTitle(perimeterData.Title)
			}
			if perimeterData.Description != "" {
				update.SetDescription(perimeterData.Description)
			}
			if perimeterData.Etag != "" {
				update.SetEtag(perimeterData.Etag)
			}
			if perimeterData.StatusJSON != nil {
				update.SetStatusJSON(perimeterData.StatusJSON)
			}
			if perimeterData.SpecJSON != nil {
				update.SetSpecJSON(perimeterData.SpecJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update service perimeter %s: %w", perimeterData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, perimeterData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for service perimeter %s: %w", perimeterData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, perimeterData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for service perimeter %s: %w", perimeterData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePerimeters removes service perimeters that were not collected in the latest run.
func (s *Service) DeleteStalePerimeters(ctx context.Context, collectedAt time.Time) error {
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

	stalePerimeters, err := tx.BronzeGCPAccessContextManagerServicePerimeter.Query().
		Where(bronzegcpaccesscontextmanagerserviceperimeter.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, perimeter := range stalePerimeters {
		if err := s.history.CloseHistory(ctx, tx, perimeter.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for service perimeter %s: %w", perimeter.ID, err)
		}

		if err := tx.BronzeGCPAccessContextManagerServicePerimeter.DeleteOne(perimeter).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete service perimeter %s: %w", perimeter.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
