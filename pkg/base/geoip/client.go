package geoip

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang/v2"
)

// Client downloads GeoIP .mmdb files.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new download client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

// CityDownloadURL returns the current month's DB-IP city download URL.
func CityDownloadURL() string {
	now := time.Now()
	return fmt.Sprintf("https://download.db-ip.com/free/dbip-city-lite-%d-%02d.mmdb.gz",
		now.Year(), now.Month())
}

// ASNDownloadURL returns the IPinfo free ASN download URL.
// Requires a valid IPinfo token.
func ASNDownloadURL(token string) string {
	return fmt.Sprintf("https://ipinfo.io/data/free/asn.mmdb?token=%s", token)
}

// ASNDbipDownloadURL returns the current month's DB-IP ASN lite download URL (gzipped).
func ASNDbipDownloadURL() string {
	now := time.Now()
	return fmt.Sprintf("https://download.db-ip.com/free/dbip-asn-lite-%d-%02d.mmdb.gz",
		now.Year(), now.Month())
}

// DownloadGzipped fetches a .mmdb.gz file, gunzips, writes atomically to destPath.
// Uses If-Modified-Since from the local file to skip re-downloading unchanged files.
// Returns true if the file was updated, false if the server returned 304 Not Modified.
func (c *Client) DownloadGzipped(ctx context.Context, url, destPath string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}
	setIfModifiedSince(req, destPath)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return false, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	return true, writeAtomic(destPath, gz)
}

// DownloadRaw fetches a raw (non-gzipped) file and writes atomically to destPath.
// Uses If-Modified-Since from the local file to skip re-downloading unchanged files.
// Returns true if the file was updated, false if the server returned 304 Not Modified.
func (c *Client) DownloadRaw(ctx context.Context, url, destPath string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}
	setIfModifiedSince(req, destPath)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	return true, writeAtomic(destPath, resp.Body)
}

// setIfModifiedSince adds an If-Modified-Since header based on the local file's mod time.
func setIfModifiedSince(req *http.Request, localPath string) {
	info, err := os.Stat(localPath)
	if err != nil {
		return
	}
	req.Header.Set("If-Modified-Since", info.ModTime().UTC().Format(http.TimeFormat))
}

// ASNSourceIsIPInfo returns true if the mmdb file at path is an IPinfo database.
// Returns false if the file is missing, unreadable, or a DB-IP/MaxMind database.
func ASNSourceIsIPInfo(path string) bool {
	db, err := maxminddb.Open(path)
	if err != nil {
		return false
	}
	defer db.Close()
	return strings.Contains(strings.ToLower(db.Metadata.DatabaseType), "ipinfo")
}

// writeAtomic writes from r to destPath via temp file + os.Rename.
func writeAtomic(destPath string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(destPath), ".geoip-*.mmdb.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename to %s: %w", destPath, err)
	}

	return nil
}
