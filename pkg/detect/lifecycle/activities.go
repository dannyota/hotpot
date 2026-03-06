package lifecycle

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	entlifecycle "danny.vn/hotpot/pkg/storage/ent/lifecycle"
)

const batchSize = 1000

// Activities holds dependencies for lifecycle detection Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entlifecycle.Client
	db            *sql.DB
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entlifecycle.Client, db *sql.DB) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		db:            db,
	}
}

// Activity function references for Temporal registration.
var (
	MatchProductsActivity  = (*Activities).MatchProducts
	ClassifyOSCoreActivity = (*Activities).ClassifyOSCore
	MarkUnmatchedActivity  = (*Activities).MarkUnmatched
	CleanupStaleActivity   = (*Activities).CleanupStale
)

// --- Types ---

type installedApp struct {
	MachineID string
	Name      string
	Version   string
}

type lifecycleRow struct {
	machineID      string
	name           string
	version        string
	classification string
	eolProductSlug *string
	eolProductName *string
	eolCategory    *string
	eolCycle       *string
	eolDate        *time.Time
	eoasDate       *time.Time
	eoesDate       *time.Time
	eolStatus      string
	latestVersion  *string
}

// --- Activity 1: MatchProducts ---

// MatchProductsParams holds input for the MatchProducts activity.
type MatchProductsParams struct {
	RunTimestamp time.Time
}

// MatchProductsResult holds output from the MatchProducts activity.
type MatchProductsResult struct {
	Matched      int
	MatchedNames []string
}

// MatchProducts loads product mappings, matches installed apps against them,
// and writes matched results to gold.lifecycle_software.
func (a *Activities) MatchProducts(ctx context.Context, params MatchProductsParams) (*MatchProductsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting MatchProducts activity")

	// 1. Load product mappings from bronze.reference_eol_identifiers + products.
	mappings, err := a.loadProductMappings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load product mappings: %w", err)
	}
	logger.Info("Loaded product mappings", "count", len(mappings))

	slugs := make([]string, len(mappings))
	for i := range mappings {
		slugs[i] = mappings[i].slug
	}

	// 2. Load EOL cycles.
	cycles, err := a.loadEOLCycles(ctx, slugs)
	if err != nil {
		return nil, fmt.Errorf("load eol cycles: %w", err)
	}

	knownCycleSets := make(map[string]map[string]bool, len(cycles))
	for slug, slugCycles := range cycles {
		set := make(map[string]bool, len(slugCycles))
		for _, c := range slugCycles {
			set[c.Cycle] = true
		}
		knownCycleSets[slug] = set
	}

	// 3. Load installed apps.
	apps, err := a.loadInstalledApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("load installed apps: %w", err)
	}
	logger.Info("Loaded installed apps", "count", len(apps))

	// 4. Match apps to products.
	matchedNames := make(map[string]bool)
	var rows []lifecycleRow

	for _, app := range apps {
		pm := matchAppToProduct(app.Name, mappings)
		if pm == nil {
			continue
		}
		matchedNames[app.Name] = true

		cycle := extractCycleFromMapping(app.Name, app.Version, pm, knownCycleSets[pm.slug])

		displayName := pm.name
		if displayName == "" {
			displayName = pm.slug
		}

		row := lifecycleRow{
			machineID:      app.MachineID,
			name:           app.Name,
			version:        app.Version,
			classification: "matched",
			eolProductSlug: &pm.slug,
			eolProductName: &displayName,
			eolCategory:    &pm.eolCategory,
			eolStatus:      "unknown",
		}

		if cycle != "" {
			row.eolCycle = &cycle
			if slugCycles, ok := cycles[pm.slug]; ok {
				for i := range slugCycles {
					if slugCycles[i].Cycle == cycle {
						row.eolDate = slugCycles[i].EOL
						row.eoasDate = slugCycles[i].EOAS
						row.eoesDate = slugCycles[i].EOES
						row.latestVersion = strPtr(slugCycles[i].Latest)
						row.eolStatus = determineEOLStatus(slugCycles[i].EOL, slugCycles[i].EOAS, slugCycles[i].EOES, params.RunTimestamp)
						break
					}
				}
			}
		}

		rows = append(rows, row)
	}

	// 5. Bulk upsert matched results.
	for i := 0; i < len(rows); i += batchSize {
		end := min(i+batchSize, len(rows))
		if err := a.upsertLifecycleBatch(ctx, rows[i:end], params.RunTimestamp); err != nil {
			return nil, fmt.Errorf("upsert matched batch: %w", err)
		}
		activity.RecordHeartbeat(ctx, fmt.Sprintf("matched %d/%d", end, len(rows)))
	}

	names := make([]string, 0, len(matchedNames))
	for n := range matchedNames {
		names = append(names, n)
	}

	logger.Info("MatchProducts complete", "matched_rows", len(rows), "matched_names", len(names))
	return &MatchProductsResult{Matched: len(rows), MatchedNames: names}, nil
}

// --- Activity 2: ClassifyOSCore ---

// ClassifyOSCoreParams holds input for the ClassifyOSCore activity.
type ClassifyOSCoreParams struct {
	RunTimestamp time.Time
	MatchedNames []string
}

// ClassifyOSCoreResult holds output from the ClassifyOSCore activity.
type ClassifyOSCoreResult struct {
	OSCore      int
	OSCoreNames []string
}

// ClassifyOSCore loads OS core references and rules, classifies unmatched apps,
// and writes os_core results to gold.lifecycle_software.
func (a *Activities) ClassifyOSCore(ctx context.Context, params ClassifyOSCoreParams) (*ClassifyOSCoreResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting ClassifyOSCore activity")

	// 1. Load OS core repo names (Ubuntu + RPM).
	osCoreNames, err := a.loadOSCoreNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("load os core names: %w", err)
	}
	logger.Info("Loaded OS core repo names", "count", len(osCoreNames))

	// 2. Load OS core rules from DB.
	osCoreExact, osCorePrefixes, osCoreSuffixes, err := a.loadOSCoreRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load os core rules: %w", err)
	}

	// 3. Load product mappings for guardedByEOL check.
	mappings, err := a.loadProductMappings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load product mappings: %w", err)
	}
	mappingBySlug := make(map[string]*productMapping, len(mappings))
	for i := range mappings {
		mappingBySlug[mappings[i].slug] = &mappings[i]
	}

	// 4. Load all apps and filter to unmatched.
	matchedSet := make(map[string]bool, len(params.MatchedNames))
	for _, n := range params.MatchedNames {
		matchedSet[n] = true
	}

	apps, err := a.loadInstalledApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("load installed apps: %w", err)
	}

	// 5. Classify by unique name.
	nameClassified := make(map[string]bool) // true = os_core, false = not
	for _, app := range apps {
		if matchedSet[app.Name] {
			continue
		}
		if _, done := nameClassified[app.Name]; done {
			continue
		}
		nameClassified[app.Name] = isOSCore(app.Name, osCoreNames, mappingBySlug, osCoreExact, osCorePrefixes, osCoreSuffixes)
	}

	// Build rows for all os_core apps.
	var rows []lifecycleRow
	for _, app := range apps {
		if matchedSet[app.Name] || !nameClassified[app.Name] {
			continue
		}
		rows = append(rows, lifecycleRow{
			machineID:      app.MachineID,
			name:           app.Name,
			version:        app.Version,
			classification: "os_core",
			eolStatus:      "unknown",
		})
	}

	for i := 0; i < len(rows); i += batchSize {
		end := min(i+batchSize, len(rows))
		if err := a.upsertLifecycleBatch(ctx, rows[i:end], params.RunTimestamp); err != nil {
			return nil, fmt.Errorf("upsert os_core batch: %w", err)
		}
		activity.RecordHeartbeat(ctx, fmt.Sprintf("os_core %d/%d", end, len(rows)))
	}

	var names []string
	for n, isCore := range nameClassified {
		if isCore {
			names = append(names, n)
		}
	}

	logger.Info("ClassifyOSCore complete", "os_core_rows", len(rows), "os_core_names", len(names))
	return &ClassifyOSCoreResult{OSCore: len(rows), OSCoreNames: names}, nil
}

// --- Activity 3: MarkUnmatched ---

// MarkUnmatchedParams holds input for the MarkUnmatched activity.
type MarkUnmatchedParams struct {
	RunTimestamp time.Time
	MatchedNames []string
	OSCoreNames  []string
}

// MarkUnmatchedResult holds output from the MarkUnmatched activity.
type MarkUnmatchedResult struct {
	Unmatched int
}

// MarkUnmatched writes remaining unclassified apps as unmatched.
func (a *Activities) MarkUnmatched(ctx context.Context, params MarkUnmatchedParams) (*MarkUnmatchedResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting MarkUnmatched activity")

	excludeNames := make(map[string]bool, len(params.MatchedNames)+len(params.OSCoreNames))
	for _, n := range params.MatchedNames {
		excludeNames[n] = true
	}
	for _, n := range params.OSCoreNames {
		excludeNames[n] = true
	}

	apps, err := a.loadInstalledApps(ctx)
	if err != nil {
		return nil, fmt.Errorf("load installed apps: %w", err)
	}

	var rows []lifecycleRow
	for _, app := range apps {
		if excludeNames[app.Name] {
			continue
		}
		rows = append(rows, lifecycleRow{
			machineID:      app.MachineID,
			name:           app.Name,
			version:        app.Version,
			classification: "unmatched",
			eolStatus:      "unknown",
		})
	}

	for i := 0; i < len(rows); i += batchSize {
		end := min(i+batchSize, len(rows))
		if err := a.upsertLifecycleBatch(ctx, rows[i:end], params.RunTimestamp); err != nil {
			return nil, fmt.Errorf("upsert unmatched batch: %w", err)
		}
		activity.RecordHeartbeat(ctx, fmt.Sprintf("unmatched %d/%d", end, len(rows)))
	}

	logger.Info("MarkUnmatched complete", "unmatched_rows", len(rows))
	return &MarkUnmatchedResult{Unmatched: len(rows)}, nil
}

// --- Activity 4: CleanupStale ---

// CleanupStaleParams holds input for the CleanupStale activity.
type CleanupStaleParams struct {
	RunTimestamp time.Time
}

// CleanupStaleResult holds output from the CleanupStale activity.
type CleanupStaleResult struct {
	Deleted int
}

// CleanupStale deletes gold.lifecycle_software rows not updated in this run.
func (a *Activities) CleanupStale(ctx context.Context, params CleanupStaleParams) (*CleanupStaleResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting CleanupStale activity")

	result, err := a.db.ExecContext(ctx,
		`DELETE FROM gold.lifecycle_software WHERE detected_at < $1`,
		params.RunTimestamp)
	if err != nil {
		return nil, fmt.Errorf("delete stale rows: %w", err)
	}

	deleted, _ := result.RowsAffected()
	logger.Info("CleanupStale complete", "deleted", deleted)
	return &CleanupStaleResult{Deleted: int(deleted)}, nil
}

// --- Data loading ---

// osCoreRPMRepos defines which RPM repos are considered OS core.
var osCoreRPMRepos = map[string]bool{
	"rhel9-baseos":    true,
	"rhel9-appstream": true,
	"rhel9-ha":        true,
	"rhel9-crb":       true,
	"rhel7-os":        true,
	"rhel7-updates":   true,
	"rhel7-sclo":      true,
	"rhel7-extras":    true,
	"epel9":           true,
	"epel7":           true,
}

func (a *Activities) loadProductMappings(ctx context.Context) ([]productMapping, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT i.product, i.identifier_type, i.value, p.category, p.name
		FROM bronze.reference_eol_identifiers i
		JOIN bronze.reference_eol_products p ON p.resource_id = i.product
		WHERE p.category NOT IN ('os', 'service', 'standard', 'device')
		  AND ((i.identifier_type = 'purl'
		        AND (i.value LIKE 'pkg:deb/%' OR i.value LIKE 'pkg:rpm/%'))
		       OR i.identifier_type = 'repology')`)
	if err != nil {
		return nil, fmt.Errorf("query identifiers: %w", err)
	}
	defer rows.Close()

	type productInfo struct {
		name        string
		eolCategory string
		prefixes    map[string]bool
		hasPackage  bool
	}
	products := make(map[string]*productInfo)

	for rows.Next() {
		var product, idType, value, eolCategory, name string
		if err := rows.Scan(&product, &idType, &value, &eolCategory, &name); err != nil {
			return nil, fmt.Errorf("scan identifier: %w", err)
		}
		pi, ok := products[product]
		if !ok {
			pi = &productInfo{name: name, eolCategory: eolCategory, prefixes: make(map[string]bool)}
			products[product] = pi
		}
		if idType == "purl" {
			if pkgName := parsePURLPackageName(value); pkgName != "" {
				pi.prefixes[pkgName] = true
				pi.hasPackage = true
			}
		} else if idType == "repology" {
			pi.prefixes["repology:"+value] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identifiers: %w", err)
	}

	// Resolve repology fallback.
	for _, pi := range products {
		var repologyNames []string
		for name := range pi.prefixes {
			if strings.HasPrefix(name, "repology:") {
				repologyNames = append(repologyNames, name)
			}
		}
		if pi.hasPackage {
			for _, name := range repologyNames {
				delete(pi.prefixes, name)
			}
		} else {
			for _, name := range repologyNames {
				delete(pi.prefixes, name)
				pi.prefixes[strings.TrimPrefix(name, "repology:")] = true
			}
		}
	}

	// Load match rules from DB.
	matchRules, err := a.loadMatchRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load match rules: %w", err)
	}

	// Ensure products with extra prefixes are loaded even without PURL/repology.
	for slug := range matchRules.extraPrefixes {
		if _, ok := products[slug]; !ok {
			products[slug] = &productInfo{prefixes: make(map[string]bool)}
		}
	}

	// Load product info for any products missing it.
	var missingSlugs []string
	for slug, pi := range products {
		if pi.name == "" {
			missingSlugs = append(missingSlugs, slug)
		}
	}
	if len(missingSlugs) > 0 {
		infoRows, err := a.db.QueryContext(ctx, `
			SELECT resource_id, name, category
			FROM bronze.reference_eol_products
			WHERE resource_id = ANY($1)`, missingSlugs)
		if err != nil {
			return nil, fmt.Errorf("query product info: %w", err)
		}
		defer infoRows.Close()
		for infoRows.Next() {
			var slug, name, category string
			if err := infoRows.Scan(&slug, &name, &category); err != nil {
				return nil, fmt.Errorf("scan product info: %w", err)
			}
			if pi, ok := products[slug]; ok {
				pi.name = name
				pi.eolCategory = category
			}
		}
		if err := infoRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate product info: %w", err)
		}
	}

	// Build mappings.
	var mappings []productMapping
	for slug, pi := range products {
		if len(pi.prefixes) == 0 && matchRules.extraPrefixes[slug] == nil {
			continue
		}
		prefixes := make([]string, 0, len(pi.prefixes))
		for p := range pi.prefixes {
			prefixes = append(prefixes, p)
		}
		sortStrings(prefixes)

		mappings = append(mappings, productMapping{
			slug:          slug,
			name:          pi.name,
			eolCategory:   pi.eolCategory,
			prefixes:      prefixes,
			extraPrefixes: matchRules.extraPrefixes[slug],
			excludes:      matchRules.excludes[slug],
			exactOnly:     pi.eolCategory == "lang",
			nameCycleMap:  matchRules.nameCycleMaps[slug],
		})
	}
	sortMappings(mappings)
	return mappings, nil
}

// matchRulesData holds parsed match rules from the DB.
type matchRulesData struct {
	extraPrefixes map[string][]string
	excludes      map[string][]string
	nameCycleMaps map[string]map[string]string
}

func (a *Activities) loadMatchRules(ctx context.Context) (*matchRulesData, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT product_slug, rule_type, os_type, value, extra_value
		FROM bronze.reference_software_match_rules`)
	if err != nil {
		return nil, fmt.Errorf("query match rules: %w", err)
	}
	defer rows.Close()

	data := &matchRulesData{
		extraPrefixes: make(map[string][]string),
		excludes:      make(map[string][]string),
		nameCycleMaps: make(map[string]map[string]string),
	}

	for rows.Next() {
		var slug, ruleType, value string
		var osType, extraValue sql.NullString
		if err := rows.Scan(&slug, &ruleType, &osType, &value, &extraValue); err != nil {
			return nil, fmt.Errorf("scan match rule: %w", err)
		}
		switch ruleType {
		case "extra_prefix":
			data.extraPrefixes[slug] = append(data.extraPrefixes[slug], value)
		case "exclude":
			data.excludes[slug] = append(data.excludes[slug], value)
		case "name_cycle_map":
			if data.nameCycleMaps[slug] == nil {
				data.nameCycleMaps[slug] = make(map[string]string)
			}
			if extraValue.Valid {
				data.nameCycleMaps[slug][value] = extraValue.String
			}
		}
	}
	return data, rows.Err()
}

func (a *Activities) loadEOLCycles(ctx context.Context, slugs []string) (map[string][]eolCycleInfo, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT product, cycle, eoas, eol, eoes, COALESCE(latest, '')
		FROM bronze.reference_eol_cycles
		WHERE product = ANY($1)
		ORDER BY product, cycle`, slugs)
	if err != nil {
		return nil, fmt.Errorf("query eol cycles: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]eolCycleInfo)
	for rows.Next() {
		var c eolCycleInfo
		var eoas, eol, eoes sql.NullTime
		if err := rows.Scan(&c.Product, &c.Cycle, &eoas, &eol, &eoes, &c.Latest); err != nil {
			return nil, fmt.Errorf("scan eol cycle: %w", err)
		}
		if eoas.Valid {
			c.EOAS = &eoas.Time
		}
		if eol.Valid {
			c.EOL = &eol.Time
		}
		if eoes.Valid {
			c.EOES = &eoes.Time
		}
		result[c.Product] = append(result[c.Product], c)
	}
	return result, rows.Err()
}

func (a *Activities) loadInstalledApps(ctx context.Context) ([]installedApp, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT s.machine_id, s.name, COALESCE(s.version, '')
		FROM inventory.software s
		WHERE s.version IS NOT NULL AND s.version != ''`)
	if err != nil {
		return nil, fmt.Errorf("query installed_software: %w", err)
	}
	defer rows.Close()

	var result []installedApp
	for rows.Next() {
		var app installedApp
		if err := rows.Scan(&app.MachineID, &app.Name, &app.Version); err != nil {
			return nil, fmt.Errorf("scan installed app: %w", err)
		}
		result = append(result, app)
	}
	return result, rows.Err()
}

func (a *Activities) loadOSCoreNames(ctx context.Context) (map[string]bool, error) {
	type nameInfo struct{ coreOnly, appLayer bool }
	names := make(map[string]*nameInfo)

	mark := func(name string, isCore bool) {
		ni, ok := names[name]
		if !ok {
			ni = &nameInfo{}
			names[name] = ni
		}
		if isCore {
			ni.coreOnly = true
		} else {
			ni.appLayer = true
		}
	}

	// Ubuntu packages.
	ubRows, err := a.db.QueryContext(ctx, `
		SELECT DISTINCT LOWER(package_name), component
		FROM bronze.reference_ubuntu_packages`)
	if err != nil {
		return nil, fmt.Errorf("query ubuntu packages: %w", err)
	}
	defer ubRows.Close()

	for ubRows.Next() {
		var name, component string
		if err := ubRows.Scan(&name, &component); err != nil {
			return nil, fmt.Errorf("scan ubuntu package: %w", err)
		}
		mark(name, component == "main" || component == "universe")
	}
	if err := ubRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ubuntu packages: %w", err)
	}

	// RPM packages.
	rpmRows, err := a.db.QueryContext(ctx, `
		SELECT DISTINCT LOWER(package_name), repo
		FROM bronze.reference_rpm_packages`)
	if err != nil {
		return nil, fmt.Errorf("query rpm packages: %w", err)
	}
	defer rpmRows.Close()

	for rpmRows.Next() {
		var name, repo string
		if err := rpmRows.Scan(&name, &repo); err != nil {
			return nil, fmt.Errorf("scan rpm package: %w", err)
		}
		mark(name, osCoreRPMRepos[repo])
	}
	if err := rpmRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rpm packages: %w", err)
	}

	result := make(map[string]bool)
	for name, ni := range names {
		if ni.coreOnly {
			result[name] = true
			if base := stripVersionSuffix(name); base != name {
				result[base] = true
			}
		}
	}
	return result, nil
}

func (a *Activities) loadOSCoreRules(ctx context.Context) (exact map[string]bool, prefixes []string, suffixes []string, err error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT rule_type, os_type, value
		FROM bronze.reference_os_core_rules`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query os core rules: %w", err)
	}
	defer rows.Close()

	exact = make(map[string]bool)
	for rows.Next() {
		var ruleType, value string
		var osType sql.NullString
		if err := rows.Scan(&ruleType, &osType, &value); err != nil {
			return nil, nil, nil, fmt.Errorf("scan os core rule: %w", err)
		}
		switch ruleType {
		case "exact":
			exact[value] = true
		case "prefix":
			prefixes = append(prefixes, value)
		case "suffix":
			suffixes = append(suffixes, value)
		}
	}
	return exact, prefixes, suffixes, rows.Err()
}

// --- Bulk upsert ---

func (a *Activities) upsertLifecycleBatch(ctx context.Context, rows []lifecycleRow, runTimestamp time.Time) error {
	if len(rows) == 0 {
		return nil
	}

	const cols = 16
	var b strings.Builder
	b.WriteString(`INSERT INTO gold.lifecycle_software
		(resource_id, detected_at, first_detected_at, machine_id, name, version,
		 classification, eol_product_slug, eol_product_name, eol_category,
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

		resourceID := r.machineID + ":" + r.name
		args = append(args, resourceID, runTimestamp, runTimestamp,
			r.machineID, r.name, nilIfEmpty(r.version),
			r.classification, r.eolProductSlug, r.eolProductName, r.eolCategory,
			r.eolCycle, r.eolDate, r.eoasDate, r.eoesDate, r.eolStatus, r.latestVersion)
	}

	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		detected_at = EXCLUDED.detected_at,
		machine_id = EXCLUDED.machine_id,
		name = EXCLUDED.name,
		version = EXCLUDED.version,
		classification = EXCLUDED.classification,
		eol_product_slug = EXCLUDED.eol_product_slug,
		eol_product_name = EXCLUDED.eol_product_name,
		eol_category = EXCLUDED.eol_category,
		eol_cycle = EXCLUDED.eol_cycle,
		eol_date = EXCLUDED.eol_date,
		eoas_date = EXCLUDED.eoas_date,
		eoes_date = EXCLUDED.eoes_date,
		eol_status = EXCLUDED.eol_status,
		latest_version = EXCLUDED.latest_version`)

	_, err := a.db.ExecContext(ctx, b.String(), args...)
	return err
}

// --- Helpers ---

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
