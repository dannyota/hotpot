package machine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	entmachine "danny.vn/hotpot/pkg/storage/ent/inventory/machine"
	"danny.vn/hotpot/pkg/storage/ent/inventory/machine/inventorymachine"
	"danny.vn/hotpot/pkg/storage/ent/inventory/machine/inventorymachinebronzelink"
	"danny.vn/hotpot/pkg/storage/ent/inventory/machine/inventorymachinenormalized"
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

	// Load existing normalized IDs for this provider to detect stale rows.
	existing, err := a.entClient.InventoryMachineNormalized.Query().
		Where(inventorymachinenormalized.ProviderEQ(params.ProviderKey)).
		Select(inventorymachinenormalized.FieldID).
		Strings(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing normalized for %s: %w", params.ProviderKey, err)
	}

	// Batch upsert all normalized rows.
	for _, rec := range records {
		activeIDs[rec.ResourceID()] = true
	}
	if err := a.upsertNormalizedBatch(ctx, records, now); err != nil {
		return nil, err
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
		deleted, err = a.entClient.InventoryMachineNormalized.Delete().
			Where(inventorymachinenormalized.IDIn(staleIDs...)).
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

// upsertNormalizedBatch bulk-upserts a batch of records via raw SQL INSERT...ON CONFLICT.
func (a *Activities) upsertNormalizedBatch(ctx context.Context, batch []NormalizedMachine, now time.Time) error {
	if len(batch) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO silver.inventory_machine_normalized
		(resource_id, provider, is_base, bronze_table, bronze_resource_id,
		 hostname, os_type, os_name, status, internal_ip, external_ip,
		 environment, cloud_project, cloud_zone, cloud_machine_type,
		 created, merge_keys_json, collected_at, first_collected_at, normalized_at)
		VALUES `)

	const cols = 20
	args := make([]any, 0, len(batch)*cols)
	for i, rec := range batch {
		if i > 0 {
			b.WriteByte(',')
		}
		base := i * cols
		b.WriteByte('(')
		for j := range cols {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(base + j + 1))
		}
		b.WriteByte(')')
		mergeKeysJSON, _ := json.Marshal(rec.MergeKeys)
		args = append(args,
			rec.ResourceID(), rec.Provider, rec.IsBase, rec.BronzeTable, rec.BronzeResourceID,
			rec.Hostname, rec.OSType, rec.OSName, rec.Status, rec.InternalIP, rec.ExternalIP,
			rec.Environment, rec.CloudProject, rec.CloudZone, rec.CloudMachineType,
			rec.Created, string(mergeKeysJSON), rec.CollectedAt, rec.FirstCollectedAt, now,
		)
	}
	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		provider = EXCLUDED.provider,
		is_base = EXCLUDED.is_base,
		bronze_table = EXCLUDED.bronze_table,
		bronze_resource_id = EXCLUDED.bronze_resource_id,
		hostname = EXCLUDED.hostname,
		os_type = EXCLUDED.os_type,
		os_name = EXCLUDED.os_name,
		status = EXCLUDED.status,
		internal_ip = EXCLUDED.internal_ip,
		external_ip = EXCLUDED.external_ip,
		environment = EXCLUDED.environment,
		cloud_project = EXCLUDED.cloud_project,
		cloud_zone = EXCLUDED.cloud_zone,
		cloud_machine_type = EXCLUDED.cloud_machine_type,
		created = EXCLUDED.created,
		merge_keys_json = EXCLUDED.merge_keys_json,
		collected_at = EXCLUDED.collected_at,
		first_collected_at = EXCLUDED.first_collected_at,
		normalized_at = EXCLUDED.normalized_at`)

	_, err := a.db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert normalized batch: %w", err)
	}
	return nil
}

// MergeMachinesResult holds merge statistics.
type MergeMachinesResult struct {
	Created int
	Updated int
	Deleted int
}

// MergeMachines reads all normalized rows, runs dedup, and writes to silver.inventory_machines.
func (a *Activities) MergeMachines(ctx context.Context) (*MergeMachinesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting machine merge")

	// Read all normalized rows.
	normalizedRows, err := a.entClient.InventoryMachineNormalized.Query().All(ctx)
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
			Created:          r.Created,
			CollectedAt:      r.CollectedAt,
			FirstCollectedAt: r.FirstCollectedAt,
			MergeKeys:        r.MergeKeysJSON,
		})
	}

	// Run merge engine.
	merged := MergeMachines(rows, a.providerOrder)

	// Load existing machines with bronze links for stable ID matching.
	existingMachines, err := a.entClient.InventoryMachine.Query().
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

		// Wrap create/update + bronze link insert in a transaction.
		tx, err := a.entClient.Tx(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin transaction for %s: %w", machineID, err)
		}

		if isNew {
			// Create new machine.
			err = tx.InventoryMachine.Create().
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
				SetNillableCreated(m.Created).
				SetCollectedAt(m.CollectedAt).
				SetFirstCollectedAt(m.FirstCollectedAt).
				SetNormalizedAt(now).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("create machine %s: %w", machineID, err)
			}
		} else {
			// Update existing machine.
			update := tx.InventoryMachine.UpdateOneID(machineID).
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
				SetNormalizedAt(now)
			if m.Created != nil {
				update = update.SetCreated(*m.Created)
			} else {
				update = update.ClearCreated()
			}
			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("update machine %s: %w", machineID, err)
			}

			// Delete old bronze links before inserting new ones.
			_, err = tx.InventoryMachineBronzeLink.Delete().
				Where(inventorymachinebronzelink.HasMachineWith(inventorymachine.IDEQ(machineID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("delete old bronze links for %s: %w", machineID, err)
			}
		}

		// Bulk insert bronze links.
		builders := make([]*entmachine.InventoryMachineBronzeLinkCreate, 0, len(m.BronzeLinks))
		for _, link := range m.BronzeLinks {
			builders = append(builders, tx.InventoryMachineBronzeLink.Create().
				SetProvider(link.Provider).
				SetBronzeTable(link.BronzeTable).
				SetBronzeResourceID(link.BronzeResourceID).
				SetMachineID(machineID))
		}
		if err = tx.InventoryMachineBronzeLink.CreateBulk(builders...).Exec(ctx); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create bronze links for %s: %w", machineID, err)
		}

		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit merge for %s: %w", machineID, err)
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
		_, err = a.entClient.InventoryMachineBronzeLink.Delete().
			Where(inventorymachinebronzelink.HasMachineWith(inventorymachine.IDIn(staleIDs...))).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale bronze links", "count", len(staleIDs), "error", err)
		}

		deleted, err = a.entClient.InventoryMachine.Delete().
			Where(inventorymachine.IDIn(staleIDs...)).
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
