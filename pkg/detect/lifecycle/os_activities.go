package lifecycle

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

// Activity function references for OS lifecycle Temporal registration.
var (
	MatchOSLifecycleActivity = (*Activities).MatchOSLifecycle
	CleanupStaleOSActivity   = (*Activities).CleanupStaleOS
)

// --- Types ---

type machineForOS struct {
	machineID string
	hostname  string
	osType    string
	osName    string
}

type osLifecycleRow struct {
	machineID      string
	hostname       string
	osType         string
	osName         string
	eolProductSlug *string
	eolProductName *string
	eolCycle       *string
	eolDate        *time.Time
	eoasDate       *time.Time
	eoesDate       *time.Time
	eolStatus      string
	latestVersion  *string
}

// --- Activity 1: MatchOSLifecycle ---

// MatchOSLifecycleParams holds input for the MatchOSLifecycle activity.
type MatchOSLifecycleParams struct {
	RunTimestamp time.Time
}

// MatchOSLifecycleResult holds output from the MatchOSLifecycle activity.
type MatchOSLifecycleResult struct {
	Matched   int
	Unmatched int
	Total     int
}

// MatchOSLifecycle loads OS EOL products, matches machines against them,
// and writes results to gold.lifecycle_os.
func (a *Activities) MatchOSLifecycle(ctx context.Context, params MatchOSLifecycleParams) (*MatchOSLifecycleResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting MatchOSLifecycle activity")

	// 1. Load OS products from reference data.
	osSlugs, err := a.loadOSSlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("load OS slugs: %w", err)
	}
	logger.Info("Loaded OS products", "count", len(osSlugs))

	// 2. Load EOL cycles for OS products.
	slugList := make([]string, 0, len(osSlugs))
	for slug := range osSlugs {
		slugList = append(slugList, slug)
	}
	cycles, err := a.loadEOLCycles(ctx, slugList)
	if err != nil {
		return nil, fmt.Errorf("load EOL cycles: %w", err)
	}

	knownCycleSets := make(map[string]map[string]bool, len(cycles))
	for slug, slugCycles := range cycles {
		set := make(map[string]bool, len(slugCycles))
		for _, c := range slugCycles {
			set[c.Cycle] = true
		}
		knownCycleSets[slug] = set
	}

	winBuildToCycle := buildWindowsBuildMap(cycles["windows"])

	// 3. Load machines.
	machines, err := a.loadMachinesForOS(ctx)
	if err != nil {
		return nil, fmt.Errorf("load machines: %w", err)
	}
	logger.Info("Loaded machines", "count", len(machines))

	// 4. Match machines to OS products.
	var matched, unmatched int
	var rows []osLifecycleRow

	for _, m := range machines {
		slug, cycle, _ := parseOSName(m.osName, knownCycleSets, winBuildToCycle)

		row := osLifecycleRow{
			machineID: m.machineID,
			hostname:  m.hostname,
			osType:    m.osType,
			osName:    m.osName,
			eolStatus: "unknown",
		}

		if slug != "" {
			matched++
			productName := osSlugs[slug]
			if productName == "" {
				productName = slug
			}
			row.eolProductSlug = &slug
			row.eolProductName = &productName

			if cycle != "" {
				row.eolCycle = &cycle
				if slugCycles, ok := cycles[slug]; ok {
					for i := range slugCycles {
						if slugCycles[i].Cycle == cycle {
							row.eolDate = slugCycles[i].EOL
							row.eoasDate = slugCycles[i].EOAS
							row.eoesDate = slugCycles[i].EOES
							row.latestVersion = nilIfEmpty(slugCycles[i].Latest)
							row.eolStatus = determineEOLStatus(slugCycles[i].EOL, slugCycles[i].EOAS, slugCycles[i].EOES, params.RunTimestamp)
							break
						}
					}
				}
			}
		} else {
			unmatched++
		}

		rows = append(rows, row)
	}

	// 5. Bulk upsert.
	for i := 0; i < len(rows); i += batchSize {
		end := min(i+batchSize, len(rows))
		if err := a.upsertOSLifecycleBatch(ctx, rows[i:end], params.RunTimestamp); err != nil {
			return nil, fmt.Errorf("upsert OS lifecycle batch: %w", err)
		}
		activity.RecordHeartbeat(ctx, fmt.Sprintf("os %d/%d", end, len(rows)))
	}

	logger.Info("MatchOSLifecycle complete", "matched", matched, "unmatched", unmatched, "total", len(rows))
	return &MatchOSLifecycleResult{Matched: matched, Unmatched: unmatched, Total: len(rows)}, nil
}

// --- Activity 2: CleanupStaleOS ---

// CleanupStaleOSParams holds input for the CleanupStaleOS activity.
type CleanupStaleOSParams struct {
	RunTimestamp time.Time
}

// CleanupStaleOSResult holds output from the CleanupStaleOS activity.
type CleanupStaleOSResult struct {
	Deleted int
}

// CleanupStaleOS deletes gold.lifecycle_os rows not updated in this run.
func (a *Activities) CleanupStaleOS(ctx context.Context, params CleanupStaleOSParams) (*CleanupStaleOSResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting CleanupStaleOS activity")

	result, err := a.db.ExecContext(ctx,
		`DELETE FROM gold.lifecycle_os WHERE detected_at < $1`,
		params.RunTimestamp)
	if err != nil {
		return nil, fmt.Errorf("delete stale OS rows: %w", err)
	}

	deleted, _ := result.RowsAffected()
	logger.Info("CleanupStaleOS complete", "deleted", deleted)
	return &CleanupStaleOSResult{Deleted: int(deleted)}, nil
}

// --- Data loading ---

func (a *Activities) loadOSSlugs(ctx context.Context) (map[string]string, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT resource_id, name
		FROM bronze.reference_eol_products
		WHERE category = 'os'`)
	if err != nil {
		return nil, fmt.Errorf("query OS products: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var slug, name string
		if err := rows.Scan(&slug, &name); err != nil {
			return nil, fmt.Errorf("scan OS product: %w", err)
		}
		result[slug] = name
	}
	return result, rows.Err()
}

func (a *Activities) loadMachinesForOS(ctx context.Context) ([]machineForOS, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT resource_id, COALESCE(hostname, ''), COALESCE(os_type, ''),
		       COALESCE(os_name, '')
		FROM silver.inventory_machines`)
	if err != nil {
		return nil, fmt.Errorf("query machines: %w", err)
	}
	defer rows.Close()

	var result []machineForOS
	for rows.Next() {
		var m machineForOS
		if err := rows.Scan(&m.machineID, &m.hostname, &m.osType, &m.osName); err != nil {
			return nil, fmt.Errorf("scan machine: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// --- Bulk upsert ---

func (a *Activities) upsertOSLifecycleBatch(ctx context.Context, rows []osLifecycleRow, runTimestamp time.Time) error {
	if len(rows) == 0 {
		return nil
	}

	const cols = 15
	var b strings.Builder
	b.WriteString(`INSERT INTO gold.lifecycle_os
		(resource_id, detected_at, first_detected_at, machine_id, hostname,
		 os_type, os_name, eol_product_slug, eol_product_name,
		 eol_cycle, eol_date, eoas_date, eoes_date, eol_status, latest_version)
		VALUES `)

	args := make([]any, 0, len(rows)*cols)
	for i, r := range rows {
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

		args = append(args, r.machineID, runTimestamp, runTimestamp,
			r.machineID, nilIfEmpty(r.hostname),
			nilIfEmpty(r.osType), nilIfEmpty(r.osName),
			r.eolProductSlug, r.eolProductName,
			r.eolCycle, r.eolDate, r.eoasDate, r.eoesDate, r.eolStatus, r.latestVersion)
	}

	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		detected_at = EXCLUDED.detected_at,
		machine_id = EXCLUDED.machine_id,
		hostname = EXCLUDED.hostname,
		os_type = EXCLUDED.os_type,
		os_name = EXCLUDED.os_name,
		eol_product_slug = EXCLUDED.eol_product_slug,
		eol_product_name = EXCLUDED.eol_product_name,
		eol_cycle = EXCLUDED.eol_cycle,
		eol_date = EXCLUDED.eol_date,
		eoas_date = EXCLUDED.eoas_date,
		eoes_date = EXCLUDED.eoes_date,
		eol_status = EXCLUDED.eol_status,
		latest_version = EXCLUDED.latest_version`)

	_, err := a.db.ExecContext(ctx, b.String(), args...)
	return err
}
