package application

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpappengineapplication"
)

// Service handles App Engine application ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new App Engine application ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for application ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of application ingestion.
type IngestResult struct {
	ProjectID        string
	ApplicationCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches the App Engine application from GCP and stores it in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch application from GCP
	app, err := s.client.GetApplication(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// Application may not exist for this project
	if app == nil {
		return &IngestResult{
			ProjectID:        params.ProjectID,
			ApplicationCount: 0,
			CollectedAt:      collectedAt,
			DurationMillis:   time.Since(startTime).Milliseconds(),
		}, nil
	}

	// Convert to data struct
	appData, err := ConvertApplication(app, params.ProjectID, collectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert application: %w", err)
	}

	// Save to database
	if err := s.saveApplication(ctx, appData); err != nil {
		return nil, fmt.Errorf("failed to save application: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		ApplicationCount: 1,
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveApplication saves an App Engine application to the database with history tracking.
func (s *Service) saveApplication(ctx context.Context, appData *ApplicationData) error {
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

	// Load existing application
	existing, err := tx.BronzeGCPAppEngineApplication.Query().
		Where(bronzegcpappengineapplication.ID(appData.ID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		tx.Rollback()
		return fmt.Errorf("failed to load existing application %s: %w", appData.ID, err)
	}

	// Compute diff
	diff := DiffApplicationData(existing, appData)

	// Skip if no changes
	if !diff.HasAnyChange() && existing != nil {
		if err := tx.BronzeGCPAppEngineApplication.UpdateOneID(appData.ID).
			SetCollectedAt(appData.CollectedAt).
			Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update collected_at for application %s: %w", appData.ID, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Create or update application
	if existing == nil {
		create := tx.BronzeGCPAppEngineApplication.Create().
			SetID(appData.ID).
			SetName(appData.Name).
			SetAuthDomain(appData.AuthDomain).
			SetLocationID(appData.LocationID).
			SetCodeBucket(appData.CodeBucket).
			SetDefaultCookieExpiration(appData.DefaultCookieExpiration).
			SetServingStatus(appData.ServingStatus).
			SetDefaultHostname(appData.DefaultHostname).
			SetDefaultBucket(appData.DefaultBucket).
			SetGcrDomain(appData.GcrDomain).
			SetDatabaseType(appData.DatabaseType).
			SetProjectID(appData.ProjectID).
			SetCollectedAt(appData.CollectedAt).
			SetFirstCollectedAt(appData.CollectedAt)

		if appData.FeatureSettingsJSON != nil {
			create.SetFeatureSettingsJSON(appData.FeatureSettingsJSON)
		}
		if appData.IapJSON != nil {
			create.SetIapJSON(appData.IapJSON)
		}
		if appData.DispatchRulesJSON != nil {
			create.SetDispatchRulesJSON(appData.DispatchRulesJSON)
		}

		_, err = create.Save(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create application %s: %w", appData.ID, err)
		}
	} else {
		update := tx.BronzeGCPAppEngineApplication.UpdateOneID(appData.ID).
			SetName(appData.Name).
			SetAuthDomain(appData.AuthDomain).
			SetLocationID(appData.LocationID).
			SetCodeBucket(appData.CodeBucket).
			SetDefaultCookieExpiration(appData.DefaultCookieExpiration).
			SetServingStatus(appData.ServingStatus).
			SetDefaultHostname(appData.DefaultHostname).
			SetDefaultBucket(appData.DefaultBucket).
			SetGcrDomain(appData.GcrDomain).
			SetDatabaseType(appData.DatabaseType).
			SetProjectID(appData.ProjectID).
			SetCollectedAt(appData.CollectedAt)

		if appData.FeatureSettingsJSON != nil {
			update.SetFeatureSettingsJSON(appData.FeatureSettingsJSON)
		}
		if appData.IapJSON != nil {
			update.SetIapJSON(appData.IapJSON)
		}
		if appData.DispatchRulesJSON != nil {
			update.SetDispatchRulesJSON(appData.DispatchRulesJSON)
		}

		_, err = update.Save(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update application %s: %w", appData.ID, err)
		}
	}

	// Track history
	if diff.IsNew {
		if err := s.history.CreateHistory(ctx, tx, appData, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create history for application %s: %w", appData.ID, err)
		}
	} else {
		if err := s.history.UpdateHistory(ctx, tx, existing, appData, diff, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update history for application %s: %w", appData.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleApplications removes applications that were not collected in the latest run.
func (s *Service) DeleteStaleApplications(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleApps, err := tx.BronzeGCPAppEngineApplication.Query().
		Where(
			bronzegcpappengineapplication.ProjectID(projectID),
			bronzegcpappengineapplication.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find stale applications: %w", err)
	}

	for _, app := range staleApps {
		if err := s.history.CloseHistory(ctx, tx, app.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for application %s: %w", app.ID, err)
		}

		if err := tx.BronzeGCPAppEngineApplication.DeleteOne(app).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete application %s: %w", app.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
