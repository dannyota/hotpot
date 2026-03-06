// debug-software is a thin SQL-driven tool for software lifecycle analysis.
// Flow: load installed apps → match EOL → filter OS core → show matched + unmatched.
// For ad-hoc queries, use psql and curl directly.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"danny.vn/hotpot/pkg/base/app"
)

func main() {
	ctx := context.Background()

	application, err := app.New(app.Options{})
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	if err := application.Start(ctx); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	defer application.Stop()

	dsn := application.ConfigService().DatabaseDSN()
	if dsn == "" {
		log.Fatal("Database not configured")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "lifecycle":
		runLifecycle(ctx, db, os.Args[2:])
	case "unmatched":
		runUnmatched(ctx, db, os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: debug-software <subcommand> [flags]

Subcommands:
  lifecycle   EOL lifecycle: match installed apps against endoflife.date
  unmatched   Show installed apps not matched by any EOL source`)
}

// --- lifecycle subcommand ---

func runLifecycle(ctx context.Context, db *sql.DB, args []string) {
	fs := flag.NewFlagSet("lifecycle", flag.ExitOnError)
	productFilter := fs.String("product", "", "Filter by endoflife.date product slug")
	categoryFilter := fs.String("category", "", "Filter by EOL category (e.g. database, lang, framework)")
	osFilter := fs.String("os", "", "Filter by OS type: linux, windows, macos")
	envFilter := fs.String("env", "", "Filter by environment (e.g., PRODUCTION)")
	projectFilter := fs.String("project", "", "Filter by GCP cloud_project (comma-separated, substring match)")
	showMachines := fs.Bool("machines", false, "Show per-machine details for each cycle")
	showAll := fs.Bool("all", false, "Include cycles without EOL reference data")
	fs.Parse(args)

	now := time.Now()

	// 1. Build product mappings from endoflife.date identifiers.
	log.Println("Loading product mappings from endoflife.date...")
	allMappings, err := loadProductMappings(ctx, db, *osFilter)
	if err != nil {
		log.Fatalf("load product mappings: %v", err)
	}

	// Filter mappings.
	var mappings []productMapping
	for _, m := range allMappings {
		if *productFilter != "" && m.slug != *productFilter {
			continue
		}
		if *categoryFilter != "" && m.eolCategory != *categoryFilter {
			continue
		}
		mappings = append(mappings, m)
	}
	if len(mappings) == 0 {
		log.Fatal("No product mappings match the given filters")
	}

	slugs := make([]string, len(mappings))
	mappingBySlug := make(map[string]*productMapping, len(mappings))
	for i := range mappings {
		slugs[i] = mappings[i].slug
		mappingBySlug[mappings[i].slug] = &mappings[i]
	}

	// 2. Load EOL cycles.
	log.Println("Loading EOL cycles...")
	cycles, err := loadEOLCycles(ctx, db, slugs)
	if err != nil {
		log.Fatalf("load EOL cycles: %v", err)
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
	log.Println("Loading installed apps...")
	projects := parseCSV(*projectFilter)
	apps, err := loadInstalledApps(ctx, db, *osFilter, *envFilter, projects)
	if err != nil {
		log.Fatalf("load installed apps: %v", err)
	}
	logAppStats("Installed", apps)

	// 4. Match apps → EOL products → cycles.
	log.Println("Matching apps to EOL products...")
	type cycleGroup struct {
		Product  string
		Slug     string
		Category string
		Cycle    string
		EOLCycle *eolCycle
		Machines map[string]bool
		Apps     map[string]map[string][]string // pkg → version → machineIDs
	}

	groups := make(map[string]*cycleGroup)
	matchedNames := make(map[string]bool)
	for _, a := range apps {
		pm := matchAppToProduct(a.Name, mappings)
		if pm == nil {
			continue
		}
		matchedNames[a.Name] = true
		cycle := extractCycleFromMapping(a.Name, a.Version, pm, knownCycleSets[pm.slug])
		if cycle == "" {
			continue
		}

		key := pm.slug + ":" + cycle
		g, ok := groups[key]
		if !ok {
			displayName := pm.name
			if displayName == "" {
				displayName = pm.slug
			}
			var matched *eolCycle
			if slugCycles, ok := cycles[pm.slug]; ok {
				for i := range slugCycles {
					if slugCycles[i].Cycle == cycle {
						matched = &slugCycles[i]
						break
					}
				}
			}
			g = &cycleGroup{
				Product:  displayName,
				Slug:     pm.slug,
				Category: pm.eolCategory,
				Cycle:    cycle,
				EOLCycle: matched,
				Machines: make(map[string]bool),
				Apps:     make(map[string]map[string][]string),
			}
			groups[key] = g
		}
		g.Machines[a.MachineID] = true
		if g.Apps[a.Name] == nil {
			g.Apps[a.Name] = make(map[string][]string)
		}
		g.Apps[a.Name][a.Version] = append(g.Apps[a.Name][a.Version], a.MachineID)
	}

	// Count unmatched (excluding OS core).
	log.Println("Loading OS core references for unmatched count...")
	osCoreNames, err := loadOSCoreNames(ctx, db)
	if err != nil {
		log.Fatalf("load OS core names: %v", err)
	}
	osCoreExact := resolveOSCoreExact(*osFilter)
	osCorePrefixes := resolveOSCorePrefixes(*osFilter)
	unmatchedNames := make(map[string]bool)
	osCoreSkipped := make(map[string]bool)
	for _, a := range apps {
		if matchedNames[a.Name] {
			continue
		}
		if isOSCore(a.Name, osCoreNames, mappingBySlug, osCoreExact, osCorePrefixes) {
			osCoreSkipped[a.Name] = true
			continue
		}
		unmatchedNames[a.Name] = true
	}
	log.Printf("  EOL matched: %d names, OS core: %d names, unmatched: %d names",
		len(matchedNames), len(osCoreSkipped), len(unmatchedNames))

	// 7. Display lifecycle report.
	sorted := make([]*cycleGroup, 0, len(groups))
	var hiddenCount int
	for _, g := range groups {
		if !*showAll && g.EOLCycle == nil {
			hiddenCount++
			continue
		}
		sorted = append(sorted, g)
	}
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && (sorted[j].Product < sorted[j-1].Product ||
			(sorted[j].Product == sorted[j-1].Product && sorted[j].Cycle < sorted[j-1].Cycle)); j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	productSet := make(map[string]bool)
	for _, g := range sorted {
		productSet[g.Slug] = true
	}

	printSection(fmt.Sprintf("EOL Lifecycle — %d products, %d cycles (endoflife.date)", len(productSet), len(sorted)))
	headers := []string{"Product", "EOL Cat", "Cycle", "EOAS", "EOL", "Latest", "Machines"}
	var rows [][]string
	for _, g := range sorted {
		eoas := "N/A"
		eol := "N/A"
		latest := ""
		if g.EOLCycle != nil {
			eoas = formatTimeRemaining(g.EOLCycle.EOAS, now)
			eol = formatTimeRemaining(g.EOLCycle.EOL, now)
			latest = g.EOLCycle.Latest
		}
		rows = append(rows, []string{
			g.Product, g.Category, g.Cycle, eoas, eol, latest,
			fmt.Sprintf("%d", len(g.Machines)),
		})
	}
	printTable(headers, rows)
	if hiddenCount > 0 {
		fmt.Printf("\n  (%d cycles without EOL reference data hidden; use --all to show)\n", hiddenCount)
	}

	fmt.Printf("\n  Summary: %d matched names, %d OS core, %d unmatched (use 'unmatched' subcommand)\n\n",
		len(matchedNames), len(osCoreSkipped), len(unmatchedNames))

	// 9. Machine sub-tables.
	if !*showMachines {
		return
	}

	allMachineIDs := make(map[string]bool)
	for _, g := range sorted {
		for id := range g.Machines {
			allMachineIDs[id] = true
		}
	}
	ids := make([]string, 0, len(allMachineIDs))
	for id := range allMachineIDs {
		ids = append(ids, id)
	}

	log.Println("Loading machine details...")
	machineMap, err := loadLifecycleMachines(ctx, db, ids)
	if err != nil {
		log.Fatalf("load lifecycle machines: %v", err)
	}

	for _, g := range sorted {
		fmt.Printf("\n--- %s %s (%d machines) ---\n\n", g.Product, g.Cycle, len(g.Machines))
		mHeaders := []string{"Hostname", "Env", "OS", "Status", "Package", "Version"}
		var mRows [][]string
		for pkgName, versions := range g.Apps {
			for version, machineIDs := range versions {
				for _, machineID := range machineIDs {
					m, ok := machineMap[machineID]
					if !ok {
						m = lifecycleMachine{Hostname: machineID[:min(16, len(machineID))] + "…", Status: "unknown"}
					}
					mRows = append(mRows, []string{
						truncate(m.Hostname, 24), truncate(m.Environment, 12),
						truncate(m.OSName, 16), m.Status,
						truncate(pkgName, 30), truncate(version, 20),
					})
				}
			}
		}
		sortStringSlices(mRows)
		printTable(mHeaders, mRows)
		fmt.Println()
	}
}

// --- unmatched subcommand ---

func runUnmatched(ctx context.Context, db *sql.DB, args []string) {
	fs := flag.NewFlagSet("unmatched", flag.ExitOnError)
	osFilter := fs.String("os", "", "Filter by OS type: linux, windows, macos")
	envFilter := fs.String("env", "", "Filter by environment (e.g., PRODUCTION)")
	projectFilter := fs.String("project", "", "Filter by GCP cloud_project (comma-separated)")
	limit := fs.Int("limit", 0, "Max unmatched names to show (0 = all)")
	fs.Parse(args)

	// 1. Load mappings + apps.
	log.Println("Loading product mappings...")
	allMappings, err := loadProductMappings(ctx, db, *osFilter)
	if err != nil {
		log.Fatalf("load product mappings: %v", err)
	}

	log.Println("Loading installed apps...")
	projects := parseCSV(*projectFilter)
	apps, err := loadInstalledApps(ctx, db, *osFilter, *envFilter, projects)
	if err != nil {
		log.Fatalf("load installed apps: %v", err)
	}
	logAppStats("Installed", apps)

	// 2. Find matched names.
	matchedNames := make(map[string]bool)
	for _, a := range apps {
		if matchAppToProduct(a.Name, allMappings) != nil {
			matchedNames[a.Name] = true
		}
	}

	// 3. Load OS core names — exclude from unmatched display.
	log.Println("Loading OS core references...")
	osCoreNames, err := loadOSCoreNames(ctx, db)
	if err != nil {
		log.Fatalf("load OS core names: %v", err)
	}
	mappingBySlug := make(map[string]*productMapping, len(allMappings))
	for i := range allMappings {
		mappingBySlug[allMappings[i].slug] = &allMappings[i]
	}

	// 4. Count machines per unmatched name (excluding EOL-matched and OS core).
	type unmatchedInfo struct {
		name     string
		machines map[string]bool
		versions map[string]bool
		sampleVer string
	}
	osCoreExact := resolveOSCoreExact(*osFilter)
	osCorePrefixes := resolveOSCorePrefixes(*osFilter)
	unmatched := make(map[string]*unmatchedInfo)
	var osCoreCount int
	for _, a := range apps {
		if matchedNames[a.Name] {
			continue
		}
		if isOSCore(a.Name, osCoreNames, mappingBySlug, osCoreExact, osCorePrefixes) {
			osCoreCount++
			continue
		}
		ui, ok := unmatched[a.Name]
		if !ok {
			ui = &unmatchedInfo{name: a.Name, machines: make(map[string]bool), versions: make(map[string]bool)}
			unmatched[a.Name] = ui
		}
		ui.machines[a.MachineID] = true
		ui.versions[a.Version] = true
		ui.sampleVer = a.Version
	}
	log.Printf("  EOL matched: %d names, OS core skipped: %d records, unmatched: %d names",
		len(matchedNames), osCoreCount, len(unmatched))

	// Sort by machine count descending.
	sorted := make([]*unmatchedInfo, 0, len(unmatched))
	for _, ui := range unmatched {
		sorted = append(sorted, ui)
	}
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && len(sorted[j].machines) > len(sorted[j-1].machines); j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	if *limit > 0 && len(sorted) > *limit {
		sorted = sorted[:*limit]
	}

	printSection(fmt.Sprintf("Unmatched Apps — %d names (sorted by install count)", len(unmatched)))
	headers := []string{"Name", "Machines", "Versions", "Sample Version"}
	var rows [][]string
	for _, ui := range sorted {
		rows = append(rows, []string{
			truncate(ui.name, 40),
			fmt.Sprintf("%d", len(ui.machines)),
			fmt.Sprintf("%d", len(ui.versions)),
			truncate(ui.sampleVer, 25),
		})
	}
	printTable(headers, rows)
	fmt.Printf("\n  Total: %d matched names, %d unmatched names\n\n", len(matchedNames), len(unmatched))
}

// --- Product mapping (endoflife.date PURL/repology identifiers) ---

type productMapping struct {
	slug          string
	name          string
	eolCategory   string
	prefixes      []string // PURL/repology-derived — subject to exactOnly
	extraPrefixes []string // manually added — always prefix-matched
	excludes      []string
	exactOnly     bool
	nameCycleMap  map[string]string // extract cycle from app name instead of version (e.g., "2014" → "12.0")
}

var productExcludes = map[string][]string{
	"ansible":       {"core", "test", "collection", "freeipa", "pcp"},
	"gstreamer":     {"packagekit", "nice", "clutter", "pipewire", "libcamera", "fdkaac"},
	"mongodb":       {"mongosh", "database", "compass"},
	"mysql":         {"connector", "router", "shell"},
	"mariadb":       {"connector"},
	"kubernetes":    {"cni", "csi"},
	"oracle-jdk":    {"openjdk", "common", "wrappers"},
	"visual-studio": {"code"},
	"mssqlserver":   {"management studio"},
	"unity":         {"hub"},
}

// productNameCycleMaps extract cycle from the app name instead of version.
// Used for products where the marketing version (e.g., "2014") differs from
// the internal version used by endoflife.date (e.g., "12.0").
var productNameCycleMaps = map[string]map[string]string{
	"mssqlserver": {
		"2008 r2": "10.50",
		"2008":    "10.0",
		"2012":    "11.0",
		"2014":    "12.0",
		"2016":    "13.0",
		"2017":    "14.0",
		"2019":    "15.0",
		"2022":    "16.0",
		"2025":    "17.0",
	},
}

// extraPrefixesAll apply regardless of OS filter.
var extraPrefixesAll = map[string][]string{
	"chrome":  {"google chrome"},
	"firefox": {"firefox"},
}

// extraPrefixesByOS are OS-specific extra prefixes.
// Key format: "os:slug" where os is "linux", "windows", or "macos".
var extraPrefixesByOS = map[string]map[string][]string{
	"linux": {
		"apache-http-server":       {"apache2", "httpd"},
		"docker-engine":            {"docker-ce", "docker"},
		"openssl":                  {"libssl"},
		"containerd":               {"containerd.io"},
		"kubernetes":               {"kubectl"},
		"red-hat-build-of-openjdk": {"jdk"},
		"azul-zulu":                {"zulu"},
		"firefox":                  {"mozilla firefox", "mozilla maintenance service"},
		"oracle-jdk":               {"java"},
		"nodejs":                   {"node.js"},
		"postgresql":               {"postgresql"},
		"openvpn":                  {"openvpn"},
		"unity":                    {"unity"},
	},
	"windows": {
		"chrome":            {"google chrome"},
		"firefox":           {"mozilla firefox", "mozilla maintenance service"},
		"office":            {"microsoft office", "microsoft 365 apps", "microsoft word", "microsoft excel", "microsoft powerpoint", "microsoft outlook", "microsoft onenote"},
		"visual-studio":     {"visual studio", "microsoft visual studio"},
		"dotnet":            {"microsoft .net runtime", "microsoft windows desktop runtime", "microsoft asp.net core"},
		"dotnetfx":          {"microsoft .net framework"},
		"oracle-jdk":        {"java"},
		"notepad-plus-plus": {"notepad++"},
		"nodejs":            {"node.js"},
		"mssqlserver":       {"microsoft sql server", "sql server browser for sql server"},
	},
	"macos": {
		"firefox":       {"firefox"},
		"office":        {"microsoft word", "microsoft excel", "microsoft powerpoint", "microsoft outlook", "microsoft onenote"},
		"visual-studio": {"visual studio"},
		"openvpn":       {"openvpn"},
		"unity":         {"unity"},
	},
}

// resolveExtraPrefixes merges the all-OS and OS-specific extra prefixes.
func resolveExtraPrefixes(osFilter string) map[string][]string {
	result := make(map[string][]string)
	for slug, prefixes := range extraPrefixesAll {
		result[slug] = append(result[slug], prefixes...)
	}
	if osFilter != "" {
		if osMap, ok := extraPrefixesByOS[osFilter]; ok {
			for slug, prefixes := range osMap {
				result[slug] = append(result[slug], prefixes...)
			}
		}
	} else {
		// No OS filter: merge all OS-specific prefixes.
		for _, osMap := range extraPrefixesByOS {
			for slug, prefixes := range osMap {
				result[slug] = append(result[slug], prefixes...)
			}
		}
	}
	return result
}

func loadProductMappings(ctx context.Context, db *sql.DB, osFilter string) ([]productMapping, error) {
	rows, err := db.QueryContext(ctx, `
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

	// Resolve OS-specific extra prefixes.
	resolved := resolveExtraPrefixes(osFilter)

	// Ensure products with extraPrefixes are loaded even without PURL/repology.
	for slug := range resolved {
		if _, ok := products[slug]; !ok {
			products[slug] = &productInfo{prefixes: make(map[string]bool)}
		}
	}

	// Load product info (name, category) for any products missing it.
	var missingSlugs []string
	for slug, pi := range products {
		if pi.name == "" {
			missingSlugs = append(missingSlugs, slug)
		}
	}
	if len(missingSlugs) > 0 {
		infoRows, err := db.QueryContext(ctx, `
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
		if len(pi.prefixes) == 0 && resolved[slug] == nil {
			continue
		}
		prefixes := make([]string, 0, len(pi.prefixes))
		for p := range pi.prefixes {
			prefixes = append(prefixes, p)
		}
		sortStrings(prefixes)

		exactOnly := pi.eolCategory == "lang"
		var excludes []string
		if extra, ok := productExcludes[slug]; ok {
			excludes = append(excludes, extra...)
		}

		var extras []string
		if ep, ok := resolved[slug]; ok {
			extras = append(extras, ep...)
		}

		mappings = append(mappings, productMapping{
			slug:          slug,
			name:          pi.name,
			eolCategory:   pi.eolCategory,
			prefixes:      prefixes,
			extraPrefixes: extras,
			excludes:      excludes,
			exactOnly:     exactOnly,
			nameCycleMap:  productNameCycleMaps[slug],
		})
	}
	sortMappings(mappings)
	log.Printf("  Loaded %d product mappings (%d EOL products with identifiers)", len(mappings), len(products))
	return mappings, nil
}

// --- OS core filtering ---

// All of Ubuntu main + universe = OS core. These packages ship with the OS
// and their EOL is tied to the Ubuntu release, not tracked independently.

// All RPM repos = OS core. These packages ship with the OS/EPEL
// and their EOL is tied to the RHEL/CentOS release.
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

func loadOSCoreNames(ctx context.Context, db *sql.DB) (map[string]bool, error) {
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
	ubRows, err := db.QueryContext(ctx, `
		SELECT DISTINCT LOWER(package_name), component
		FROM bronze.reference_ubuntu_packages`)
	if err != nil {
		return nil, fmt.Errorf("query ubuntu packages: %w", err)
	}
	defer ubRows.Close()

	var ubTotal, ubCore int
	for ubRows.Next() {
		var name, component string
		if err := ubRows.Scan(&name, &component); err != nil {
			return nil, fmt.Errorf("scan ubuntu package: %w", err)
		}
		ubTotal++
		isCore := component == "main" || component == "universe"
		if isCore {
			ubCore++
		}
		mark(name, isCore)
	}
	if err := ubRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ubuntu packages: %w", err)
	}

	// RPM packages.
	rpmRows, err := db.QueryContext(ctx, `
		SELECT DISTINCT LOWER(package_name), repo
		FROM bronze.reference_rpm_packages`)
	if err != nil {
		return nil, fmt.Errorf("query rpm packages: %w", err)
	}
	defer rpmRows.Close()

	var rpmTotal, rpmCore int
	for rpmRows.Next() {
		var name, repo string
		if err := rpmRows.Scan(&name, &repo); err != nil {
			return nil, fmt.Errorf("scan rpm package: %w", err)
		}
		rpmTotal++
		isCore := osCoreRPMRepos[repo]
		if isCore {
			rpmCore++
		}
		mark(name, isCore)
	}
	if err := rpmRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rpm packages: %w", err)
	}

	// A name is OS core if it appears in ANY core repo. These get their EOL
	// from the OS release itself.
	result := make(map[string]bool)
	for name, ni := range names {
		if ni.coreOnly {
			result[name] = true
			// Also store the version-stripped prefix so that
			// "linux-headers-6.14.0-37-generic" matches reference
			// "linux-headers-6.8.0-1003-gke" via prefix "linux-headers".
			if base := stripVersionSuffix(name); base != name {
				result[base] = true
			}
		}
	}

	log.Printf("  Ubuntu: %d entries (%d core), RPM: %d entries (%d core)", ubTotal, ubCore, rpmTotal, rpmCore)
	log.Printf("  Core-only names: %d (out of %d total reference names)", len(result), len(names))
	return result, nil
}

// osCorePrefixesAll are prefixes that are OS core regardless of OS type.
var osCorePrefixesAll = []string{
	"google-cloud-cli",   // gcloud CLI and plugins (rolling release)
	"google-cloud-sdk",   // older gcloud SDK naming
	"google-cloud-ops-",  // GCE monitoring/ops agents
	"google-cloud-sap-",  // GCE SAP agents
	"google-rhui-client", // RHEL repo config for GCE
	"gce-disk-expand",    // GCE disk auto-expand utility
	"gcsfuse",            // GCS FUSE mount utility
}

// osCorePrefixesByOS are prefixes that are OS core for a specific OS type.
var osCorePrefixesByOS = map[string][]string{
	"linux": {
		"linux-",    // kernel packages (HWE, modules, headers, images)
		"redhat-",   // RHEL meta packages
		"gpg-pubkey", // RPM signing keys
		"oem-",      // Ubuntu OEM hardware enablement meta packages
		"lib",       // shared libraries (libllvm, libicu, libprotobuf, etc.)
		"mesa-lib",  // Mesa graphics libraries (mesa-libgallium)
	},
	"windows": {
		// VC++ redistributables (runtime deps, not standalone software).
		"microsoft visual c++",
		// Hardware drivers — Intel.
		"intel(",        // intel(r) management engine, serial io, graphics, etc.
		"intel®",        // intel® driver & support assistant, proset, etc.
		"oneapi ",       // Intel OneAPI runtime
		"realtek ",      // Realtek audio, ethernet, wireless, card reader
		"displaylink ",  // DisplayLink graphics driver
		"thunderbolt",   // Thunderbolt software
		// Hardware drivers — peripherals & audio.
		"dolby ",           // Dolby audio/video drivers and settings
		"synaptics ",       // Synaptics touchpad/trackpoint/fingerprint
		"elan ",            // ELAN touchpad/trackpoint drivers
		"logi ",            // Logitech drivers (options+, bolt, plugin service)
		"logitech ",        // Logitech (older naming)
		"razer ",           // Razer gaming peripherals
		"steelseries ",     // SteelSeries gaming peripherals
		"corsair ",         // Corsair gaming peripherals
		"jabra ",           // Jabra audio devices
		"brother ",         // Brother printer drivers
		"fuji",             // Fuji Xerox / Fujifilm drivers (no space: fuji xerox, fujifilm)
		"smart noise ",     // Smart noise cancellation (Lenovo)
		"smart microphone ", // Smart microphone settings (Lenovo)
		"myasus",           // ASUS MyASUS utility
		// OEM vendor tools.
		"lenovo ",    // Lenovo vantage, system update, service bridge
		"dell ",      // Dell supportassist, command update, optimizer
		"hp ",        // HP support solutions, notifications, wolf security
		"thinkpad ",  // ThinkPad utilities
		// OS built-in components.
		"windows pc health",        // Windows PC Health Check
		"windows 11 ",              // Windows 11 editions and tools
		"windows 10 ",              // Windows 10 editions and tools
		"windows subsystem for",    // WSL
		"windows sdk",              // Windows SDK
		"windows software dev",     // Windows SDK variants
		"windows driver package",   // Windows driver packages
		"windows security",         // Windows Security app
		"windows widgets",          // Windows Widgets
		"microsoft update health",  // Microsoft Update Health Tools
		"microsoft search in bing", // Bing search add-on
		"microsoft store",          // Microsoft Store and related
		"microsoft edge game",      // Microsoft Edge Game Assist
		"update for windows",       // Windows cumulative updates
		// UWP built-in app prefixes.
		"xbox",          // Xbox Live, Game Bar, identity provider, etc.
		"solitaire",     // Solitaire & Casual Games, Solitaire Collection
		"calendar, mail", // UWP Mail & Calendar app
		"global.",       // Internal widget framework components
		// Microsoft bundled apps.
		"microsoft edge",      // Edge browser (bundled with Windows)
		"microsoft onedrive",  // OneDrive (bundled with Windows)
		"microsoft teams",     // Teams (bundled with Windows 11)
		"teams machine-wide",  // Teams installer component
		// Microsoft runtime/SDK dependencies.
		"microsoft odbc driver",    // SQL Server ODBC client drivers
		"microsoft ole db driver",  // SQL Server OLE DB client drivers
		"microsoft system clr",     // SQL Server CLR types
		"microsoft vss writer",     // SQL Server VSS backup writers
		"microsoft report viewer",  // Reporting runtime
		"microsoft help viewer",    // VS help component
		"microsoft access database", // Office database engine runtime
		"microsoft asp.net",        // ASP.NET MVC/web runtimes
		"microsoft silverlight",    // Deprecated runtime
		"microsoft web deploy",     // IIS deployment tool
		"microsoft mpi",            // HPC runtime
		"microsoft primary interop", // COM interop assemblies
		"microsoft edge webview",   // Edge WebView2 runtime
		"microsoft .net core sdk",  // .NET Core SDK (old)
		"microsoft .net core runtime", // .NET Core runtime (old)
		"microsoft .net compact",   // Ancient .NET CF
		"microsoft .net sdk",       // .NET SDK
		"microsoft azure",          // Azure SDK/tools/emulator
		"microsoft gameinput",      // Gaming input component
		"microsoft bitlocker",      // BitLocker admin tools
		"power automate",           // Power Automate for desktop
		"vs_",                      // VS internal components (vs_coreeditorfonts)
	},
	"macos": {}, // Apple built-in apps handled via osCoreExact
}

// osCoreExactAll are exact names filtered as OS core regardless of OS type.
var osCoreExactAll = map[string]bool{
	"gmail":         true, // Chrome PWA shortcut
	"docs":          true,
	"youtube":       true,
	"slides":        true,
	"sheets":        true,
	"outlook (pwa)": true,
}

// osCoreExactByOS are exact names filtered as OS core for a specific OS type.
var osCoreExactByOS = map[string]map[string]bool{
	"linux": {},
	"windows": {
		// Built-in UWP/Store apps.
		"weather":              true,
		"maps":                 true,
		"camera":               true,
		"calculator":           true,
		"photos":               true,
		"clock":                true,
		"news":                 true,
		"tips":                 true,
		"sticky notes":         true,
		"people":               true,
		"cortana":              true,
		"sound recorder":       true,
		"voice recorder":       true,
		"media player":         true,
		"groove music":         true,
		"movies & tv":          true,
		"paint 3d":             true,
		"print 3d":             true,
		"3d viewer":            true,
		"mixed reality portal": true,
		"feedback hub":         true,
		"app installer":        true,
		"settings":             true,
		"get help":             true,
		"store experience host": true,
		"snipping tool":        true,
		"alarms & clock":       true,
		"your phone":           true,
		"phone link":           true,
		"game speech window":   true,
		"game bar":             true,
		"microsoft to do":      true,
		"microsoft bing":       true,
		"microsoft news":       true,
		"operator messages":    true,
		"desktoppackagemetadata": true,
		// Built-in variants.
		"trackpoint":            true, // ThinkPad TrackPoint driver
		"onenote for windows 10": true, // UWP OneNote
		"microsoft 365 (office)": true, // UWP Office hub (not real Office)
		"supportassist":         true, // Dell SupportAssist (no "dell " prefix)
		"spotify widget":        true, // Spotify UWP widget
		"localservicecomponents": true, // Windows internal
		"python launcher":       true, // Python launcher (not Python itself)
		// PWA shortcuts (real Office is "microsoft word" etc.).
		"outlook":               true,
		"word":                  true,
		"excel":                 true,
		"powerpoint":            true,
		"office":                true,
		"microsoft lists":       true,
		"microsoft teams (pwa)": true,
		"whatsapp web":          true,
		"github":                true,
		"notebooklm":            true,
		"postman docs":          true,
		"google password manager": true,
	},
	"macos": {
		// Apple built-in apps — lifecycle tied to macOS release.
		"safari":           true,
		"mail":             true,
		"maps":             true,
		"photos":           true,
		"music":            true,
		"tv":               true,
		"news":             true,
		"stocks":           true,
		"notes":            true,
		"reminders":        true,
		"calendar":         true,
		"contacts":         true,
		"messages":         true,
		"facetime":         true,
		"freeform":         true,
		"weather":          true,
		"clock":            true,
		"calculator":       true,
		"passwords":        true,
		"shortcuts":        true,
		"preview":          true,
		"books":            true,
		"podcasts":         true,
		"home":             true,
		"findmy":           true,
		"find my":          true,
		"photo booth":      true,
		"voice memos":      true,
		"quicktime player": true,
		"terminal":         true,
		"console":          true,
		"activity monitor": true,
		"disk utility":     true,
		"migration assistant":    true,
		"system information":     true,
		"bluetooth file exchange": true,
		"font book":       true,
		"digital color meter": true,
		"grapher":         true,
		"screenshot":      true,
		"stickies":        true,
		"chess":           true,
		"textedit":        true,
		"image capture":   true,
		"automator":       true,
		"script editor":   true,
		"keychain access": true,
		"directory utility":  true,
		"system preferences": true,
		"system settings":    true,
		"app store":       true,
		"siri":            true,
		"time machine":    true,
		"iphone mirroring": true,
		"keynote":         true,
		"numbers":         true,
		"pages":           true,
		"garageband":      true,
		"imovie":          true,
		"xcode":           true,
		"instruments":     true,
		"filmerge":        true,
		"accessibility inspector": true,
		// Apple built-in utilities and system tools.
		"dictionary":        true,
		"voiceover utility": true,
		"colorsync utility": true,
		"boot camp assistant": true,
		"mission control":   true,
		"airport utility":   true,
		"screen sharing":    true,
		"print center":      true,
		"tips":              true,
		"image playground":  true,
		"launchpad":         true,
		"magnifier":         true,
		"games":             true,
		"journal":           true,
		"phone":             true,
		"apps":              true,
		"icon composer":     true,
		"create ml":         true,
		"reality composer pro": true,
		"simulator":         true,
		"runner":            true,
		// macOS security and system internals.
		"xprotect":          true,
		"mrt":               true, // Malware Removal Tool
		// Apple mobile sync/device helpers.
		"applemobiledevicehelper": true,
		"applemobilesync":         true,
		"mobiledeviceupdater":     true,
		// Apple scripting templates.
		"cocoa-applescript applet":               true,
		"droplet with settable properties":       true,
		"recursive file processing droplet":      true,
		"recursive image file processing droplet": true,
		// Missed earlier / alternate spellings.
		"audio midi setup":     true,
		"filemerge":            true,
		"digital colour meter": true, // British spelling variant
		"print centre":         true, // British spelling variant
		"testflight":           true,
		"developer":            true,
	},
}

// osCoreSuffixesAll are suffixes that indicate repo infrastructure packages.
// Real packages with these suffixes (gnome-keyring, python3-keyring) are already
// caught by OS core reference data — only repo infra remnants remain unmatched.
var osCoreSuffixesAll = []string{
	"-keyring",       // repo signing keys (brave-keyring, synaptics-repository-keyring)
	"-repo",          // repo config packages (pgdg-redhat-repo)
	"-release-notes", // OS release notes packages
	"-release_notes", // RHEL release notes (uses underscore)
}

// resolveOSCorePrefixes returns the combined prefixes for the given OS filter.
func resolveOSCorePrefixes(osFilter string) []string {
	result := append([]string(nil), osCorePrefixesAll...)
	if osFilter != "" {
		result = append(result, osCorePrefixesByOS[osFilter]...)
	} else {
		for _, ps := range osCorePrefixesByOS {
			result = append(result, ps...)
		}
	}
	return result
}

// resolveOSCoreExact returns the combined exact names for the given OS filter.
func resolveOSCoreExact(osFilter string) map[string]bool {
	result := make(map[string]bool, len(osCoreExactAll))
	for k := range osCoreExactAll {
		result[k] = true
	}
	if osFilter != "" {
		for k := range osCoreExactByOS[osFilter] {
			result[k] = true
		}
	} else {
		for _, m := range osCoreExactByOS {
			for k := range m {
				result[k] = true
			}
		}
	}
	return result
}

// isOSCore checks if a name is OS core by exact match, version-stripped prefix,
// or known OS core patterns.
func isOSCore(name string, osCoreNames map[string]bool, eolSlugs map[string]*productMapping, osCoreExact map[string]bool, osCorePrefixes []string) bool {
	if osCoreNames[name] || osCoreExact[name] {
		return true
	}
	if base := stripVersionSuffix(name); base != name && osCoreNames[base] {
		return true
	}
	// Skip pattern-based classification if the name would match an EOL product
	// in step 1. For exactOnly products (e.g. libreoffice), only the exact slug
	// is protected — sub-packages like libreoffice-calc remain OS core.
	if guardedByEOL(name, eolSlugs) {
		return false
	}
	for _, p := range osCorePrefixes {
		if strings.HasPrefix(name, p) || name == p {
			return true
		}
	}
	for _, s := range osCoreSuffixesAll {
		if strings.HasSuffix(name, s) || strings.Contains(name, s+"-") {
			return true
		}
	}
	return false
}

// guardedByEOL returns true if the name could match an EOL product in step 1,
// meaning pattern-based OS core filters (prefix/suffix) should not apply.
// For exactOnly products, only the exact slug is guarded — sub-packages are not.
func guardedByEOL(name string, eolSlugs map[string]*productMapping) bool {
	if m, ok := eolSlugs[name]; ok && !m.exactOnly {
		return true
	}
	if base := strings.SplitN(name, "-", 2)[0]; base != name {
		if m, ok := eolSlugs[base]; ok && !m.exactOnly {
			return true
		}
	}
	return false
}

// stripVersionSuffix removes version numbers after a separator from package names.
// "linux-headers-6.14.0-37-generic" → "linux-headers"
// "libpython3.10-minimal" → "libpython"
// "nginx" → "nginx" (unchanged)
// "postgresql17-ee" → "postgresql17-ee" (unchanged — digit embedded in word)
func stripVersionSuffix(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] >= '0' && name[i] <= '9' {
			// Only strip when digit follows a separator (- or .)
			if i > 0 && (name[i-1] == '-' || name[i-1] == '.') {
				return strings.TrimRight(name[:i-1], "-.")
			}
		}
	}
	return name
}

// --- DB queries ---

type eolCycle struct {
	Product           string
	Cycle             string
	ReleaseDate       sql.NullTime
	EOAS              sql.NullTime
	EOL               sql.NullTime
	EOES              sql.NullTime
	Latest            string
	LatestReleaseDate sql.NullTime
	LTS               sql.NullTime
}

type installedApp struct {
	MachineID string
	Name      string
	Version   string
}

type lifecycleMachine struct {
	Hostname    string
	Environment string
	OSType      string
	OSName      string
	Status      string
	MachineID   string
}

func loadEOLCycles(ctx context.Context, db *sql.DB, slugs []string) (map[string][]eolCycle, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT product, cycle, release_date, eoas, eol, eoes,
		       COALESCE(latest, ''), latest_release_date, lts
		FROM bronze.reference_eol_cycles
		WHERE product = ANY($1)
		ORDER BY product, cycle`, slugs)
	if err != nil {
		return nil, fmt.Errorf("query reference_eol_cycles: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]eolCycle)
	for rows.Next() {
		var c eolCycle
		if err := rows.Scan(&c.Product, &c.Cycle, &c.ReleaseDate, &c.EOAS,
			&c.EOL, &c.EOES, &c.Latest, &c.LatestReleaseDate, &c.LTS); err != nil {
			return nil, fmt.Errorf("scan eol cycle: %w", err)
		}
		result[c.Product] = append(result[c.Product], c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eol cycles: %w", err)
	}
	total := 0
	for _, v := range result {
		total += len(v)
	}
	log.Printf("  Loaded %d cycles across %d products", total, len(result))
	return result, nil
}

func loadInstalledApps(ctx context.Context, db *sql.DB, osFilter, envFilter string, projects []string) ([]installedApp, error) {
	query := `
		SELECT s.machine_id, s.name, COALESCE(s.version, '')
		FROM inventory.software s
		JOIN inventory.machines m ON m.resource_id = s.machine_id
		WHERE s.version IS NOT NULL AND s.version != ''`

	argN := 1
	var args []any
	if osFilter != "" {
		query += fmt.Sprintf(` AND m.os_type = $%d`, argN)
		args = append(args, osFilter)
		argN++
	}
	if envFilter != "" {
		query += fmt.Sprintf(` AND m.environment = $%d`, argN)
		args = append(args, envFilter)
		argN++
	}
	for _, proj := range projects {
		query += fmt.Sprintf(` AND m.cloud_project ILIKE '%%' || $%d || '%%'`, argN)
		args = append(args, proj)
		argN++
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query installed_software: %w", err)
	}
	defer rows.Close()

	var result []installedApp
	for rows.Next() {
		var a installedApp
		if err := rows.Scan(&a.MachineID, &a.Name, &a.Version); err != nil {
			return nil, fmt.Errorf("scan installed app: %w", err)
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func loadLifecycleMachines(ctx context.Context, db *sql.DB, machineIDs []string) (map[string]lifecycleMachine, error) {
	if len(machineIDs) == 0 {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT m.resource_id, m.hostname, COALESCE(m.environment, ''), m.os_type,
		       COALESCE(m.os_name, ''), m.status
		FROM inventory.machines m
		WHERE m.resource_id = ANY($1)`, machineIDs)
	if err != nil {
		return nil, fmt.Errorf("query lifecycle machines: %w", err)
	}
	defer rows.Close()

	result := make(map[string]lifecycleMachine)
	for rows.Next() {
		var m lifecycleMachine
		if err := rows.Scan(&m.MachineID, &m.Hostname, &m.Environment, &m.OSType,
			&m.OSName, &m.Status); err != nil {
			return nil, fmt.Errorf("scan lifecycle machine: %w", err)
		}
		result[m.MachineID] = m
	}
	return result, rows.Err()
}

// --- Matching helpers ---

func matchAppToProduct(name string, mappings []productMapping) *productMapping {
	// Try original name.
	if m := matchName(name, mappings); m != nil {
		return m
	}
	// Fallback: normalize name (strip parenthesized suffixes, embedded versions).
	if norm := normalizeAppName(name); norm != name {
		return matchName(norm, mappings)
	}
	return nil
}

func matchName(name string, mappings []productMapping) *productMapping {
	for i := range mappings {
		for _, prefix := range mappings[i].prefixes {
			if mappings[i].exactOnly {
				if name == prefix {
					return &mappings[i]
				}
				continue
			}
			if matchWithExcludes(name, prefix, mappings[i].excludes) {
				return &mappings[i]
			}
		}
		for _, prefix := range mappings[i].extraPrefixes {
			if matchWithExcludes(name, prefix, mappings[i].excludes) {
				return &mappings[i]
			}
		}
	}
	return nil
}

func matchWithExcludes(name, prefix string, excludes []string) bool {
	if !matchPrefix(name, prefix) {
		return false
	}
	if len(excludes) > 0 && len(name) > len(prefix)+1 {
		suffix := name[len(prefix)+1:]
		for _, ex := range excludes {
			// Check each segment (split by - or space) for the exclude word.
			for _, seg := range splitSegments(suffix) {
				if seg == ex {
					return false
				}
			}
		}
	}
	return true
}

func splitSegments(s string) []string {
	f := func(c rune) bool { return c == '-' || c == ' ' }
	return strings.FieldsFunc(s, f)
}

// matchPrefix checks if name starts with prefix followed by a separator (- or space) or is exact.
func matchPrefix(name, prefix string) bool {
	if name == prefix {
		return true
	}
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	sep := name[len(prefix)]
	return sep == '-' || sep == ' '
}

// normalizeAppName strips platform suffixes and embedded version digits.
//
//	"mozilla firefox (x64 en-us)" → "mozilla firefox"
//	"postgresql17-ee-libs"        → "postgresql-ee-libs"
//	"openssl3-libs"               → "openssl-libs"
//	"nginx"                       → "nginx" (unchanged)
func normalizeAppName(name string) string {
	// Strip trailing parenthesized groups: (x64), (64-bit), (en-us), etc.
	for strings.HasSuffix(name, ")") {
		idx := strings.LastIndex(name, " (")
		if idx < 0 {
			break
		}
		name = strings.TrimSpace(name[:idx])
	}
	// Strip embedded version digits: letter immediately followed by digits,
	// where digits are followed by separator or end.
	// postgresql17-ee → postgresql-ee, openssl3-libs → openssl-libs
	for i := 1; i < len(name); i++ {
		if name[i] >= '0' && name[i] <= '9' && name[i-1] >= 'a' && name[i-1] <= 'z' {
			j := i
			for j < len(name) && ((name[j] >= '0' && name[j] <= '9') ||
				(name[j] == '.' && j+1 < len(name) && name[j+1] >= '0' && name[j+1] <= '9')) {
				j++
			}
			if j == len(name) || name[j] == '-' || name[j] == ' ' {
				name = name[:i] + name[j:]
				break
			}
		}
	}
	return strings.TrimRight(name, "- .")
}

// extractCycleFromMapping tries nameCycleMap first (for products like SQL Server
// where the cycle is in the app name, not the version), then falls back to version parsing.
func extractCycleFromMapping(appName, version string, pm *productMapping, knownCycles map[string]bool) string {
	if pm.nameCycleMap != nil {
		// Try longest match first (e.g., "2008 r2" before "2008").
		for keyword, cycle := range pm.nameCycleMap {
			if strings.Contains(appName, keyword) {
				// Check if a longer keyword also matches (e.g., "2008 r2" vs "2008").
				bestKeyword, bestCycle := keyword, cycle
				for k2, c2 := range pm.nameCycleMap {
					if len(k2) > len(bestKeyword) && strings.Contains(appName, k2) {
						bestKeyword, bestCycle = k2, c2
					}
				}
				return bestCycle
			}
		}
	}
	return extractCycleWithFallback(version, knownCycles)
}

func extractCycleWithFallback(version string, knownCycles map[string]bool) string {
	c1 := extractCycle(version, 1)
	if knownCycles[c1] {
		return c1
	}
	c2 := extractCycle(version, 2)
	if knownCycles[c2] {
		return c2
	}
	return c2
}

func extractCycle(version string, depth int) string {
	v := version
	if idx := strings.IndexByte(v, ':'); idx >= 0 {
		v = v[idx+1:]
	}
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		v = v[:idx]
	}
	if idx := strings.IndexByte(v, '+'); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) < depth {
		return v
	}
	return strings.Join(parts[:depth], ".")
}

func parsePURLPackageName(purl string) string {
	parts := strings.SplitN(purl, "/", 3)
	if len(parts) < 3 {
		return ""
	}
	name := parts[2]
	if idx := strings.IndexByte(name, '?'); idx >= 0 {
		name = name[:idx]
	}
	if strings.HasPrefix(name, "/") || strings.Contains(name, ".") {
		return ""
	}
	return name
}

// --- Formatting helpers ---

func formatTimeRemaining(t sql.NullTime, now time.Time) string {
	if !t.Valid {
		return "N/A"
	}
	d := t.Time.Sub(now)
	date := t.Time.Format("2006-01-02")
	if d <= 0 {
		return "EXPIRED"
	}
	days := int(d.Hours() / 24)
	months := days / 30
	if days < 90 {
		return fmt.Sprintf("%dd (%s)", days, date)
	}
	if months < 24 {
		return fmt.Sprintf("%dm [!] (%s)", months, date)
	}
	return fmt.Sprintf("%dm (%s)", months, date)
}

func printTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  "+strings.Join(headers, "\t"))
	seps := make([]string, len(headers))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(w, "  "+strings.Join(seps, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, "  "+strings.Join(row, "\t"))
	}
	w.Flush()
}

func printSection(title string) {
	fmt.Printf("\n=== %s ===\n\n", title)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func logAppStats(label string, apps []installedApp) {
	names := make(map[string]bool)
	nameVer := make(map[string]bool)
	for _, a := range apps {
		names[a.Name] = true
		nameVer[a.Name+"\x00"+a.Version] = true
	}
	log.Printf("  %s: %d records, %d name+version, %d names", label, len(apps), len(nameVer), len(names))
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

// --- Sort helpers ---

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func sortMappings(m []productMapping) {
	for i := 1; i < len(m); i++ {
		for j := i; j > 0 && m[j].slug < m[j-1].slug; j-- {
			m[j], m[j-1] = m[j-1], m[j]
		}
	}
}

func sortStringSlices(rows [][]string) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && (rows[j][0] < rows[j-1][0] ||
			(rows[j][0] == rows[j-1][0] && len(rows[j]) > 4 && len(rows[j-1]) > 4 && rows[j][4] < rows[j-1][4])); j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}
