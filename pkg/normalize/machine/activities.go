package machine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	entmachine "github.com/dannyota/hotpot/pkg/storage/ent/machine"
	"github.com/dannyota/hotpot/pkg/storage/ent/machine/silvermachine"
	"github.com/dannyota/hotpot/pkg/storage/ent/machine/silvermachinebronzelink"
	"github.com/dannyota/hotpot/pkg/storage/ent/machine/silvermachinenormalized"
)

func generateID() string {
	return uuid.New().String()
}

// Activities holds dependencies for normalize/merge Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entmachine.Client
	db            *sql.DB
	providers     map[string]Provider
	providerOrder []string
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entmachine.Client, db *sql.DB, providers []Provider) *Activities {
	pmap := make(map[string]Provider, len(providers))
	order := make([]string, 0, len(providers))
	for _, p := range providers {
		pmap[p.Key()] = p
		order = append(order, p.Key())
	}
	return &Activities{
		configService: configService,
		entClient:     entClient,
		db:            db,
		providers:     pmap,
		providerOrder: order,
	}
}

// NormalizeMachineProviderActivity function reference for Temporal registration.
var NormalizeMachineProviderActivity = (*Activities).NormalizeMachineProvider

// MergeMachinesActivity function reference for Temporal registration.
var MergeMachinesActivity = (*Activities).MergeMachines

// NormalizeProviderParams identifies which provider to normalize.
type NormalizeProviderParams struct {
	ProviderKey string
}

// NormalizeProviderResult holds normalization statistics.
type NormalizeProviderResult struct {
	ProviderKey string
	Upserted    int
	Deleted     int
}

// NormalizeProvider loads bronze data for one provider, creates/updates machine_normalized,
// and deletes stale rows.
func (a *Activities) NormalizeMachineProvider(ctx context.Context, params NormalizeProviderParams) (*NormalizeProviderResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing provider", "provider", params.ProviderKey)

	provider, ok := a.providers[params.ProviderKey]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", params.ProviderKey)
	}

	// Load and normalize bronze data.
	records, err := provider.Load(ctx, a.db)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", params.ProviderKey, err)
	}

	now := time.Now()
	activeIDs := make(map[string]bool, len(records))

	// Load existing normalized rows for this provider to distinguish create vs update.
	existingIDs := make(map[string]bool)
	existing, err := a.entClient.SilverMachineNormalized.Query().
		Where(silvermachinenormalized.ProviderEQ(params.ProviderKey)).
		Select(silvermachinenormalized.FieldID).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing normalized for %s: %w", params.ProviderKey, err)
	}
	for _, id := range existing {
		existingIDs[id] = true
	}

	// Create or update normalized rows.
	for _, rec := range records {
		rid := rec.ResourceID()
		activeIDs[rid] = true

		if existingIDs[rid] {
			// Update existing row.
			_, err := a.entClient.SilverMachineNormalized.UpdateOneID(rid).
				SetProvider(rec.Provider).
				SetIsBase(rec.IsBase).
				SetBronzeTable(rec.BronzeTable).
				SetBronzeResourceID(rec.BronzeResourceID).
				SetHostname(rec.Hostname).
				SetOsType(rec.OSType).
				SetOsName(rec.OSName).
				SetStatus(rec.Status).
				SetInternalIP(rec.InternalIP).
				SetExternalIP(rec.ExternalIP).
				SetEnvironment(rec.Environment).
				SetCloudProject(rec.CloudProject).
				SetCloudZone(rec.CloudZone).
				SetCloudMachineType(rec.CloudMachineType).
				SetMergeKeysJSON(rec.MergeKeys).
				SetCollectedAt(rec.CollectedAt).
				SetFirstCollectedAt(rec.FirstCollectedAt).
				SetNormalizedAt(now).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update normalized %s: %w", rid, err)
			}
		} else {
			// Create new row.
			err := a.entClient.SilverMachineNormalized.Create().
				SetID(rid).
				SetProvider(rec.Provider).
				SetIsBase(rec.IsBase).
				SetBronzeTable(rec.BronzeTable).
				SetBronzeResourceID(rec.BronzeResourceID).
				SetHostname(rec.Hostname).
				SetOsType(rec.OSType).
				SetOsName(rec.OSName).
				SetStatus(rec.Status).
				SetInternalIP(rec.InternalIP).
				SetExternalIP(rec.ExternalIP).
				SetEnvironment(rec.Environment).
				SetCloudProject(rec.CloudProject).
				SetCloudZone(rec.CloudZone).
				SetCloudMachineType(rec.CloudMachineType).
				SetMergeKeysJSON(rec.MergeKeys).
				SetCollectedAt(rec.CollectedAt).
				SetFirstCollectedAt(rec.FirstCollectedAt).
				SetNormalizedAt(now).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create normalized %s: %w", rid, err)
			}
		}
	}

	// Delete stale normalized rows for this provider.
	var staleIDs []string
	for _, id := range existing {
		if !activeIDs[id] {
			staleIDs = append(staleIDs, id)
		}
	}

	deleted := 0
	if len(staleIDs) > 0 {
		deleted, err = a.entClient.SilverMachineNormalized.Delete().
			Where(silvermachinenormalized.IDIn(staleIDs...)).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale normalized rows",
				"provider", params.ProviderKey, "count", len(staleIDs), "error", err)
		}
	}

	logger.Info("Normalized provider",
		"provider", params.ProviderKey,
		"upserted", len(records),
		"deleted", deleted)

	return &NormalizeProviderResult{
		ProviderKey: params.ProviderKey,
		Upserted:    len(records),
		Deleted:     deleted,
	}, nil
}

// MergeMachinesResult holds merge statistics.
type MergeMachinesResult struct {
	Created int
	Updated int
	Deleted int
}

// MergeMachines reads all normalized rows, runs dedup, and writes to silver.machines.
func (a *Activities) MergeMachines(ctx context.Context) (*MergeMachinesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting machine merge")

	// Read all normalized rows.
	normalizedRows, err := a.entClient.SilverMachineNormalized.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query normalized rows: %w", err)
	}

	// Convert ent models to domain types.
	rows := make([]NormalizedMachine, 0, len(normalizedRows))
	for _, r := range normalizedRows {
		rows = append(rows, NormalizedMachine{
			Provider:         r.Provider,
			IsBase:           r.IsBase,
			BronzeTable:      r.BronzeTable,
			BronzeResourceID: r.BronzeResourceID,
			Hostname:         r.Hostname,
			OSType:           r.OsType,
			OSName:           r.OsName,
			Status:           r.Status,
			InternalIP:       r.InternalIP,
			ExternalIP:       r.ExternalIP,
			Environment:      r.Environment,
			CloudProject:     r.CloudProject,
			CloudZone:        r.CloudZone,
			CloudMachineType: r.CloudMachineType,
			CollectedAt:      r.CollectedAt,
			FirstCollectedAt: r.FirstCollectedAt,
			MergeKeys:        r.MergeKeysJSON,
		})
	}

	// Run merge engine.
	merged := MergeMachines(rows, a.providerOrder)

	// Load existing machines with bronze links for stable ID matching.
	existingMachines, err := a.entClient.SilverMachine.Query().
		WithBronzeLinks().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing machines: %w", err)
	}

	// Build map: bronze_resource_id → existing machine resource_id.
	bronzeToMachineID := make(map[string]string)
	existingMachineIDs := make(map[string]bool, len(existingMachines))
	for _, m := range existingMachines {
		existingMachineIDs[m.ID] = true
		for _, link := range m.Edges.BronzeLinks {
			bronzeToMachineID[link.BronzeResourceID] = m.ID
		}
	}

	now := time.Now()
	var created, updated int

	// Track which existing machine IDs are still active.
	activeMachineIDs := make(map[string]bool)

	for _, m := range merged {
		// Find stable ID: check if any bronze_resource_id matches an existing machine.
		var machineID string
		for _, link := range m.BronzeLinks {
			if id, ok := bronzeToMachineID[link.BronzeResourceID]; ok {
				machineID = id
				break
			}
		}

		isNew := machineID == ""
		if isNew {
			machineID = generateID()
			created++
		} else {
			updated++
		}
		activeMachineIDs[machineID] = true

		if isNew {
			// Create new machine.
			err := a.entClient.SilverMachine.Create().
				SetID(machineID).
				SetHostname(m.Hostname).
				SetOsType(m.OSType).
				SetOsName(m.OSName).
				SetStatus(m.Status).
				SetInternalIP(m.InternalIP).
				SetExternalIP(m.ExternalIP).
				SetEnvironment(m.Environment).
				SetCloudProject(m.CloudProject).
				SetCloudZone(m.CloudZone).
				SetCloudMachineType(m.CloudMachineType).
				SetCollectedAt(m.CollectedAt).
				SetFirstCollectedAt(m.FirstCollectedAt).
				SetNormalizedAt(now).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create machine %s: %w", machineID, err)
			}
		} else {
			// Update existing machine.
			_, err := a.entClient.SilverMachine.UpdateOneID(machineID).
				SetHostname(m.Hostname).
				SetOsType(m.OSType).
				SetOsName(m.OSName).
				SetStatus(m.Status).
				SetInternalIP(m.InternalIP).
				SetExternalIP(m.ExternalIP).
				SetEnvironment(m.Environment).
				SetCloudProject(m.CloudProject).
				SetCloudZone(m.CloudZone).
				SetCloudMachineType(m.CloudMachineType).
				SetCollectedAt(m.CollectedAt).
				SetNormalizedAt(now).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update machine %s: %w", machineID, err)
			}

			// Delete old bronze links before inserting new ones.
			_, err = a.entClient.SilverMachineBronzeLink.Delete().
				Where(silvermachinebronzelink.HasMachineWith(silvermachine.IDEQ(machineID))).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("delete old bronze links for %s: %w", machineID, err)
			}
		}

		// Insert bronze links.
		for _, link := range m.BronzeLinks {
			err := a.entClient.SilverMachineBronzeLink.Create().
				SetProvider(link.Provider).
				SetBronzeTable(link.BronzeTable).
				SetBronzeResourceID(link.BronzeResourceID).
				SetMachineID(machineID).
				Exec(ctx)
			if err != nil {
				return nil, fmt.Errorf("create bronze link for %s: %w", machineID, err)
			}
		}
	}

	// Delete stale machines.
	var staleIDs []string
	for _, m := range existingMachines {
		if !activeMachineIDs[m.ID] {
			staleIDs = append(staleIDs, m.ID)
		}
	}

	deleted := 0
	if len(staleIDs) > 0 {
		// Delete links first, then machines.
		_, err = a.entClient.SilverMachineBronzeLink.Delete().
			Where(silvermachinebronzelink.HasMachineWith(silvermachine.IDIn(staleIDs...))).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale bronze links", "count", len(staleIDs), "error", err)
		}

		deleted, err = a.entClient.SilverMachine.Delete().
			Where(silvermachine.IDIn(staleIDs...)).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale machines", "count", len(staleIDs), "error", err)
		}
	}

	logger.Info("Machine merge complete",
		"created", created,
		"updated", updated,
		"deleted", deleted,
		"total", len(merged))

	return &MergeMachinesResult{
		Created: created,
		Updated: updated,
		Deleted: deleted,
	}, nil
}
