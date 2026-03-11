package cpe

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"danny.vn/hotpot/pkg/base/httputil"
)

const cpeFeedURL     = "https://nvd.nist.gov/feeds/json/cpe/2.0/nvdcpe-2.0.tar.gz"
const cpeFeedReferer = "https://nvd.nist.gov/vuln/data-feeds"

// setBrowserHeaders sets headers mimicking a Chrome browser click from the NVD data-feeds page.
// NVD throttles non-browser requests significantly (~5x slower).
func setBrowserHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", cpeFeedReferer)
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
}

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
	req, err := http.NewRequest(http.MethodHead, cpeFeedURL, nil)
	if err != nil {
		return "", fmt.Errorf("HEAD %s: %w", cpeFeedURL, err)
	}
	setBrowserHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HEAD %s: %w", cpeFeedURL, err)
	}
	resp.Body.Close()
	return resp.Header.Get("Last-Modified"), nil
}

// Download fetches the CPE tar.gz, streams and parses all JSON files inside.
// Filters out hardware entries (part=h). Calls heartbeat periodically.
func (c *Client) Download(heartbeat func(string)) ([]CPEData, error) {
	req, err := http.NewRequest(http.MethodGet, cpeFeedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", cpeFeedURL, err)
	}
	setBrowserHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", cpeFeedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", cpeFeedURL, resp.StatusCode)
	}

	body := httputil.NewProgressReader(resp.Body, resp.ContentLength, "nvd-cpe", 5*time.Second, heartbeat)

	gr, err := gzip.NewReader(body)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	var all []CPEData
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar next: %w", err)
		}
		if !strings.HasSuffix(hdr.Name, ".json") {
			continue
		}

		chunk, err := parseChunk(tr)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", hdr.Name, err)
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
