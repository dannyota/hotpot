package router

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputerouter"
)

// Service handles GCP Compute router ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new router ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for router ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of router ingestion.
type IngestResult struct {
	ProjectID      string
	RouterCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches routers from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch routers from GCP
	routers, err := s.client.ListRouters(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list routers: %w", err)
	}

	// Convert to data structs
	routerDataList := make([]*RouterData, 0, len(routers))
	for _, r := range routers {
		data, err := ConvertRouter(r, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert router: %w", err)
		}
		routerDataList = append(routerDataList, data)
	}

	// Save to database
	if err := s.saveRouters(ctx, routerDataList); err != nil {
		return nil, fmt.Errorf("failed to save routers: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		RouterCount:    len(routerDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveRouters saves routers to the database with history tracking.
func (s *Service) saveRouters(ctx context.Context, routers []*RouterData) error {
	if len(routers) == 0 {
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

	for _, routerData := range routers {
		// Load existing router
		existing, err := tx.BronzeGCPComputeRouter.Query().
			Where(bronzegcpcomputerouter.ID(routerData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing router %s: %w", routerData.Name, err)
		}

		// Compute diff
		diff := DiffRouterData(existing, routerData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeRouter.UpdateOneID(routerData.ID).
				SetCollectedAt(routerData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for router %s: %w", routerData.Name, err)
			}
			continue
		}

		// Create or update router
		if existing == nil {
			// Create new router
			create := tx.BronzeGCPComputeRouter.Create().
				SetID(routerData.ID).
				SetName(routerData.Name).
				SetProjectID(routerData.ProjectID).
				SetCollectedAt(routerData.CollectedAt).
				SetFirstCollectedAt(routerData.CollectedAt)

			if routerData.Description != "" {
				create.SetDescription(routerData.Description)
			}
			if routerData.SelfLink != "" {
				create.SetSelfLink(routerData.SelfLink)
			}
			if routerData.CreationTimestamp != "" {
				create.SetCreationTimestamp(routerData.CreationTimestamp)
			}
			if routerData.Network != "" {
				create.SetNetwork(routerData.Network)
			}
			if routerData.Region != "" {
				create.SetRegion(routerData.Region)
			}
			if routerData.BgpAsn != 0 {
				create.SetBgpAsn(routerData.BgpAsn)
			}
			if routerData.BgpAdvertiseMode != "" {
				create.SetBgpAdvertiseMode(routerData.BgpAdvertiseMode)
			}
			if routerData.BgpAdvertisedGroupsJSON != nil {
				create.SetBgpAdvertisedGroupsJSON(routerData.BgpAdvertisedGroupsJSON)
			}
			if routerData.BgpAdvertisedIPRangesJSON != nil {
				create.SetBgpAdvertisedIPRangesJSON(routerData.BgpAdvertisedIPRangesJSON)
			}
			if routerData.BgpKeepaliveInterval != 0 {
				create.SetBgpKeepaliveInterval(routerData.BgpKeepaliveInterval)
			}
			if routerData.BgpPeersJSON != nil {
				create.SetBgpPeersJSON(routerData.BgpPeersJSON)
			}
			if routerData.InterfacesJSON != nil {
				create.SetInterfacesJSON(routerData.InterfacesJSON)
			}
			if routerData.NatsJSON != nil {
				create.SetNatsJSON(routerData.NatsJSON)
			}
			if routerData.EncryptedInterconnectRouter {
				create.SetEncryptedInterconnectRouter(routerData.EncryptedInterconnectRouter)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create router %s: %w", routerData.Name, err)
			}
		} else {
			// Update existing router
			update := tx.BronzeGCPComputeRouter.UpdateOneID(routerData.ID).
				SetName(routerData.Name).
				SetProjectID(routerData.ProjectID).
				SetCollectedAt(routerData.CollectedAt)

			if routerData.Description != "" {
				update.SetDescription(routerData.Description)
			}
			if routerData.SelfLink != "" {
				update.SetSelfLink(routerData.SelfLink)
			}
			if routerData.CreationTimestamp != "" {
				update.SetCreationTimestamp(routerData.CreationTimestamp)
			}
			if routerData.Network != "" {
				update.SetNetwork(routerData.Network)
			}
			if routerData.Region != "" {
				update.SetRegion(routerData.Region)
			}
			if routerData.BgpAsn != 0 {
				update.SetBgpAsn(routerData.BgpAsn)
			}
			if routerData.BgpAdvertiseMode != "" {
				update.SetBgpAdvertiseMode(routerData.BgpAdvertiseMode)
			}
			if routerData.BgpAdvertisedGroupsJSON != nil {
				update.SetBgpAdvertisedGroupsJSON(routerData.BgpAdvertisedGroupsJSON)
			}
			if routerData.BgpAdvertisedIPRangesJSON != nil {
				update.SetBgpAdvertisedIPRangesJSON(routerData.BgpAdvertisedIPRangesJSON)
			}
			if routerData.BgpKeepaliveInterval != 0 {
				update.SetBgpKeepaliveInterval(routerData.BgpKeepaliveInterval)
			}
			if routerData.BgpPeersJSON != nil {
				update.SetBgpPeersJSON(routerData.BgpPeersJSON)
			}
			if routerData.InterfacesJSON != nil {
				update.SetInterfacesJSON(routerData.InterfacesJSON)
			}
			if routerData.NatsJSON != nil {
				update.SetNatsJSON(routerData.NatsJSON)
			}
			if routerData.EncryptedInterconnectRouter {
				update.SetEncryptedInterconnectRouter(routerData.EncryptedInterconnectRouter)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update router %s: %w", routerData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, routerData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for router %s: %w", routerData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, routerData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for router %s: %w", routerData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleRouters removes routers that were not collected in the latest run.
// Also closes history records for deleted routers.
func (s *Service) DeleteStaleRouters(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale routers
	staleRouters, err := tx.BronzeGCPComputeRouter.Query().
		Where(
			bronzegcpcomputerouter.ProjectID(projectID),
			bronzegcpcomputerouter.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale router
	for _, r := range staleRouters {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, r.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for router %s: %w", r.ID, err)
		}

		// Delete router
		if err := tx.BronzeGCPComputeRouter.DeleteOne(r).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete router %s: %w", r.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
