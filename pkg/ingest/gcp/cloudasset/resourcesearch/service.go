package resourcesearch

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcloudassetresourcesearch"
)

// Service handles resource search ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new resource search ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of resource search ingestion.
type IngestResult struct {
	ResourceCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches resource search results from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch resource search results from GCP
	rawResources, err := s.client.SearchAllResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search resources: %w", err)
	}

	// Convert to resource data
	resourceDataList := make([]*ResourceSearchData, 0, len(rawResources))
	for _, raw := range rawResources {
		data := ConvertResourceSearch(raw.OrgName, raw.Resource, collectedAt)
		resourceDataList = append(resourceDataList, data)
	}

	// Save to database
	if err := s.saveResources(ctx, resourceDataList); err != nil {
		return nil, fmt.Errorf("failed to save resource search results: %w", err)
	}

	return &IngestResult{
		ResourceCount:  len(resourceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveResources saves resource search results to the database with history tracking.
func (s *Service) saveResources(ctx context.Context, resources []*ResourceSearchData) error {
	if len(resources) == 0 {
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

	for _, resourceData := range resources {
		// Load existing resource
		existing, err := tx.BronzeGCPCloudAssetResourceSearch.Query().
			Where(bronzegcpcloudassetresourcesearch.ID(resourceData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing resource search %s: %w", resourceData.ID, err)
		}

		// Compute diff
		diff := DiffResourceSearchData(existing, resourceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPCloudAssetResourceSearch.UpdateOneID(resourceData.ID).
				SetCollectedAt(resourceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for resource search %s: %w", resourceData.ID, err)
			}
			continue
		}

		// Create or update resource
		if existing == nil {
			create := tx.BronzeGCPCloudAssetResourceSearch.Create().
				SetID(resourceData.ID).
				SetAssetType(resourceData.AssetType).
				SetOrganizationID(resourceData.OrganizationID).
				SetCollectedAt(resourceData.CollectedAt).
				SetFirstCollectedAt(resourceData.CollectedAt)

			if resourceData.Project != "" {
				create.SetProject(resourceData.Project)
			}
			if resourceData.DisplayName != "" {
				create.SetDisplayName(resourceData.DisplayName)
			}
			if resourceData.Description != "" {
				create.SetDescription(resourceData.Description)
			}
			if resourceData.Location != "" {
				create.SetLocation(resourceData.Location)
			}
			if resourceData.LabelsJSON != nil {
				create.SetLabelsJSON(resourceData.LabelsJSON)
			}
			if resourceData.NetworkTagsJSON != nil {
				create.SetNetworkTagsJSON(resourceData.NetworkTagsJSON)
			}
			if resourceData.AdditionalAttributesJSON != nil {
				create.SetAdditionalAttributesJSON(resourceData.AdditionalAttributesJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create resource search %s: %w", resourceData.ID, err)
			}
		} else {
			update := tx.BronzeGCPCloudAssetResourceSearch.UpdateOneID(resourceData.ID).
				SetAssetType(resourceData.AssetType).
				SetOrganizationID(resourceData.OrganizationID).
				SetCollectedAt(resourceData.CollectedAt)

			if resourceData.Project != "" {
				update.SetProject(resourceData.Project)
			}
			if resourceData.DisplayName != "" {
				update.SetDisplayName(resourceData.DisplayName)
			}
			if resourceData.Description != "" {
				update.SetDescription(resourceData.Description)
			}
			if resourceData.Location != "" {
				update.SetLocation(resourceData.Location)
			}
			if resourceData.LabelsJSON != nil {
				update.SetLabelsJSON(resourceData.LabelsJSON)
			}
			if resourceData.NetworkTagsJSON != nil {
				update.SetNetworkTagsJSON(resourceData.NetworkTagsJSON)
			}
			if resourceData.AdditionalAttributesJSON != nil {
				update.SetAdditionalAttributesJSON(resourceData.AdditionalAttributesJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update resource search %s: %w", resourceData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, resourceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for resource search %s: %w", resourceData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, resourceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for resource search %s: %w", resourceData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleResources removes resource search results that were not collected in the latest run.
func (s *Service) DeleteStaleResources(ctx context.Context, collectedAt time.Time) error {
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

	staleResources, err := tx.BronzeGCPCloudAssetResourceSearch.Query().
		Where(bronzegcpcloudassetresourcesearch.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, resource := range staleResources {
		if err := s.history.CloseHistory(ctx, tx, resource.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for resource search %s: %w", resource.ID, err)
		}

		if err := tx.BronzeGCPCloudAssetResourceSearch.DeleteOne(resource).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete resource search %s: %w", resource.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
