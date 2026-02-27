package rpm

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ulikunitz/xz"
)

// repoDef defines a single RPM repository to ingest.
type repoDef struct {
	Name       string // e.g. "rhel9-baseos"
	RepomdURL  string // URL to repomd.xml
	Compressed string // "gz" or "xz"
}

var repos = []repoDef{
	{
		Name:       "rhel9-baseos",
		RepomdURL:  "https://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/repodata/repomd.xml",
		Compressed: "gz",
	},
	{
		Name:       "rhel9-appstream",
		RepomdURL:  "https://mirror.stream.centos.org/9-stream/AppStream/x86_64/os/repodata/repomd.xml",
		Compressed: "gz",
	},
	{
		Name:       "epel9",
		RepomdURL:  "https://dl.fedoraproject.org/pub/epel/9/Everything/x86_64/repodata/repomd.xml",
		Compressed: "xz",
	},
}

// Client downloads and parses RPM repository metadata.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new RPM client.
func NewClient(httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient}
}

// RPMPackageData holds a parsed RPM package entry.
type RPMPackageData struct {
	PackageName string
	Repo        string
	Arch        string
	Version     string
	RPMGroup    string
	Summary     string
	URL         string
}

// Download fetches all configured RPM repos and parses their primary.xml metadata.
func (c *Client) Download(heartbeat func()) ([]RPMPackageData, error) {
	var all []RPMPackageData

	for _, repo := range repos {
		packages, err := c.downloadRepo(repo)
		if err != nil {
			return nil, fmt.Errorf("download %s: %w", repo.Name, err)
		}
		all = append(all, packages...)
		heartbeat()
	}

	return all, nil
}

func (c *Client) downloadRepo(repo repoDef) ([]RPMPackageData, error) {
	// Step 1: Fetch repomd.xml to find primary.xml location
	primaryURL, err := c.discoverPrimaryURL(repo)
	if err != nil {
		return nil, fmt.Errorf("discover primary URL: %w", err)
	}

	// Step 2: Download primary.xml.gz/.xz to temp
	resp, err := c.httpClient.Get(primaryURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", primaryURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", primaryURL, resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "rpm-primary-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return nil, fmt.Errorf("download to temp: %w", err)
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek temp file: %w", err)
	}

	// Step 3: Decompress
	var reader io.Reader
	switch repo.Compressed {
	case "gz":
		gz, err := gzip.NewReader(tmpFile)
		if err != nil {
			return nil, fmt.Errorf("gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	case "xz":
		xzReader, err := xz.NewReader(tmpFile)
		if err != nil {
			return nil, fmt.Errorf("xz reader: %w", err)
		}
		reader = xzReader
	default:
		return nil, fmt.Errorf("unsupported compression: %s", repo.Compressed)
	}

	// Step 4: Parse XML
	return parsePrimaryXML(reader, repo.Name)
}

// repomd XML structures
// Namespace: http://linux.duke.edu/metadata/repo
type repomdXML struct {
	XMLName xml.Name     `xml:"repomd"`
	Data    []repomdData `xml:"data"`
}

type repomdData struct {
	Type     string         `xml:"type,attr"`
	Location repomdLocation `xml:"location"`
}

type repomdLocation struct {
	Href string `xml:"href,attr"`
}

func (c *Client) discoverPrimaryURL(repo repoDef) (string, error) {
	resp, err := c.httpClient.Get(repo.RepomdURL)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", repo.RepomdURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: status %d", repo.RepomdURL, resp.StatusCode)
	}

	var repomd repomdXML
	if err := xml.NewDecoder(resp.Body).Decode(&repomd); err != nil {
		return "", fmt.Errorf("decode repomd.xml: %w", err)
	}

	for _, d := range repomd.Data {
		if d.Type == "primary" {
			// Location href is relative to the repo base URL
			baseURL := strings.TrimSuffix(repo.RepomdURL, "repodata/repomd.xml")
			return baseURL + d.Location.Href, nil
		}
	}

	return "", fmt.Errorf("no primary data found in repomd.xml")
}

// primary.xml structures
// Namespace: http://linux.duke.edu/metadata/common (default)
// RPM namespace: http://linux.duke.edu/metadata/rpm
type primaryMetadata struct {
	XMLName  xml.Name         `xml:"metadata"`
	Packages []primaryPackage `xml:"package"`
}

type primaryPackage struct {
	Type    string         `xml:"type,attr"`
	Name    string         `xml:"name"`
	Arch    string         `xml:"arch"`
	Version primaryVersion `xml:"version"`
	Summary string         `xml:"summary"`
	URL     string         `xml:"url"`
	Format  primaryFormat  `xml:"format"`
}

type primaryVersion struct {
	Epoch string `xml:"epoch,attr"`
	Ver   string `xml:"ver,attr"`
	Rel   string `xml:"rel,attr"`
}

type primaryFormat struct {
	Group string `xml:"http://linux.duke.edu/metadata/rpm group"`
}

func parsePrimaryXML(r io.Reader, repoName string) ([]RPMPackageData, error) {
	var metadata primaryMetadata
	if err := xml.NewDecoder(r).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("decode primary.xml: %w", err)
	}

	var result []RPMPackageData
	for _, pkg := range metadata.Packages {
		if pkg.Type != "rpm" {
			continue
		}

		version := pkg.Version.Ver
		if pkg.Version.Rel != "" {
			version += "-" + pkg.Version.Rel
		}

		result = append(result, RPMPackageData{
			PackageName: pkg.Name,
			Repo:        repoName,
			Arch:        pkg.Arch,
			Version:     version,
			RPMGroup:    pkg.Format.Group,
			Summary:     pkg.Summary,
			URL:         pkg.URL,
		})
	}

	return result, nil
}
