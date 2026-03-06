package cpe

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"danny.vn/hotpot/pkg/base/httputil"
)

const cpeFeedURL = "https://nvd.nist.gov/feeds/json/cpe/2.0/nvdcpe-2.0.zip"

// Client downloads and parses the NVD CPE Dictionary.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new CPE client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// CPEData holds a parsed CPE entry.
type CPEData struct {
	CPEName    string
	Part       string
	Vendor     string
	Product    string
	Version    string
	Title      string
	Deprecated bool
}

// cpeFeed is the top-level JSON structure of each chunk file.
type cpeFeed struct {
	Products []cpeProduct `json:"products"`
}

type cpeProduct struct {
	CPE cpeFeedItem `json:"cpe"`
}

type cpeFeedItem struct {
	CPEName    string     `json:"cpeName"`
	Deprecated bool       `json:"deprecated"`
	Titles     []cpeTitle `json:"titles"`
}

type cpeTitle struct {
	Title string `json:"title"`
	Lang  string `json:"lang"`
}

// LastModified returns the Last-Modified time from the CPE feed.
// Returns empty string if the request fails.
func (c *Client) LastModified() (string, error) {
	resp, err := c.httpClient.Head(cpeFeedURL)
	if err != nil {
		return "", fmt.Errorf("HEAD %s: %w", cpeFeedURL, err)
	}
	resp.Body.Close()
	return resp.Header.Get("Last-Modified"), nil
}

// Download fetches the CPE zip, extracts and parses all JSON files inside.
// Filters out hardware entries (part=h). Calls heartbeat periodically.
func (c *Client) Download(heartbeat func(string)) ([]CPEData, error) {
	resp, err := c.httpClient.Get(cpeFeedURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", cpeFeedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", cpeFeedURL, resp.StatusCode)
	}

	// Download to temp file (zip requires random access)
	tmpFile, err := os.CreateTemp("", "nvdcpe-*.zip")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	body := httputil.NewProgressReader(resp.Body, resp.ContentLength, "nvd-cpe", 5*time.Second, heartbeat)
	size, err := io.Copy(tmpFile, body)
	if err != nil {
		return nil, fmt.Errorf("download to temp: %w", err)
	}

	heartbeat("parsing CPE data")

	// Open as zip
	zr, err := zip.NewReader(tmpFile, size)
	if err != nil {
		return nil, fmt.Errorf("zip reader: %w", err)
	}

	var all []CPEData
	for _, f := range zr.File {
		if !strings.HasSuffix(f.Name, ".json") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", f.Name, err)
		}

		chunk, err := parseChunk(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", f.Name, err)
		}
		all = append(all, chunk...)
		heartbeat(fmt.Sprintf("parsed %d CPE entries", len(all)))
	}

	return all, nil
}

// parseChunk decodes a single JSON chunk from the tar and extracts CPE data.
func parseChunk(r io.Reader) ([]CPEData, error) {
	var feed cpeFeed
	if err := json.NewDecoder(r).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}

	var result []CPEData
	for _, p := range feed.Products {
		item := p.CPE
		parts := parseCPEName(item.CPEName)
		if parts == nil {
			continue
		}
		// Skip hardware
		if parts.Part == "h" {
			continue
		}

		title := ""
		for _, t := range item.Titles {
			if t.Lang == "en" {
				title = t.Title
				break
			}
		}

		result = append(result, CPEData{
			CPEName:    item.CPEName,
			Part:       parts.Part,
			Vendor:     parts.Vendor,
			Product:    parts.Product,
			Version:    parts.Version,
			Title:      title,
			Deprecated: item.Deprecated,
		})
	}
	return result, nil
}

type cpeParts struct {
	Part    string
	Vendor  string
	Product string
	Version string
}

// parseCPEName parses a CPE 2.3 formatted string.
// Format: cpe:2.3:part:vendor:product:version:...
func parseCPEName(name string) *cpeParts {
	fields := strings.Split(name, ":")
	if len(fields) < 6 || fields[0] != "cpe" || fields[1] != "2.3" {
		return nil
	}
	return &cpeParts{
		Part:    fields[2],
		Vendor:  fields[3],
		Product: fields[4],
		Version: fields[5],
	}
}
