package xeol

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
	_ "modernc.org/sqlite"

	"danny.vn/hotpot/pkg/base/httputil"
)

const listingURL = "https://data.xeol.io/xeol/databases/listing.json"

// Client downloads and parses the xeol EOL database.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new xeol client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// XeolData contains all parsed data from the xeol database.
type XeolData struct {
	Products []XeolProduct
	Cycles   []XeolCycle
	Purls    []XeolPurl
	Vulns    []XeolVuln
}

// XeolProduct holds a parsed xeol product entry.
type XeolProduct struct {
	ID        string
	Name      string
	Permalink string
}

// XeolCycle holds a parsed xeol cycle entry.
type XeolCycle struct {
	ID                string
	ProductID         string
	ReleaseCycle      string
	EOL               *time.Time
	EOLBool           bool
	LatestRelease     string
	LatestReleaseDate *time.Time
	ReleaseDate       *time.Time
}

// XeolPurl holds a parsed xeol purl entry.
type XeolPurl struct {
	ID        string
	ProductID string
	PURL      string
}

// XeolVuln holds a parsed xeol vulnerability entry.
type XeolVuln struct {
	ID         string
	ProductID  string
	Version    string
	IssueCount int
	Issues     string
}

// listingJSON is the top-level listing response.
type listingJSON struct {
	Available map[string][]listingEntry `json:"available"`
}

type listingEntry struct {
	Built    string `json:"built"`
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

// Download fetches the latest xeol database and extracts all data.
func (c *Client) Download(heartbeat func(string)) (*XeolData, error) {
	// Step 1: Find latest DB URL
	heartbeat("fetching xeol database listing")
	dbURL, err := c.findLatestDBURL()
	if err != nil {
		return nil, fmt.Errorf("find latest DB URL: %w", err)
	}
	slog.Info("Found latest xeol database", "url", dbURL)

	// Step 2: Download tar.xz
	heartbeat("downloading xeol database")
	dbPath, err := c.downloadDB(dbURL, heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download xeol DB: %w", err)
	}
	defer os.Remove(dbPath)

	// Step 3: Query SQLite
	heartbeat("querying xeol database")
	data, err := queryDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("query xeol DB: %w", err)
	}

	slog.Info("Parsed xeol data",
		"products", len(data.Products),
		"cycles", len(data.Cycles),
		"purls", len(data.Purls),
		"vulns", len(data.Vulns),
	)
	return data, nil
}

func (c *Client) findLatestDBURL() (string, error) {
	resp, err := c.httpClient.Get(listingURL)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", listingURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: status %d", listingURL, resp.StatusCode)
	}

	var listing listingJSON
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return "", fmt.Errorf("decode listing: %w", err)
	}

	// Find the latest entry across all schema versions
	var latestURL string
	var latestTime time.Time

	for _, entries := range listing.Available {
		for _, e := range entries {
			t, err := time.Parse(time.RFC3339Nano, e.Built)
			if err != nil {
				continue
			}
			if t.After(latestTime) {
				latestTime = t
				latestURL = e.URL
			}
		}
	}

	if latestURL == "" {
		return "", fmt.Errorf("no database entries found in listing")
	}

	return latestURL, nil
}

func (c *Client) downloadDB(dbURL string, heartbeat func(string)) (string, error) {
	resp, err := c.httpClient.Get(dbURL)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", dbURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: status %d", dbURL, resp.StatusCode)
	}

	// Download to temp file
	tmpArchive, err := os.CreateTemp("", "xeol-db-*")
	if err != nil {
		return "", fmt.Errorf("create temp archive: %w", err)
	}
	defer os.Remove(tmpArchive.Name())
	defer tmpArchive.Close()

	body := httputil.NewProgressReader(resp.Body, resp.ContentLength, "xeol-db", 5*time.Second, heartbeat)
	if _, err := io.Copy(tmpArchive, body); err != nil {
		return "", fmt.Errorf("download to temp: %w", err)
	}
	if _, err := tmpArchive.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek temp file: %w", err)
	}

	// Decompress: try xz first, fall back to gzip
	var decompressed io.Reader
	if strings.HasSuffix(dbURL, ".tar.xz") {
		r, err := xz.NewReader(tmpArchive)
		if err != nil {
			return "", fmt.Errorf("xz reader: %w", err)
		}
		decompressed = r
	} else {
		r, err := gzip.NewReader(tmpArchive)
		if err != nil {
			return "", fmt.Errorf("gzip reader: %w", err)
		}
		defer r.Close()
		decompressed = r
	}

	// Extract xeol.db from tar
	tr := tar.NewReader(decompressed)
	tmpDB, err := os.CreateTemp("", "xeol-*.db")
	if err != nil {
		return "", fmt.Errorf("create temp DB: %w", err)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			tmpDB.Close()
			os.Remove(tmpDB.Name())
			return "", fmt.Errorf("xeol.db not found in archive")
		}
		if err != nil {
			tmpDB.Close()
			os.Remove(tmpDB.Name())
			return "", fmt.Errorf("tar next: %w", err)
		}

		if strings.HasSuffix(hdr.Name, ".db") {
			heartbeat("extracting xeol.db")
			if _, err := io.Copy(tmpDB, tr); err != nil {
				tmpDB.Close()
				os.Remove(tmpDB.Name())
				return "", fmt.Errorf("extract xeol.db: %w", err)
			}
			tmpDB.Close()
			return tmpDB.Name(), nil
		}
	}
}

func queryDB(dbPath string) (*XeolData, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	defer db.Close()

	products, idMap, err := queryProducts(db)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}

	cycles, err := queryCycles(db, idMap)
	if err != nil {
		return nil, fmt.Errorf("query cycles: %w", err)
	}

	purls, err := queryPurls(db, idMap)
	if err != nil {
		return nil, fmt.Errorf("query purls: %w", err)
	}

	vulns, err := queryVulns(db, idMap)
	if err != nil {
		return nil, fmt.Errorf("query vulns: %w", err)
	}

	return &XeolData{
		Products: products,
		Cycles:   cycles,
		Purls:    purls,
		Vulns:    vulns,
	}, nil
}

// queryProducts returns products and a map of SQLite product ID -> composite ID for child lookups.
func queryProducts(db *sql.DB) ([]XeolProduct, map[int64]string, error) {
	rows, err := db.Query(`SELECT id, name, permalink FROM products`)
	if err != nil {
		return nil, nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	idMap := make(map[int64]string)
	var products []XeolProduct

	for rows.Next() {
		var (
			productID int64
			name      string
			permalink string
		)

		if err := rows.Scan(&productID, &name, &permalink); err != nil {
			return nil, nil, fmt.Errorf("scan: %w", err)
		}

		ecosystem := deriveEcosystem(permalink)
		id := ecosystem + ":" + name

		// Deduplicate: same ecosystem+name can appear if permalink domains overlap.
		if seen[id] {
			id = fmt.Sprintf("%s:%s:%d", ecosystem, name, productID)
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		idMap[productID] = id

		products = append(products, XeolProduct{
			ID:        id,
			Name:      name,
			Permalink: permalink,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate: %w", err)
	}

	return products, idMap, nil
}

func queryCycles(db *sql.DB, idMap map[int64]string) ([]XeolCycle, error) {
	rows, err := db.Query(`
		SELECT id, product_id, release_cycle, eol, eol_bool, latest_release, latest_release_date, release_date
		FROM cycles
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var cycles []XeolCycle
	for rows.Next() {
		var (
			cycleID                                    int64
			productID                                  int64
			releaseCycle                               string
			eolStr, latestRelease                      sql.NullString
			latestReleaseDateStr, releaseDateStr        sql.NullString
			eolBool                                    sql.NullBool
		)

		if err := rows.Scan(&cycleID, &productID, &releaseCycle, &eolStr, &eolBool, &latestRelease, &latestReleaseDateStr, &releaseDateStr); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		parentID, ok := idMap[productID]
		if !ok {
			continue // orphan cycle
		}

		c := XeolCycle{
			ID:           fmt.Sprintf("%s:%d", parentID, cycleID),
			ProductID:    parentID,
			ReleaseCycle: releaseCycle,
		}

		if eolBool.Valid && eolBool.Bool {
			c.EOLBool = true
		}
		if eolStr.Valid && eolStr.String != "" {
			if t, err := time.Parse(time.RFC3339, eolStr.String); err == nil {
				c.EOL = &t
			}
		}
		if latestRelease.Valid {
			c.LatestRelease = latestRelease.String
		}
		if latestReleaseDateStr.Valid && latestReleaseDateStr.String != "" {
			if t, err := time.Parse(time.RFC3339, latestReleaseDateStr.String); err == nil {
				c.LatestReleaseDate = &t
			}
		}
		if releaseDateStr.Valid && releaseDateStr.String != "" {
			if t, err := time.Parse(time.RFC3339, releaseDateStr.String); err == nil {
				c.ReleaseDate = &t
			}
		}

		cycles = append(cycles, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate: %w", err)
	}

	return cycles, nil
}

func queryPurls(db *sql.DB, idMap map[int64]string) ([]XeolPurl, error) {
	rows, err := db.Query(`SELECT product_id, purl FROM purls`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var purls []XeolPurl
	for rows.Next() {
		var (
			productID int64
			purl      string
		)

		if err := rows.Scan(&productID, &purl); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		parentID, ok := idMap[productID]
		if !ok {
			continue // orphan purl
		}

		purls = append(purls, XeolPurl{
			ID:        parentID + ":" + purl,
			ProductID: parentID,
			PURL:      purl,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate: %w", err)
	}

	return purls, nil
}

func queryVulns(db *sql.DB, idMap map[int64]string) ([]XeolVuln, error) {
	rows, err := db.Query(`SELECT product_id, version, issue_count, issues FROM vulns`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var vulns []XeolVuln
	for rows.Next() {
		var (
			productID  int64
			version    string
			issueCount int
			issues     string
		)

		if err := rows.Scan(&productID, &version, &issueCount, &issues); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		parentID, ok := idMap[productID]
		if !ok {
			continue // orphan vuln
		}

		vulns = append(vulns, XeolVuln{
			ID:         parentID + ":" + version,
			ProductID:  parentID,
			Version:    version,
			IssueCount: issueCount,
			Issues:     issues,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate: %w", err)
	}

	return vulns, nil
}

// deriveEcosystem maps a permalink URL to an ecosystem identifier.
func deriveEcosystem(permalink string) string {
	switch {
	case strings.Contains(permalink, "npmjs.com"):
		return "npm"
	case strings.Contains(permalink, "pypi.org"):
		return "pypi"
	case strings.Contains(permalink, "sonatype.com") || strings.Contains(permalink, "maven"):
		return "maven"
	case strings.Contains(permalink, "crates.io"):
		return "cargo"
	case strings.Contains(permalink, "nuget.org"):
		return "nuget"
	case strings.Contains(permalink, "rubygems.org"):
		return "rubygems"
	case strings.Contains(permalink, "pkg.xeol.io"):
		return "golang"
	default:
		return "other"
	}
}
