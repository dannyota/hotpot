// Package matchrule loads configurable detection rules from the config schema.
// Both normalize and detect layers use this to avoid hardcoded pattern lists.
package matchrule

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Service loads and caches match rules from config tables.
// Rules are loaded lazily on first use and refreshed periodically.
type Service struct {
	db          *sql.DB
	mu          sync.RWMutex
	rules       *RuleSet
	lastLoaded  time.Time
	refreshEvery time.Duration
}

// NewService creates a matchrule service. Rules are loaded lazily on first Get().
func NewService(db *sql.DB) *Service {
	return &Service{
		db:           db,
		refreshEvery: 5 * time.Minute,
	}
}

// Get returns the cached RuleSet, loading or refreshing from DB if needed.
func (s *Service) Get(ctx context.Context) (*RuleSet, error) {
	s.mu.RLock()
	if s.rules != nil && time.Since(s.lastLoaded) < s.refreshEvery {
		defer s.mu.RUnlock()
		return s.rules, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock.
	if s.rules != nil && time.Since(s.lastLoaded) < s.refreshEvery {
		return s.rules, nil
	}

	rules, err := load(ctx, s.db)
	if err != nil {
		// If we have stale rules, return them rather than failing.
		if s.rules != nil {
			return s.rules, nil
		}
		return nil, err
	}
	s.rules = rules
	s.lastLoaded = time.Now()
	return rules, nil
}

// URIAttackPattern holds a single URI attack detection pattern.
type URIAttackPattern struct {
	PatternType string // "lfi", "sqli", "rce", "xss", "ssrf"
	MatchMode   string // "substring", "regex"
	Pattern     string
}

// AuthEndpointPattern holds a single auth endpoint detection pattern.
type AuthEndpointPattern struct {
	PatternType string // "login", "otp", "password_reset", "register", "admin", "token"
	MatchMode   string // "keyword", "substring", "regex"
	Pattern     string
}

// RuleSet holds all active rules loaded from config tables.
type RuleSet struct {
	HostingDomains    map[string]bool
	HostingKeywords   []string
	ScannerKeywords   []string
	LibraryUAs        map[string]bool
	URIAttackPatterns []URIAttackPattern
	AuthPatterns      []AuthEndpointPattern
}

// load reads all active rules from config tables.
func load(ctx context.Context, db *sql.DB) (*RuleSet, error) {
	rs := &RuleSet{
		HostingDomains: make(map[string]bool),
		LibraryUAs:     make(map[string]bool),
	}

	// Load hosting indicators.
	rows, err := db.QueryContext(ctx,
		`SELECT indicator_type, value FROM config.hosting_indicators WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("load hosting indicators: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var typ, val string
		if err := rows.Scan(&typ, &val); err != nil {
			return nil, fmt.Errorf("scan hosting indicator: %w", err)
		}
		switch typ {
		case "domain":
			rs.HostingDomains[val] = true
		case "keyword":
			rs.HostingKeywords = append(rs.HostingKeywords, val)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hosting indicators: %w", err)
	}

	// Load scanner patterns.
	rows2, err := db.QueryContext(ctx,
		`SELECT keyword FROM config.scanner_patterns WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("load scanner patterns: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var kw string
		if err := rows2.Scan(&kw); err != nil {
			return nil, fmt.Errorf("scan scanner pattern: %w", err)
		}
		rs.ScannerKeywords = append(rs.ScannerKeywords, kw)
	}
	if err := rows2.Err(); err != nil {
		return nil, fmt.Errorf("iterate scanner patterns: %w", err)
	}

	// Load library UA families.
	rows3, err := db.QueryContext(ctx,
		`SELECT family FROM config.library_uas WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("load library UAs: %w", err)
	}
	defer rows3.Close()
	for rows3.Next() {
		var fam string
		if err := rows3.Scan(&fam); err != nil {
			return nil, fmt.Errorf("scan library UA: %w", err)
		}
		rs.LibraryUAs[fam] = true
	}
	if err := rows3.Err(); err != nil {
		return nil, fmt.Errorf("iterate library UAs: %w", err)
	}

	// Load URI attack patterns.
	rows4, err := db.QueryContext(ctx,
		`SELECT pattern_type, match_mode, pattern FROM config.uri_attack_patterns WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("load uri attack patterns: %w", err)
	}
	defer rows4.Close()
	for rows4.Next() {
		var p URIAttackPattern
		if err := rows4.Scan(&p.PatternType, &p.MatchMode, &p.Pattern); err != nil {
			return nil, fmt.Errorf("scan uri attack pattern: %w", err)
		}
		rs.URIAttackPatterns = append(rs.URIAttackPatterns, p)
	}
	if err := rows4.Err(); err != nil {
		return nil, fmt.Errorf("iterate uri attack patterns: %w", err)
	}

	// Load auth endpoint patterns.
	rows5, err := db.QueryContext(ctx,
		`SELECT pattern_type, match_mode, pattern FROM config.auth_endpoint_patterns WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("load auth endpoint patterns: %w", err)
	}
	defer rows5.Close()
	for rows5.Next() {
		var p AuthEndpointPattern
		if err := rows5.Scan(&p.PatternType, &p.MatchMode, &p.Pattern); err != nil {
			return nil, fmt.Errorf("scan auth endpoint pattern: %w", err)
		}
		rs.AuthPatterns = append(rs.AuthPatterns, p)
	}
	if err := rows5.Err(); err != nil {
		return nil, fmt.Errorf("iterate auth endpoint patterns: %w", err)
	}

	return rs, nil
}

// IsHostingDomain returns true if the AS domain belongs to a known hosting/cloud provider.
func (rs *RuleSet) IsHostingDomain(domain, asnType string) bool {
	if asnType == "hosting" {
		return true
	}
	if domain == "" {
		return false
	}
	if rs.HostingDomains[domain] {
		return true
	}
	lower := strings.ToLower(domain)
	for _, kw := range rs.HostingKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// IsScannerUA checks if a user agent matches a known scanner tool.
func (rs *RuleSet) IsScannerUA(ua string) bool {
	lower := strings.ToLower(ua)
	for _, kw := range rs.ScannerKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// ScannerLikeClause builds a SQL WHERE clause fragment matching scanner keywords.
func (rs *RuleSet) ScannerLikeClause() string {
	if len(rs.ScannerKeywords) == 0 {
		return "false"
	}
	parts := make([]string, len(rs.ScannerKeywords))
	for i, kw := range rs.ScannerKeywords {
		parts[i] = fmt.Sprintf("lower(user_agent) LIKE '%%%s%%'", kw)
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// IsLibraryUA returns true if the UA family is a known automated/library client.
func (rs *RuleSet) IsLibraryUA(family string) bool {
	return rs.LibraryUAs[family]
}

// AuthPatternClause builds a SQL WHERE clause fragment matching auth endpoint
// patterns for the given pattern type. All modes use ILIKE for case-insensitive
// substring matching; regex patterns use ~* (case-insensitive regex).
func (rs *RuleSet) AuthPatternClause(patternType string) string {
	var parts []string
	for _, p := range rs.AuthPatterns {
		if p.PatternType != patternType {
			continue
		}
		escaped := strings.ReplaceAll(p.Pattern, "'", "''")
		switch p.MatchMode {
		case "keyword", "substring":
			parts = append(parts, fmt.Sprintf("t.uri ILIKE '%%%s%%'", escaped))
		case "regex":
			parts = append(parts, fmt.Sprintf("t.uri ~* '%s'", escaped))
		}
	}
	if len(parts) == 0 {
		return "false"
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

// URIAttackClause builds a SQL WHERE clause fragment matching URI attack
// patterns for the given pattern type. Substring patterns use ILIKE,
// regex patterns use ~* (case-insensitive).
func (rs *RuleSet) URIAttackClause(patternType string) string {
	var parts []string
	for _, p := range rs.URIAttackPatterns {
		if p.PatternType != patternType {
			continue
		}
		// Escape single quotes in pattern for SQL safety.
		escaped := strings.ReplaceAll(p.Pattern, "'", "''")
		switch p.MatchMode {
		case "substring":
			parts = append(parts, fmt.Sprintf("uri ILIKE '%%%s%%'", escaped))
		case "regex":
			parts = append(parts, fmt.Sprintf("uri ~* '%s'", escaped))
		}
	}
	if len(parts) == 0 {
		return "false"
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}
