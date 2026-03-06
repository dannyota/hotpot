package eol

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"danny.vn/hotpot/pkg/base/httputil"
)

const tarballURL = "https://github.com/endoflife-date/endoflife.date/archive/refs/heads/master.tar.gz"

// Client downloads and parses the endoflife.date repository.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new EOL client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// ProductData holds a parsed product with its release cycles.
type ProductData struct {
	Slug        string
	Name        string
	Category    string
	Tags        []string
	Identifiers []IdentifierData
	Cycles      []CycleData
}

// IdentifierData holds a parsed product identifier (purl, repology, cpe).
type IdentifierData struct {
	Type  string // "purl", "repology", "cpe"
	Value string
}

// CycleData holds a parsed release cycle.
type CycleData struct {
	Cycle           string
	ReleaseDate     *time.Time
	EOAS            *time.Time
	EOL             *time.Time
	EOES            *time.Time
	Latest          string
	LatestReleaseDate *time.Time
	LTS             *time.Time
}

// yamlFrontmatter maps the YAML structure in products/*.md files.
type yamlFrontmatter struct {
	Title       string              `yaml:"title"`
	Category    string              `yaml:"category"`
	Permalink   string              `yaml:"permalink"`
	Tags        yamlTags            `yaml:"tags"`
	Identifiers []map[string]string `yaml:"identifiers"`
	Releases    []yamlRelease       `yaml:"releases"`
}

// yamlTags handles the tags field which can be a single string or a list.
type yamlTags []string

func (t *yamlTags) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try as a list first.
	var list []string
	if err := unmarshal(&list); err == nil {
		*t = list
		return nil
	}
	// Fall back to a single string.
	var s string
	if err := unmarshal(&s); err == nil {
		if s != "" {
			*t = []string{s}
		}
		return nil
	}
	return nil
}

type yamlRelease struct {
	ReleaseCycle      string      `yaml:"releaseCycle"`
	ReleaseDate       interface{} `yaml:"releaseDate"`
	EOAS              interface{} `yaml:"eoas"`
	EOL               interface{} `yaml:"eol"`
	EOES              interface{} `yaml:"eoes"`
	Latest            string      `yaml:"latest"`
	LatestReleaseDate interface{} `yaml:"latestReleaseDate"`
	LTS               interface{} `yaml:"lts"`
}

// Download fetches the endoflife.date GitHub tarball and parses all product YAML files.
func (c *Client) Download(heartbeat func(string)) ([]ProductData, error) {
	resp, err := c.httpClient.Get(tarballURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", tarballURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", tarballURL, resp.StatusCode)
	}

	// Download to temp file
	tmpFile, err := os.CreateTemp("", "endoflife-*.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	body := httputil.NewProgressReader(resp.Body, resp.ContentLength, "endoflife.date", 5*time.Second, heartbeat)
	if _, err := io.Copy(tmpFile, body); err != nil {
		return nil, fmt.Errorf("download to temp: %w", err)
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek temp file: %w", err)
	}

	heartbeat("parsing endoflife.date products")

	// Parse tar.gz
	gz, err := gzip.NewReader(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var products []ProductData

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar next: %w", err)
		}

		// Only process products/*.md files.
		// Tarball prefix is like "endoflife.date-master/products/rhel.md"
		name := hdr.Name
		dir := filepath.Dir(name)
		base := filepath.Base(dir)

		if base != "products" || !strings.HasSuffix(name, ".md") {
			continue
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}

		product, err := parseProductFile(filepath.Base(name), content)
		if err != nil {
			slog.Warn("Skipping product file", "file", name, "error", err)
			continue
		}

		products = append(products, *product)

		if len(products)%50 == 0 {
			heartbeat(fmt.Sprintf("parsed %d products", len(products)))
		}
	}

	slog.Info("Parsed endoflife.date products", "count", len(products))
	return products, nil
}

// parseProductFile extracts YAML frontmatter from a product .md file.
func parseProductFile(filename string, content []byte) (*ProductData, error) {
	// Extract YAML frontmatter between --- delimiters
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("extract frontmatter: %w", err)
	}

	var fm yamlFrontmatter
	if err := yaml.Unmarshal(frontmatter, &fm); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}

	if fm.Title == "" {
		return nil, fmt.Errorf("missing title")
	}

	// Derive slug from filename (rhel.md → rhel)
	slug := strings.TrimSuffix(filename, ".md")

	product := &ProductData{
		Slug:     slug,
		Name:     fm.Title,
		Category: fm.Category,
		Tags:     []string(fm.Tags),
	}

	// Parse identifiers: each entry is a single-key map like {"purl": "pkg:deb/ubuntu/nginx"}.
	for _, m := range fm.Identifiers {
		for idType, idValue := range m {
			product.Identifiers = append(product.Identifiers, IdentifierData{
				Type:  idType,
				Value: idValue,
			})
		}
	}

	for _, rel := range fm.Releases {
		if rel.ReleaseCycle == "" {
			continue
		}
		cycle := CycleData{
			Cycle:             rel.ReleaseCycle,
			ReleaseDate:       parseFlexDate(rel.ReleaseDate),
			EOAS:              parseFlexDate(rel.EOAS),
			EOL:               parseFlexDate(rel.EOL),
			EOES:              parseFlexDate(rel.EOES),
			Latest:            rel.Latest,
			LatestReleaseDate: parseFlexDate(rel.LatestReleaseDate),
			LTS:               parseFlexDate(rel.LTS),
		}
		product.Cycles = append(product.Cycles, cycle)
	}

	return product, nil
}

// extractFrontmatter extracts content between --- delimiters.
func extractFrontmatter(content []byte) ([]byte, error) {
	// Find first ---
	idx := bytes.Index(content, []byte("---"))
	if idx < 0 {
		return nil, fmt.Errorf("no opening --- found")
	}
	rest := content[idx+3:]

	// Find second ---
	idx2 := bytes.Index(rest, []byte("---"))
	if idx2 < 0 {
		return nil, fmt.Errorf("no closing --- found")
	}

	return rest[:idx2], nil
}

// parseFlexDate parses a date field that can be a date string or boolean.
// Returns nil for booleans or unparseable values.
func parseFlexDate(v interface{}) *time.Time {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case string:
		t, err := time.Parse("2006-01-02", val)
		if err != nil {
			return nil
		}
		return &t
	case time.Time:
		return &val
	default:
		// bool, int, or other — no date to store
		return nil
	}
}
