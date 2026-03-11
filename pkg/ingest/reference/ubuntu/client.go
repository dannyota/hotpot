package ubuntu

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"danny.vn/hotpot/pkg/base/httputil"
)

// FeedDef defines a single Ubuntu Packages.gz feed to download.
type FeedDef struct {
	Release   string // e.g. "noble", "jammy"
	Component string // e.g. "main", "universe"
	URL       string
}

// Feeds lists all Ubuntu Packages.gz feeds to ingest.
var Feeds = []FeedDef{
	{"noble", "main", "http://archive.ubuntu.com/ubuntu/dists/noble/main/binary-amd64/Packages.gz"},
	{"noble", "universe", "http://archive.ubuntu.com/ubuntu/dists/noble/universe/binary-amd64/Packages.gz"},
	{"jammy", "main", "http://archive.ubuntu.com/ubuntu/dists/jammy/main/binary-amd64/Packages.gz"},
	{"jammy", "universe", "http://archive.ubuntu.com/ubuntu/dists/jammy/universe/binary-amd64/Packages.gz"},
}

// Client downloads and parses Ubuntu Packages indexes.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Ubuntu client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// UbuntuPackageData holds a parsed Ubuntu package entry.
type UbuntuPackageData struct {
	PackageName string
	Release     string
	Component   string
	Section     string
	Description string
}

// DownloadFeed fetches a single Ubuntu Packages.gz feed and parses it.
func (c *Client) DownloadFeed(feed FeedDef, heartbeat func(string)) ([]UbuntuPackageData, error) {
	resp, err := c.httpClient.Get(feed.URL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", feed.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", feed.URL, resp.StatusCode)
	}

	// Download to temp file
	tmpFile, err := os.CreateTemp("", "ubuntu-packages-*.gz")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	label := fmt.Sprintf("%s/%s", feed.Release, feed.Component)
	body := httputil.NewProgressReader(resp.Body, resp.ContentLength, label, 5*time.Second, heartbeat)
	if _, err := io.Copy(tmpFile, body); err != nil {
		return nil, fmt.Errorf("download to temp: %w", err)
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek temp file: %w", err)
	}

	gz, err := gzip.NewReader(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	return parsePackages(gz, feed.Release, feed.Component)
}

// parsePackages parses the Debian Packages format (stanza-based, field: value).
func parsePackages(r io.Reader, release, component string) ([]UbuntuPackageData, error) {
	var result []UbuntuPackageData
	scanner := bufio.NewScanner(r)
	// Increase buffer for long lines
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var pkg, section, description string
	// Deduplicate — Packages index can list the same package name multiple
	// times (e.g., different architectures merged into one index).
	seen := make(map[string]struct{})

	flush := func() {
		if pkg != "" && section != "" {
			if _, dup := seen[pkg]; !dup {
				seen[pkg] = struct{}{}
				result = append(result, UbuntuPackageData{
					PackageName: pkg,
					Release:     release,
					Component:   component,
					Section:     section,
					Description: description,
				})
			}
		}
		pkg = ""
		section = ""
		description = ""
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Empty line = end of stanza
		if line == "" {
			flush()
			continue
		}

		// Skip continuation lines (start with space or tab)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}

		switch key {
		case "Package":
			pkg = value
		case "Section":
			section = value
		case "Description":
			description = value
		}
	}

	// Flush last stanza
	flush()

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return result, nil
}
