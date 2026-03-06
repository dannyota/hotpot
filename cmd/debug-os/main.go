// debug-os is a thin SQL-driven tool for OS lifecycle analysis.
// Flow: load machines → parse os_name → match EOL cycles → show lifecycle report.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
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
	fmt.Fprintln(os.Stderr, `Usage: debug-os <subcommand> [flags]

Subcommands:
  lifecycle   OS EOL lifecycle: match machines against endoflife.date
  unmatched   Show machines whose OS could not be matched`)
}

// --- lifecycle subcommand ---

type cycleGroup struct {
	Product       string
	Slug          string
	Cycle         string
	EOLCycle      *eolCycle
	Machines      []machineRecord
	MinorVersions map[string]int // minor version → machine count
}

func runLifecycle(ctx context.Context, db *sql.DB, args []string) {
	fs := flag.NewFlagSet("lifecycle", flag.ExitOnError)
	osFilter := fs.String("os", "", "Filter by OS type: linux, windows, macos")
	envFilter := fs.String("env", "", "Filter by environment (e.g., PRODUCTION)")
	projectFilter := fs.String("project", "", "Filter by GCP cloud_project (comma-separated, substring match)")
	statusFilter := fs.String("status", "", "Filter by machine status (e.g., running, stopped)")
	showMachines := fs.Bool("machines", false, "Show per-machine details for each cycle")
	showAll := fs.Bool("all", false, "Include cycles without EOL reference data")
	showMinor := fs.Bool("minor", false, "Show minor version breakdown within each cycle")
	fs.Parse(args)

	now := time.Now()

	// 1. Load OS EOL products.
	log.Println("Loading OS EOL products...")
	osSlugs, err := loadOSSlugs(ctx, db)
	if err != nil {
		log.Fatalf("load OS slugs: %v", err)
	}

	// 2. Load EOL cycles for OS products.
	log.Println("Loading EOL cycles...")
	slugList := make([]string, 0, len(osSlugs))
	for slug := range osSlugs {
		slugList = append(slugList, slug)
	}
	cycles, err := loadEOLCycles(ctx, db, slugList)
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

	// Build Windows build-number → cycle map.
	winBuildToCycle := buildWindowsBuildMap(cycles["windows"])

	// 3. Load machines.
	log.Println("Loading machines...")
	projects := parseCSV(*projectFilter)
	machines, err := loadMachines(ctx, db, *osFilter, *envFilter, *statusFilter, projects)
	if err != nil {
		log.Fatalf("load machines: %v", err)
	}
	log.Printf("  Loaded %d machines", len(machines))

	// 4. Match machines → OS products → cycles.
	log.Println("Matching machines to OS products...")
	groups := make(map[string]*cycleGroup)

	type unmatchedInfo struct {
		osName   string
		osType   string
		machines int
	}
	unmatchedMap := make(map[string]*unmatchedInfo)

	for _, m := range machines {
		slug, cycle, fullVer := parseOSName(m.OSName, m.OSType, knownCycleSets, winBuildToCycle)
		if slug == "" {
			displayName := m.OSName
			if displayName == "" {
				displayName = "(empty)"
			}
			key := displayName + "\x00" + m.OSType
			ui, ok := unmatchedMap[key]
			if !ok {
				ui = &unmatchedInfo{osName: displayName, osType: m.OSType}
				unmatchedMap[key] = ui
			}
			ui.machines++
			continue
		}

		key := slug + ":" + cycle
		g, ok := groups[key]
		if !ok {
			displayName := osSlugs[slug]
			if displayName == "" {
				displayName = slug
			}
			var matched *eolCycle
			if slugCycles, ok := cycles[slug]; ok {
				for i := range slugCycles {
					if slugCycles[i].Cycle == cycle {
						matched = &slugCycles[i]
						break
					}
				}
			}
			g = &cycleGroup{
				Product:       displayName,
				Slug:          slug,
				Cycle:         cycle,
				EOLCycle:      matched,
				MinorVersions: make(map[string]int),
			}
			groups[key] = g
		}
		g.Machines = append(g.Machines, m)
		if fullVer != "" {
			g.MinorVersions[fullVer]++
		}
	}

	// 5. Display lifecycle report.
	sorted := make([]*cycleGroup, 0, len(groups))
	var hiddenCount int
	for _, g := range groups {
		if !*showAll && g.EOLCycle == nil {
			hiddenCount++
			continue
		}
		sorted = append(sorted, g)
	}
	sortCycleGroups(sorted)

	productSet := make(map[string]bool)
	matchedCount := 0
	for _, g := range sorted {
		productSet[g.Slug] = true
		matchedCount += len(g.Machines)
	}

	printSection(fmt.Sprintf("OS Lifecycle — %d products, %d cycles", len(productSet), len(sorted)))
	headers := []string{"Product", "Cycle", "EOL", "EOS", "Machines"}
	var rows [][]string
	for _, g := range sorted {
		eol := "N/A"
		eos := "N/A"
		if g.EOLCycle != nil {
			eol = formatTimeRemaining(g.EOLCycle.EOL, now)
			if g.EOLCycle.EOES.Valid {
				eos = formatTimeRemaining(g.EOLCycle.EOES, now)
			} else if g.EOLCycle.EOAS.Valid {
				eos = formatTimeRemaining(g.EOLCycle.EOAS, now)
			}
		}
		rows = append(rows, []string{
			g.Product, formatCycle(g.Slug, g.Cycle), eol, eos,
			fmt.Sprintf("%d", len(g.Machines)),
		})
	}
	printTable(headers, rows)
	if hiddenCount > 0 {
		fmt.Printf("\n  (%d cycles without EOL reference data hidden; use --all to show)\n", hiddenCount)
	}
	fmt.Printf("\n  %d matched machines\n", matchedCount)

	// 6. Unmatched machines table.
	unmatchedTotal := 0
	for _, ui := range unmatchedMap {
		unmatchedTotal += ui.machines
	}
	if unmatchedTotal > 0 {
		unmatchedSorted := make([]*unmatchedInfo, 0, len(unmatchedMap))
		for _, ui := range unmatchedMap {
			unmatchedSorted = append(unmatchedSorted, ui)
		}
		for i := 1; i < len(unmatchedSorted); i++ {
			for j := i; j > 0 && unmatchedSorted[j].machines > unmatchedSorted[j-1].machines; j-- {
				unmatchedSorted[j], unmatchedSorted[j-1] = unmatchedSorted[j-1], unmatchedSorted[j]
			}
		}

		printSection(fmt.Sprintf("Unmatched — %d machines", unmatchedTotal))
		uHeaders := []string{"OS Name", "OS Type", "Machines"}
		var uRows [][]string
		for _, ui := range unmatchedSorted {
			uRows = append(uRows, []string{
				truncate(ui.osName, 50), ui.osType,
				fmt.Sprintf("%d", ui.machines),
			})
		}
		printTable(uHeaders, uRows)
	}

	fmt.Printf("\n  Total: %d machines (%d matched, %d unmatched)\n\n",
		len(machines), matchedCount, unmatchedTotal)

	// Minor version breakdown.
	if *showMinor {
		for _, g := range sorted {
			if len(g.MinorVersions) <= 1 {
				continue
			}
			fmt.Printf("  %s %s breakdown:\n", g.Product, g.Cycle)
			type verCount struct {
				ver   string
				count int
			}
			vers := make([]verCount, 0, len(g.MinorVersions))
			for v, c := range g.MinorVersions {
				vers = append(vers, verCount{v, c})
			}
			// Sort by version string.
			for i := 1; i < len(vers); i++ {
				for j := i; j > 0 && vers[j].ver < vers[j-1].ver; j-- {
					vers[j], vers[j-1] = vers[j-1], vers[j]
				}
			}
			for _, vc := range vers {
				fmt.Printf("    %-12s %d machines\n", vc.ver, vc.count)
			}
			fmt.Println()
		}
	}

	// 6. Machine sub-tables.
	if !*showMachines {
		return
	}

	for _, g := range sorted {
		fmt.Printf("\n--- %s %s (%d machines) ---\n\n", g.Product, g.Cycle, len(g.Machines))
		mHeaders := []string{"Hostname", "Env", "OS Name", "Status", "Project"}
		var mRows [][]string
		for _, m := range g.Machines {
			mRows = append(mRows, []string{
				truncate(m.Hostname, 30), truncate(m.Environment, 12),
				truncate(m.OSName, 40), m.Status,
				truncate(m.CloudProject, 30),
			})
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
	statusFilter := fs.String("status", "", "Filter by machine status (e.g., running, stopped)")
	limit := fs.Int("limit", 0, "Max unmatched os_name values to show (0 = all)")
	fs.Parse(args)

	// 1. Load OS EOL cycle sets (for parseOSName).
	log.Println("Loading OS EOL products...")
	osSlugs, err := loadOSSlugs(ctx, db)
	if err != nil {
		log.Fatalf("load OS slugs: %v", err)
	}
	slugList := make([]string, 0, len(osSlugs))
	for slug := range osSlugs {
		slugList = append(slugList, slug)
	}
	cycles, err := loadEOLCycles(ctx, db, slugList)
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
	winBuildToCycle := buildWindowsBuildMap(cycles["windows"])

	// 2. Load machines.
	log.Println("Loading machines...")
	projects := parseCSV(*projectFilter)
	machines, err := loadMachines(ctx, db, *osFilter, *envFilter, *statusFilter, projects)
	if err != nil {
		log.Fatalf("load machines: %v", err)
	}
	log.Printf("  Loaded %d machines", len(machines))

	// 3. Find unmatched.
	type unmatchedInfo struct {
		osName   string
		osType   string
		machines int
	}
	unmatched := make(map[string]*unmatchedInfo)
	var matchedCount int
	for _, m := range machines {
		slug, _, _ := parseOSName(m.OSName, m.OSType, knownCycleSets, winBuildToCycle)
		if slug != "" {
			matchedCount++
			continue
		}
		displayName := m.OSName
		if displayName == "" {
			displayName = "(empty)"
		}
		key := displayName + "\x00" + m.OSType
		ui, ok := unmatched[key]
		if !ok {
			ui = &unmatchedInfo{osName: displayName, osType: m.OSType}
			unmatched[key] = ui
		}
		ui.machines++
	}
	log.Printf("  Matched: %d, unmatched: %d machines (%d distinct os_name values)",
		matchedCount, len(machines)-matchedCount, len(unmatched))

	// Sort by machine count descending.
	sorted := make([]*unmatchedInfo, 0, len(unmatched))
	for _, ui := range unmatched {
		sorted = append(sorted, ui)
	}
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].machines > sorted[j-1].machines; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}

	if *limit > 0 && len(sorted) > *limit {
		sorted = sorted[:*limit]
	}

	printSection(fmt.Sprintf("Unmatched OS — %d os_name values (sorted by machine count)", len(unmatched)))
	headers := []string{"OS Name", "OS Type", "Machines"}
	var rows [][]string
	for _, ui := range sorted {
		rows = append(rows, []string{
			truncate(ui.osName, 50), ui.osType,
			fmt.Sprintf("%d", ui.machines),
		})
	}
	printTable(headers, rows)
	fmt.Printf("\n  Total: %d matched, %d unmatched machines\n\n", matchedCount, len(machines)-matchedCount)
}

// --- OS name parsing ---

// osPattern defines how to match an os_name string to an endoflife.date product slug.
type osPattern struct {
	slug       string
	match      func(lower string) bool
	extractVer func(lower string) string
	cycleDepth int // how many version segments to keep (1 = major, 2 = major.minor); 0 = use extractVer as-is
}

// versionRe extracts the first version-like string (digits possibly with dots).
var versionRe = regexp.MustCompile(`(\d+(?:\.\d+)*)`)

var osPatterns = []osPattern{
	// Windows Server must come before Windows desktop.
	// Handles both "Windows Server 2019" and GreenNode "Windows-Server2012R2".
	{
		slug: "windows-server",
		match: func(s string) bool {
			return strings.Contains(s, "windows server") || strings.Contains(s, "windows-server")
		},
		extractVer: func(s string) string {
			if idx := strings.Index(s, "windows server"); idx >= 0 {
				return extractFirstVersion(s[idx:])
			}
			if idx := strings.Index(s, "windows-server"); idx >= 0 {
				return extractFirstVersion(s[idx:])
			}
			return ""
		},
		cycleDepth: 1,
	},
	{
		slug: "windows",
		match: func(s string) bool {
			return strings.Contains(s, "windows") &&
				!strings.Contains(s, "windows server") &&
				!strings.Contains(s, "windows-server")
		},
		extractVer: func(s string) string { return extractAfter(s, "windows") },
		cycleDepth: 1,
	},
	{
		slug:       "macos",
		match:      func(s string) bool { return strings.Contains(s, "macos") || strings.Contains(s, "mac os") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	// macOS revision-only: S1 sometimes returns just "15.7.4 (24G517)" without "macOS" prefix.
	{
		slug:       "macos",
		match:      func(s string) bool { return isMacOSRevision(s) },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "ubuntu",
		match:      func(s string) bool { return strings.Contains(s, "ubuntu") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 2,
	},
	{
		slug: "rhel",
		match: func(s string) bool {
			return strings.Contains(s, "red hat enterprise") ||
				strings.Contains(s, "redhat-enterprise") ||
				strings.Contains(s, "red-hat-enterprise") ||
				strings.HasPrefix(s, "rhel")
		},
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "centos",
		match:      func(s string) bool { return strings.Contains(s, "centos") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "rocky-linux",
		match:      func(s string) bool { return strings.Contains(s, "rocky") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "almalinux",
		match:      func(s string) bool { return strings.Contains(s, "alma") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "oracle-linux",
		match:      func(s string) bool { return strings.Contains(s, "oracle linux") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug: "amazon-linux",
		match: func(s string) bool {
			return strings.Contains(s, "amazon linux") || strings.Contains(s, "amzn")
		},
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "debian",
		match:      func(s string) bool { return strings.Contains(s, "debian") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	// SLES: "SUSE Linux Enterprise Server 15 SP5" → cycle "15.5"
	{
		slug:  "sles",
		match: func(s string) bool { return strings.Contains(s, "suse") && strings.Contains(s, "enterprise") },
		extractVer: func(s string) string {
			ver := extractFirstVersion(s)
			if idx := strings.Index(s, " sp"); idx >= 0 {
				sp := extractFirstVersion(s[idx:])
				if sp != "" {
					return ver + "." + sp
				}
			}
			return ver
		},
		cycleDepth: 0,
	},
	{
		slug: "opensuse",
		match: func(s string) bool {
			return strings.Contains(s, "opensuse") || strings.Contains(s, "open suse")
		},
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 2,
	},
}

// windowsBuildRe extracts the build number from "Windows 11 Pro (Build 26200)".
var windowsBuildRe = regexp.MustCompile(`(?i)\(build\s+(\d+)\)`)

// parseOSName matches an os_name string to an endoflife.date product slug and cycle.
func parseOSName(osName, osType string, knownCycles map[string]map[string]bool, winBuildToCycle map[string]string) (slug, cycle, fullVer string) {
	if osName == "" {
		return "", "", ""
	}
	lower := strings.ToLower(osName)

	for _, p := range osPatterns {
		if !p.match(lower) {
			continue
		}

		// Windows desktop: use build number for precise cycle matching.
		if p.slug == "windows" {
			if m := windowsBuildRe.FindStringSubmatch(osName); m != nil {
				build := m[1]
				if c, ok := winBuildToCycle[build]; ok {
					return "windows", c, build
				}
			}
			// Fall back to major version (10/11) if no build match.
			ver := p.extractVer(lower)
			return p.slug, extractCycle(ver, 1), ver
		}

		ver := p.extractVer(lower)
		if ver == "" {
			return p.slug, "", ""
		}
		if p.cycleDepth > 0 {
			cycle = extractCycle(ver, p.cycleDepth)
		} else {
			cycle = ver
		}
		// Try matching against known cycles — prefer more specific (minor) over major.
		if p.cycleDepth >= 1 {
			if known, ok := knownCycles[p.slug]; ok {
				c2 := extractCycle(ver, 2)
				if known[c2] {
					return p.slug, c2, ver
				}
				c1 := extractCycle(ver, 1)
				if known[c1] {
					return p.slug, c1, ver
				}
			}
		}
		return p.slug, cycle, ver
	}
	return "", "", ""
}

// buildWindowsBuildMap creates a map from Windows build number to the best
// endoflife.date cycle name. For "Pro" edition we prefer "-w" (workstation)
// cycles, and for cycles without suffix we use the base cycle.
// Each build maps to exactly one cycle (the most relevant for non-enterprise).
func buildWindowsBuildMap(windowsCycles []eolCycle) map[string]string {
	result := make(map[string]string)
	for _, c := range windowsCycles {
		if c.Latest == "" {
			continue
		}
		// latest is like "10.0.26200" → extract build "26200"
		parts := strings.Split(c.Latest, ".")
		if len(parts) < 3 {
			continue
		}
		build := parts[2]

		// Prefer: base cycle (no suffix) > "-w" > anything else.
		// Skip "-e", "-e-lts", "-iot-lts" variants.
		cycleName := c.Cycle
		if strings.HasSuffix(cycleName, "-e") || strings.Contains(cycleName, "-e-") || strings.Contains(cycleName, "-iot") {
			// Only use enterprise/iot if we don't have a better match.
			if _, ok := result[build]; !ok {
				result[build] = cycleName
			}
			continue
		}
		// "-w" or base cycle — prefer this.
		result[build] = cycleName
	}
	return result
}

// macOSRevisionRe matches S1 os_revision format: "15.7.4 (24G517)" or "26.2 (25C56)".
// Version number followed by optional space and parenthesized build identifier.
var macOSRevisionRe = regexp.MustCompile(`^\d+\.\d+(?:\.\d+)?\s*\(`)

// isMacOSRevision detects bare macOS version strings without "macOS" prefix.
// S1's bestOSName falls back to os_revision when os_name is generic "macOS",
// producing strings like "15.7.4 (24G517)".
func isMacOSRevision(s string) bool {
	return macOSRevisionRe.MatchString(s)
}

// extractAfter finds the keyword in s and returns the first version after it.
func extractAfter(s, keyword string) string {
	idx := strings.Index(s, keyword)
	if idx < 0 {
		return ""
	}
	rest := s[idx+len(keyword):]
	return extractFirstVersion(rest)
}

// extractFirstVersion returns the first version-like substring (digits with dots).
func extractFirstVersion(s string) string {
	return versionRe.FindString(s)
}

// extractCycle takes a version string and returns the first N dot-separated segments.
func extractCycle(version string, depth int) string {
	parts := strings.Split(version, ".")
	if len(parts) < depth {
		return version
	}
	return strings.Join(parts[:depth], ".")
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

type machineRecord struct {
	MachineID    string
	Hostname     string
	OSType       string
	OSName       string
	Status       string
	Environment  string
	CloudProject string
}

func loadOSSlugs(ctx context.Context, db *sql.DB) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate OS products: %w", err)
	}
	log.Printf("  Loaded %d OS products from endoflife.date", len(result))
	return result, nil
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

func loadMachines(ctx context.Context, db *sql.DB, osFilter, envFilter, statusFilter string, projects []string) ([]machineRecord, error) {
	query := `
		SELECT resource_id, hostname, os_type,
		       COALESCE(os_name, ''), status,
		       COALESCE(environment, ''), COALESCE(cloud_project, '')
		FROM inventory.machines
		WHERE 1=1`

	argN := 1
	var args []any
	if osFilter != "" {
		query += fmt.Sprintf(` AND os_type = $%d`, argN)
		args = append(args, osFilter)
		argN++
	}
	if envFilter != "" {
		query += fmt.Sprintf(` AND environment = $%d`, argN)
		args = append(args, envFilter)
		argN++
	}
	if statusFilter != "" {
		query += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, statusFilter)
		argN++
	}
	for _, proj := range projects {
		query += fmt.Sprintf(` AND cloud_project ILIKE '%%' || $%d || '%%'`, argN)
		args = append(args, proj)
		argN++
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query machines: %w", err)
	}
	defer rows.Close()

	var result []machineRecord
	for rows.Next() {
		var m machineRecord
		if err := rows.Scan(&m.MachineID, &m.Hostname, &m.OSType,
			&m.OSName, &m.Status, &m.Environment, &m.CloudProject); err != nil {
			return nil, fmt.Errorf("scan machine: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// --- Formatting helpers ---

// formatCycle returns a human-readable cycle name.
// Windows cycles like "11-25h2-w" become "11 25H2", "10-1607-e-lts" becomes "10 1607 LTS".
func formatCycle(slug, cycle string) string {
	if slug != "windows" {
		return cycle
	}
	// Strip the major version prefix "10-" or "11-".
	parts := strings.SplitN(cycle, "-", 2)
	if len(parts) < 2 {
		return cycle
	}
	major := parts[0]
	rest := parts[1]

	// Remove edition suffix: -w, -e, -e-lts, -iot-lts → extract release + qualifier.
	release := rest
	qualifier := ""
	if strings.HasSuffix(release, "-iot-lts") {
		release = strings.TrimSuffix(release, "-iot-lts")
		qualifier = " IoT LTS"
	} else if strings.HasSuffix(release, "-e-lts") {
		release = strings.TrimSuffix(release, "-e-lts")
		qualifier = " LTS"
	} else if strings.HasSuffix(release, "-w") {
		release = strings.TrimSuffix(release, "-w")
	} else if strings.HasSuffix(release, "-e") {
		release = strings.TrimSuffix(release, "-e")
	}

	return major + " " + strings.ToUpper(release) + qualifier
}

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

func sortCycleGroups(groups []*cycleGroup) {
	for i := 1; i < len(groups); i++ {
		for j := i; j > 0 && (groups[j].Product < groups[j-1].Product ||
			(groups[j].Product == groups[j-1].Product && groups[j].Cycle < groups[j-1].Cycle)); j-- {
			groups[j], groups[j-1] = groups[j-1], groups[j]
		}
	}
}

func sortStringSlices(rows [][]string) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j][0] < rows[j-1][0]; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}
