package geoip

import (
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"

	"github.com/oschwald/maxminddb-golang/v2"
)

// Result holds the combined GeoIP + ASN lookup result for an IP address.
type Result struct {
	CountryCode string
	CountryName string
	City        string
	Subdivision string
	Latitude    float64
	Longitude   float64
	ASN         int
	OrgName     string
	ASDomain    string // Domain of the AS (e.g., "google.com", "cloudflare.com") — from IPinfo
	ASNType     string // AS type: isp, hosting, business, education — from IPinfo paid only
	IsInternal  bool
}

// Lookup wraps dual mmdb readers (city + ASN) with safe reload.
type Lookup struct {
	mu         sync.RWMutex
	cityDB     *maxminddb.Reader
	asnDB      *maxminddb.Reader
	cityPath   string
	asnPath    string
	asnIPInfo  bool // true if ASN db is IPinfo format (has "domain" field)
	asnHasType bool // true if ASN db has "type" field (IPinfo paid)
}

// cityRecord maps the mmdb city/location fields.
type cityRecord struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Subdivisions []struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

// asnRecordDBIP maps DB-IP / MaxMind ASN fields.
type asnRecordDBIP struct {
	AutonomousSystemNumber       int    `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

// asnRecordIPInfo maps IPinfo ASN fields.
// Free: asn, name, domain. Paid adds: type, country.
type asnRecordIPInfo struct {
	ASN    string `maxminddb:"asn"`    // "AS15169"
	Name   string `maxminddb:"name"`   // "Google LLC"
	Domain string `maxminddb:"domain"` // "google.com"
	Type   string `maxminddb:"type"`   // "isp", "hosting", "business", "education" (paid only)
}

// NewLookup opens both .mmdb files. Missing files are OK — graceful degradation.
// Auto-detects IPinfo vs DB-IP ASN format from database metadata.
func NewLookup(cityPath, asnPath string) *Lookup {
	l := &Lookup{
		cityPath: cityPath,
		asnPath:  asnPath,
	}
	l.cityDB = openDB(cityPath)
	l.asnDB = openDB(asnPath)
	l.detectASNFormat()
	return l
}

func openDB(path string) *maxminddb.Reader {
	if path == "" {
		return nil
	}
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil
	}
	return db
}

// LookupIP returns all available geo + ASN info for an IP.
// For internal IPs (RFC1918/loopback), returns Result{IsInternal: true} with no geo data.
func (l *Lookup) LookupIP(ipStr string) Result {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Result{}
	}

	if isInternalIP(ip) {
		return Result{IsInternal: true}
	}

	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return Result{}
	}
	addr = addr.Unmap()

	var r Result

	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.cityDB != nil {
		var rec cityRecord
		if err := l.cityDB.Lookup(addr).Decode(&rec); err == nil {
			r.CountryCode = rec.Country.ISOCode
			r.CountryName = rec.Country.Names["en"]
			r.City = rec.City.Names["en"]
			if len(rec.Subdivisions) > 0 {
				r.Subdivision = rec.Subdivisions[0].Names["en"]
			}
			r.Latitude = rec.Location.Latitude
			r.Longitude = rec.Location.Longitude
		}
	}

	if l.asnDB != nil {
		if l.asnIPInfo {
			var rec asnRecordIPInfo
			if err := l.asnDB.Lookup(addr).Decode(&rec); err == nil {
				r.ASN = parseIPInfoASN(rec.ASN)
				r.OrgName = rec.Name
				r.ASDomain = rec.Domain
				if l.asnHasType {
					r.ASNType = rec.Type
				}
			}
		} else {
			var rec asnRecordDBIP
			if err := l.asnDB.Lookup(addr).Decode(&rec); err == nil {
				r.ASN = rec.AutonomousSystemNumber
				r.OrgName = rec.AutonomousSystemOrganization
			}
		}
	}

	return r
}

// detectASNFormat checks the ASN database metadata to determine format.
func (l *Lookup) detectASNFormat() {
	if l.asnDB == nil {
		return
	}
	dbType := l.asnDB.Metadata.DatabaseType
	// IPinfo databases use "ipinfo" prefix in their database type.
	if strings.Contains(strings.ToLower(dbType), "ipinfo") {
		l.asnIPInfo = true
		// Probe a known IP to check if "type" field is present (paid version).
		addr, _ := netip.ParseAddr("8.8.8.8")
		var rec asnRecordIPInfo
		if err := l.asnDB.Lookup(addr).Decode(&rec); err == nil {
			l.asnHasType = rec.Type != ""
		}
	}
}

// parseIPInfoASN converts IPinfo ASN string "AS15169" to int 15169.
func parseIPInfoASN(s string) int {
	s = strings.TrimPrefix(s, "AS")
	n, _ := strconv.Atoi(s)
	return n
}

// Reload reopens both files. Called at start of each normalize run
// to pick up updates from the download workflow.
func (l *Lookup) Reload() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cityDB != nil {
		l.cityDB.Close()
	}
	if l.asnDB != nil {
		l.asnDB.Close()
	}

	l.cityDB = openDB(l.cityPath)
	l.asnDB = openDB(l.asnPath)
	l.detectASNFormat()
	return nil
}

// Close closes both mmdb readers.
func (l *Lookup) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cityDB != nil {
		l.cityDB.Close()
		l.cityDB = nil
	}
	if l.asnDB != nil {
		l.asnDB.Close()
		l.asnDB = nil
	}
	return nil
}

// Internal IP ranges.
var internalNets = func() []*net.IPNet {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}
	var nets []*net.IPNet
	for _, cidr := range cidrs {
		_, n, _ := net.ParseCIDR(cidr)
		nets = append(nets, n)
	}
	return nets
}()

func isInternalIP(ip net.IP) bool {
	for _, n := range internalNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
