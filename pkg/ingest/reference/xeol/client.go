package xeol

import (
	"archive/tar"
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

	"github.com/dannyota/hotpot/pkg/base/httputil"
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

// XeolProductData holds a parsed xeol product entry.
type XeolProductData struct {
	ID          string
	Name        string
	PURL        string
	Permalink   string
	EOL         *time.Time
	EOLBool     bool
	LatestCycle string
	ReleaseDate *time.Time
	Latest      string
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

// Download fetches the latest xeol database and extracts product data.
func (c *Client) Download(heartbeat func(string)) ([]XeolProductData, error) {
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
	products, err := queryDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("query xeol DB: %w", err)
	}

	slog.Info("Parsed xeol products", "count", len(products))
	return products, nil
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
	tmpArchive, err := os.CreateTemp("", "xeol-db-*.tar.xz")
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

	// Decompress xz
	xzReader, err := xz.NewReader(tmpArchive)
	if err != nil {
		return "", fmt.Errorf("xz reader: %w", err)
	}

	// Extract xeol.db from tar
	tr := tar.NewReader(xzReader)
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

func queryDB(dbPath string) ([]XeolProductData, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	defer db.Close()

	// Query products with optional PURL and best cycle data.
	// LEFT JOIN purls: pick first PURL per product.
	// LEFT JOIN cycles: pick the cycle with the latest release_date.
	rows, err := db.Query(`
		SELECT
			p.id,
			p.name,
			p.permalink,
			pu.purl,
			c.release_cycle,
			c.eol,
			c.eol_bool,
			c.release_date,
			c.latest_release
		FROM products p
		LEFT JOIN (
			SELECT product_id, MIN(purl) as purl
			FROM purls
			GROUP BY product_id
		) pu ON pu.product_id = p.id
		LEFT JOIN (
			SELECT DISTINCT product_id, release_cycle, eol, eol_bool, release_date, latest_release,
				ROW_NUMBER() OVER (PARTITION BY product_id ORDER BY release_date DESC) as rn
			FROM cycles
		) c ON c.product_id = p.id AND c.rn = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	var products []XeolProductData
	for rows.Next() {
		var (
			productID                   int64
			name, permalink             string
			purl, cycle, latest         sql.NullString
			eolStr, releaseDateStr      sql.NullString
			eolBool                     sql.NullBool
		)

		if err := rows.Scan(&productID, &name, &permalink, &purl, &cycle, &eolStr, &eolBool, &releaseDateStr, &latest); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		ecosystem := deriveEcosystem(permalink)
		id := ecosystem + ":" + name

		// Deduplicate: same ecosystem+name can appear if permalink domains overlap.
		// Use SQLite product_id as tiebreaker — keep first seen.
		if seen[id] {
			id = fmt.Sprintf("%s:%s:%d", ecosystem, name, productID)
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		p := XeolProductData{
			ID:        id,
			Name:      name,
			Permalink: permalink,
		}

		if purl.Valid {
			p.PURL = purl.String
		}
		if cycle.Valid {
			p.LatestCycle = cycle.String
		}
		if latest.Valid {
			p.Latest = latest.String
		}
		if eolBool.Valid && eolBool.Bool {
			p.EOLBool = true
		}

		if eolStr.Valid && eolStr.String != "" {
			if t, err := time.Parse("2006-01-02", eolStr.String); err == nil {
				p.EOL = &t
			}
		}
		if releaseDateStr.Valid && releaseDateStr.String != "" {
			if t, err := time.Parse("2006-01-02", releaseDateStr.String); err == nil {
				p.ReleaseDate = &t
			}
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return products, nil
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
