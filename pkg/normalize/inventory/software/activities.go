package software

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	entsoftware "danny.vn/hotpot/pkg/storage/ent/software"
	"danny.vn/hotpot/pkg/storage/ent/software/inventorysoftware"
	"danny.vn/hotpot/pkg/storage/ent/software/inventorysoftwarebronzelink"
	"danny.vn/hotpot/pkg/storage/ent/software/inventorysoftwarenormalized"
)

const (
	batchSize   = 1000
	concurrency = 4
)

// Activities holds dependencies for normalize/merge Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entsoftware.Client
	db            *sql.DB
	providers     map[string]Provider
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entsoftware.Client, db *sql.DB, providers []Provider) *Activities {
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

// NormalizeSoftwareProviderActivity function reference for Temporal registration.
var NormalizeSoftwareProviderActivity = (*Activities).NormalizeSoftwareProvider

// MergeSoftwareActivity function reference for Temporal registration.
var MergeSoftwareActivity = (*Activities).MergeSoftware

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

// NormalizeSoftwareProvider loads bronze data for one provider, bulk-upserts
// to installed_software_normalized, and deletes stale rows.
func (a *Activities) NormalizeSoftwareProvider(ctx context.Context, params NormalizeProviderParams) (*NormalizeProviderResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing provider", "provider", params.ProviderKey)

	provider, ok := a.providers[params.ProviderKey]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", params.ProviderKey)
	}

	records, err := provider.Load(ctx, a.db)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", params.ProviderKey, err)
	}
	logger.Info("Loaded bronze records", "provider", params.ProviderKey, "count", len(records))

	now := time.Now()

	// Batch upsert with parallel workers.
	batches := makeBatches(records, batchSize)
	var done atomic.Int64
	total := int64(len(records))

	errCh := make(chan error, len(batches))
	sem := make(chan struct{}, concurrency)

	for _, batch := range batches {
		sem <- struct{}{}
		go func(batch []NormalizedSoftware) {
			defer func() { <-sem }()
			if err := a.upsertNormalizedBatch(ctx, batch, now); err != nil {
				errCh <- err
				return
			}
			n := done.Add(int64(len(batch)))
			activity.RecordHeartbeat(ctx, fmt.Sprintf("upserted %d/%d", n, total))
			logger.Info("Upsert progress", "provider", params.ProviderKey,
				"done", n, "total", total)
		}(batch)
	}
	// Wait for all workers.
	for range concurrency {
		sem <- struct{}{}
	}
	close(errCh)
	if err := <-errCh; err != nil {
		return nil, err
	}

	// Delete stale: anything for this provider not updated this run.
	deleted, err := a.entClient.InventorySoftwareNormalized.Delete().
		Where(
			inventorysoftwarenormalized.ProviderEQ(params.ProviderKey),
			inventorysoftwarenormalized.NormalizedAtLT(now),
		).Exec(ctx)
	if err != nil {
		slog.Warn("Failed to delete stale normalized rows",
			"provider", params.ProviderKey, "error", err)
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
func (a *Activities) upsertNormalizedBatch(ctx context.Context, batch []NormalizedSoftware, now time.Time) error {
	if len(batch) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO inventory.software_normalized
		(resource_id, provider, is_base, bronze_table, bronze_resource_id,
		 machine_id, name, version, publisher,
		 collected_at, first_collected_at, normalized_at)
		VALUES `)

	args := make([]any, 0, len(batch)*12)
	for i, rec := range batch {
		if i > 0 {
			b.WriteByte(',')
		}
		base := i * 12
		b.WriteByte('(')
		for j := range 12 {
			if j > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(base + j + 1))
		}
		b.WriteByte(')')
		args = append(args,
			rec.ResourceID(), rec.Provider, rec.IsBase, rec.BronzeTable, rec.BronzeResourceID,
			rec.MachineID, rec.Name, rec.Version, rec.Publisher,
			rec.CollectedAt, rec.FirstCollectedAt, now,
		)
	}
	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		machine_id = EXCLUDED.machine_id,
		name = EXCLUDED.name,
		version = EXCLUDED.version,
		publisher = EXCLUDED.publisher,
		collected_at = EXCLUDED.collected_at,
		first_collected_at = EXCLUDED.first_collected_at,
		normalized_at = EXCLUDED.normalized_at`)

	_, err := a.db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert normalized batch: %w", err)
	}
	return nil
}

// MergeSoftwareResult holds merge statistics.
type MergeSoftwareResult struct {
	Created int
	Updated int
	Deleted int
}

type bronzeLink struct {
	Provider         string
	BronzeTable      string
	BronzeResourceID string
}

type mergedEntry struct {
	ID        string // UUID — reused or generated
	MachineID string
	Name      string
	Version   string
	Publisher string
	CollectedAt      time.Time
	FirstCollectedAt time.Time
	BronzeLinks      []bronzeLink
}

// MergeSoftware reads all normalized rows, deduplicates by (machine_id, name)
// with S1 priority, and writes to inventory.software.
func (a *Activities) MergeSoftware(ctx context.Context) (*MergeSoftwareResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting installed software merge")

	// Read all normalized rows.
	normalizedRows, err := a.entClient.InventorySoftwareNormalized.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query normalized rows: %w", err)
	}
	logger.Info("Loaded normalized rows", "count", len(normalizedRows))

	// Convert ent models to domain types.
	rows := make([]NormalizedSoftware, 0, len(normalizedRows))
	for _, r := range normalizedRows {
		rows = append(rows, NormalizedSoftware{
			Provider:         r.Provider,
			IsBase:           r.IsBase,
			BronzeTable:      r.BronzeTable,
			BronzeResourceID: r.BronzeResourceID,
			MachineID:        r.MachineID,
			Name:             r.Name,
			Version:          r.Version,
			Publisher:        r.Publisher,
			CollectedAt:      r.CollectedAt,
			FirstCollectedAt: r.FirstCollectedAt,
		})
	}

	// Sort: is_base DESC (true first), then provider ASC for determinism.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].IsBase != rows[j].IsBase {
			return rows[i].IsBase
		}
		return rows[i].Provider < rows[j].Provider
	})

	// Deduplicate by (machine_id, name). First (base provider) wins for fields.
	mergedMap := make(map[string]*mergedEntry, len(rows))
	for _, row := range rows {
		key := row.MachineID + ":" + row.Name
		if existing, ok := mergedMap[key]; ok {
			existing.BronzeLinks = append(existing.BronzeLinks, bronzeLink{
				Provider:         row.Provider,
				BronzeTable:      row.BronzeTable,
				BronzeResourceID: row.BronzeResourceID,
			})
		} else {
			mergedMap[key] = &mergedEntry{
				MachineID:        row.MachineID,
				Name:             row.Name,
				Version:          row.Version,
				Publisher:        row.Publisher,
				CollectedAt:      row.CollectedAt,
				FirstCollectedAt: row.FirstCollectedAt,
				BronzeLinks: []bronzeLink{{
					Provider:         row.Provider,
					BronzeTable:      row.BronzeTable,
					BronzeResourceID: row.BronzeResourceID,
				}},
			}
		}
	}
	logger.Info("Deduplicated entries", "normalized", len(rows), "merged", len(mergedMap))

	// Load existing installed_software for stable UUID matching.
	existingRecords, err := a.entClient.InventorySoftware.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing installed software: %w", err)
	}
	logger.Info("Loaded existing software", "count", len(existingRecords))

	// Build map: "machine_id:name" → existing UUID.
	existingMap := make(map[string]string, len(existingRecords))
	existingIDs := make(map[string]bool, len(existingRecords))
	for _, sw := range existingRecords {
		key := sw.MachineID + ":" + sw.Name
		existingMap[key] = sw.ID
		existingIDs[sw.ID] = true
	}

	// Assign stable UUIDs.
	var created, updated int
	activeSoftwareIDs := make(map[string]bool, len(mergedMap))
	entries := make([]*mergedEntry, 0, len(mergedMap))
	for key, entry := range mergedMap {
		if id, ok := existingMap[key]; ok {
			entry.ID = id
			updated++
		} else {
			entry.ID = uuid.New().String()
			created++
		}
		activeSoftwareIDs[entry.ID] = true
		entries = append(entries, entry)
	}

	now := time.Now()

	// Batch upsert installed_software + bronze_links in parallel.
	batches := makeMergedBatches(entries, batchSize)
	var done atomic.Int64
	total := int64(len(entries))

	errCh := make(chan error, len(batches))
	sem := make(chan struct{}, concurrency)

	for _, batch := range batches {
		sem <- struct{}{}
		go func(batch []*mergedEntry) {
			defer func() { <-sem }()
			if err := a.upsertMergedBatch(ctx, batch, now); err != nil {
				errCh <- err
				return
			}
			n := done.Add(int64(len(batch)))
			activity.RecordHeartbeat(ctx, fmt.Sprintf("merged %d/%d", n, total))
			logger.Info("Merge progress", "done", n, "total", total)
		}(batch)
	}
	for range concurrency {
		sem <- struct{}{}
	}
	close(errCh)
	if err := <-errCh; err != nil {
		return nil, err
	}

	// Delete stale: bronze_links first, then software.
	var staleIDs []string
	for _, sw := range existingRecords {
		if !activeSoftwareIDs[sw.ID] {
			staleIDs = append(staleIDs, sw.ID)
		}
	}

	deleted := 0
	if len(staleIDs) > 0 {
		_, err = a.entClient.InventorySoftwareBronzeLink.Delete().
			Where(inventorysoftwarebronzelink.HasSoftwareWith(
				inventorysoftware.IDIn(staleIDs...))).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale bronze links", "count", len(staleIDs), "error", err)
		}
		deleted, err = a.entClient.InventorySoftware.Delete().
			Where(inventorysoftware.IDIn(staleIDs...)).
			Exec(ctx)
		if err != nil {
			slog.Warn("Failed to delete stale installed software", "count", len(staleIDs), "error", err)
		}
	}

	logger.Info("Installed software merge complete",
		"created", created,
		"updated", updated,
		"deleted", deleted,
		"total", len(entries))

	return &MergeSoftwareResult{
		Created: created,
		Updated: updated,
		Deleted: deleted,
	}, nil
}

// upsertMergedBatch bulk-upserts a batch of merged entries to installed_software
// and rebuilds their bronze_links, all in a single transaction.
func (a *Activities) upsertMergedBatch(ctx context.Context, batch []*mergedEntry, now time.Time) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Bulk upsert installed_software.
	{
		var b strings.Builder
		b.WriteString(`INSERT INTO inventory.software
			(resource_id, machine_id, name, version, publisher,
			 collected_at, first_collected_at, normalized_at)
			VALUES `)

		args := make([]any, 0, len(batch)*8)
		for i, e := range batch {
			if i > 0 {
				b.WriteByte(',')
			}
			base := i * 8
			b.WriteByte('(')
			for j := range 8 {
				if j > 0 {
					b.WriteByte(',')
				}
				b.WriteByte('$')
				b.WriteString(strconv.Itoa(base + j + 1))
			}
			b.WriteByte(')')
			args = append(args, e.ID, e.MachineID, e.Name, e.Version, e.Publisher,
				e.CollectedAt, e.FirstCollectedAt, now)
		}
		b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
			version = EXCLUDED.version,
			publisher = EXCLUDED.publisher,
			collected_at = EXCLUDED.collected_at,
			normalized_at = EXCLUDED.normalized_at`)

		if _, err := tx.ExecContext(ctx, b.String(), args...); err != nil {
			return fmt.Errorf("upsert installed_software batch: %w", err)
		}
	}

	// 2. Delete old bronze_links for this batch.
	{
		ids := make([]any, len(batch))
		placeholders := make([]string, len(batch))
		for i, e := range batch {
			ids[i] = e.ID
			placeholders[i] = "$" + strconv.Itoa(i+1)
		}
		q := `DELETE FROM inventory.software_bronze_links
			WHERE inventory_software_bronze_links IN (` + strings.Join(placeholders, ",") + `)`
		if _, err := tx.ExecContext(ctx, q, ids...); err != nil {
			return fmt.Errorf("delete old bronze links: %w", err)
		}
	}

	// 3. Bulk insert new bronze_links.
	{
		var totalLinks int
		for _, e := range batch {
			totalLinks += len(e.BronzeLinks)
		}
		if totalLinks > 0 {
			var b strings.Builder
			b.WriteString(`INSERT INTO inventory.software_bronze_links
				(provider, bronze_table, bronze_resource_id, inventory_software_bronze_links)
				VALUES `)

			args := make([]any, 0, totalLinks*4)
			idx := 0
			for _, e := range batch {
				for _, link := range e.BronzeLinks {
					if idx > 0 {
						b.WriteByte(',')
					}
					base := idx * 4
					b.WriteByte('(')
					for j := range 4 {
						if j > 0 {
							b.WriteByte(',')
						}
						b.WriteByte('$')
						b.WriteString(strconv.Itoa(base + j + 1))
					}
					b.WriteByte(')')
					args = append(args, link.Provider, link.BronzeTable, link.BronzeResourceID, e.ID)
					idx++
				}
			}
			if _, err := tx.ExecContext(ctx, b.String(), args...); err != nil {
				return fmt.Errorf("insert bronze links: %w", err)
			}
		}
	}

	return tx.Commit()
}

// makeBatches splits a slice into chunks of the given size.
func makeBatches(records []NormalizedSoftware, size int) [][]NormalizedSoftware {
	var batches [][]NormalizedSoftware
	for i := 0; i < len(records); i += size {
		end := i + size
		if end > len(records) {
			end = len(records)
		}
		batches = append(batches, records[i:end])
	}
	return batches
}

func makeMergedBatches(entries []*mergedEntry, size int) [][]*mergedEntry {
	var batches [][]*mergedEntry
	for i := 0; i < len(entries); i += size {
		end := i + size
		if end > len(entries) {
			end = len(entries)
		}
		batches = append(batches, entries[i:end])
	}
	return batches
}

// heartbeatLoop sends periodic heartbeats until the done channel is closed.
// Use this for long-running phases that don't have natural batch boundaries.
func heartbeatLoop(ctx context.Context, done <-chan struct{}, wg *sync.WaitGroup, msg string) {
	defer wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			activity.RecordHeartbeat(ctx, msg)
		}
	}
}
