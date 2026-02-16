package application

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpappengineapplication"
)

// HistoryService manages App Engine application history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new App Engine application.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ApplicationData, now time.Time) error {
	create := tx.BronzeHistoryGCPAppEngineApplication.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetAuthDomain(data.AuthDomain).
		SetLocationID(data.LocationID).
		SetCodeBucket(data.CodeBucket).
		SetDefaultCookieExpiration(data.DefaultCookieExpiration).
		SetServingStatus(data.ServingStatus).
		SetDefaultHostname(data.DefaultHostname).
		SetDefaultBucket(data.DefaultBucket).
		SetGcrDomain(data.GcrDomain).
		SetDatabaseType(data.DatabaseType).
		SetProjectID(data.ProjectID)

	if data.FeatureSettingsJSON != nil {
		create.SetFeatureSettingsJSON(data.FeatureSettingsJSON)
	}
	if data.IapJSON != nil {
		create.SetIapJSON(data.IapJSON)
	}
	if data.DispatchRulesJSON != nil {
		create.SetDispatchRulesJSON(data.DispatchRulesJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create App Engine application history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed App Engine application.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPAppEngineApplication, new *ApplicationData, diff *ApplicationDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAppEngineApplication.Query().
		Where(
			bronzehistorygcpappengineapplication.ResourceID(old.ID),
			bronzehistorygcpappengineapplication.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current App Engine application history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPAppEngineApplication.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current App Engine application history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPAppEngineApplication.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetAuthDomain(new.AuthDomain).
			SetLocationID(new.LocationID).
			SetCodeBucket(new.CodeBucket).
			SetDefaultCookieExpiration(new.DefaultCookieExpiration).
			SetServingStatus(new.ServingStatus).
			SetDefaultHostname(new.DefaultHostname).
			SetDefaultBucket(new.DefaultBucket).
			SetGcrDomain(new.GcrDomain).
			SetDatabaseType(new.DatabaseType).
			SetProjectID(new.ProjectID)

		if new.FeatureSettingsJSON != nil {
			create.SetFeatureSettingsJSON(new.FeatureSettingsJSON)
		}
		if new.IapJSON != nil {
			create.SetIapJSON(new.IapJSON)
		}
		if new.DispatchRulesJSON != nil {
			create.SetDispatchRulesJSON(new.DispatchRulesJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new App Engine application history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted App Engine application.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAppEngineApplication.Query().
		Where(
			bronzehistorygcpappengineapplication.ResourceID(resourceID),
			bronzehistorygcpappengineapplication.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current App Engine application history: %w", err)
	}

	err = tx.BronzeHistoryGCPAppEngineApplication.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close App Engine application history: %w", err)
	}

	return nil
}
