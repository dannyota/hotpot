package apiendpoint

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	entapiendpoint "danny.vn/hotpot/pkg/storage/ent/inventory/apiendpoint"
	"danny.vn/hotpot/pkg/storage/ent/inventory/apiendpoint/inventoryapiendpoint"
	"danny.vn/hotpot/pkg/storage/ent/inventory/apiendpoint/inventoryapiendpointbronzelink"
)

// Activities holds dependencies for API endpoint normalize activities.
type Activities struct {
	configService *config.Service
	entClient     *entapiendpoint.Client
	db            *sql.DB
	providers     map[string]Provider
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entapiendpoint.Client, db *sql.DB, providers []Provider) *Activities {
	pmap := make(map[string]Provider, len(providers))
	for _, p := range providers {
		pmap[p.Key()] = p
	}
	return &Activities{
		configService: configService,
		entClient:     entClient,
		db:            db,
		providers:     pmap,
	}
}

// Activity function references for Temporal registration.
var NormalizeApiEndpointsActivity = (*Activities).NormalizeApiEndpoints

// NormalizeApiEndpointsResult holds normalization statistics.
type NormalizeApiEndpointsResult struct {
	Created int
	Updated int
	Deleted int
}

// NormalizeApiEndpoints loads from all providers, upserts endpoints with bronze links, deletes stale rows.
func (a *Activities) NormalizeApiEndpoints(ctx context.Context) (*NormalizeApiEndpointsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing API endpoints")

	// Load all provider records.
	var allRecords []NormalizedApiEndpoint
	for _, provider := range a.providers {
		records, err := provider.Load(ctx, a.db)
		if err != nil {
			return nil, fmt.Errorf("load provider %s: %w", provider.Key(), err)
		}
		allRecords = append(allRecords, records...)
	}

	// Load existing endpoints with bronze links for stable ID matching.
	existingEndpoints, err := a.entClient.InventoryApiEndpoint.Query().
		WithBronzeLinks().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing api endpoints: %w", err)
	}

	// Build map: bronze_resource_id → existing endpoint ID.
	bronzeToEndpointID := make(map[string]string)
	existingEndpointIDs := make(map[string]bool, len(existingEndpoints))
	for _, ep := range existingEndpoints {
		existingEndpointIDs[ep.ID] = true
		for _, link := range ep.Edges.BronzeLinks {
			bronzeToEndpointID[link.BronzeResourceID] = ep.ID
		}
	}

	now := time.Now()
	var created, updated int
	activeEndpointIDs := make(map[string]bool)

	for _, rec := range allRecords {
		// Find stable ID via bronze link lookup.
		endpointID, isExisting := bronzeToEndpointID[rec.BronzeResourceID]
		if !isExisting {
			endpointID = uuid.New().String()
		}
		activeEndpointIDs[endpointID] = true

		if isExisting {
			// Update existing endpoint.
			_, err := a.entClient.InventoryApiEndpoint.UpdateOneID(endpointID).
				SetNillableName(nilIfEmpty(rec.Name)).
				SetNillableService(nilIfEmpty(rec.Service)).
				SetURIPattern(rec.URIPattern).
				SetMethods(rec.Methods).
				SetIsActive(rec.IsActive).
				SetNillableAccessLevel(nilIfEmpty(rec.AccessLevel)).
				SetCollectedAt(rec.CollectedAt).
				SetNormalizedAt(now).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update endpoint %s: %w", endpointID, err)
			}

			// Delete old bronze links before inserting new ones.
			_, err = a.entClient.InventoryApiEndpointBronzeLink.Delete().
				Where(inventoryapiendpointbronzelink.HasAPIEndpointWith(inventoryapiendpoint.IDEQ(endpointID))).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("delete old bronze links for %s: %w", endpointID, err)
			}
			updated++
		} else {
			// Create new endpoint.
			err := a.entClient.InventoryApiEndpoint.Create().
				SetID(endpointID).
				SetNillableName(nilIfEmpty(rec.Name)).
				SetNillableService(nilIfEmpty(rec.Service)).
				SetURIPattern(rec.URIPattern).
				SetMethods(rec.Methods).
				SetIsActive(rec.IsActive).
				SetNillableAccessLevel(nilIfEmpty(rec.AccessLevel)).
				SetCollectedAt(rec.CollectedAt).
				SetFirstCollectedAt(rec.FirstCollectedAt).
				SetNormalizedAt(now).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create endpoint %s: %w", endpointID, err)
			}
			created++
		}

		// Insert bronze link.
		err := a.entClient.InventoryApiEndpointBronzeLink.Create().
			SetProvider(rec.Provider).
			SetBronzeTable(rec.BronzeTable).
			SetBronzeResourceID(rec.BronzeResourceID).
			SetAPIEndpointID(endpointID).
			Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("create bronze link for %s: %w", endpointID, err)
		}
	}

	// Delete stale endpoints.
	var staleIDs []string
	for _, ep := range existingEndpoints {
		if !activeEndpointIDs[ep.ID] {
			staleIDs = append(staleIDs, ep.ID)
		}
	}

	deleted := 0
	if len(staleIDs) > 0 {
		// Delete links first, then endpoints (FK order).
		_, err = a.entClient.InventoryApiEndpointBronzeLink.Delete().
			Where(inventoryapiendpointbronzelink.HasAPIEndpointWith(inventoryapiendpoint.IDIn(staleIDs...))).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale bronze links", "count", len(staleIDs), "error", err)
		}

		deleted, err = a.entClient.InventoryApiEndpoint.Delete().
			Where(inventoryapiendpoint.IDIn(staleIDs...)).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale endpoints", "count", len(staleIDs), "error", err)
		}
	}

	logger.Info("API endpoint normalization complete",
		"created", created,
		"updated", updated,
		"deleted", deleted)

	return &NormalizeApiEndpointsResult{
		Created: created,
		Updated: updated,
		Deleted: deleted,
	}, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
