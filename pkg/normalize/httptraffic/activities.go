package httptraffic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/geoip"
	"danny.vn/hotpot/pkg/base/matchrule"
	enthttptraffic "danny.vn/hotpot/pkg/storage/ent/httptraffic"
)

// Activities holds dependencies for HTTP traffic normalize activities.
type Activities struct {
	configService *config.Service
	entClient     *enthttptraffic.Client
	db            *sql.DB
	geoip         *geoip.Lookup
	matchRules    *matchrule.Service

	cachedMatcher   *PathMatcher
	cachedMatcherAt time.Time
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *enthttptraffic.Client, db *sql.DB, geoipLookup *geoip.Lookup, matchRules *matchrule.Service) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		db:            db,
		geoip:         geoipLookup,
		matchRules:    matchRules,
	}
}

const pathMatcherCacheTTL = 5 * time.Minute

// getPathMatcher returns a cached PathMatcher, rebuilding it if the cache is
// older than 5 minutes.
func (a *Activities) getPathMatcher(ctx context.Context) (*PathMatcher, error) {
	if a.cachedMatcher != nil && time.Since(a.cachedMatcherAt) < pathMatcherCacheTTL {
		return a.cachedMatcher, nil
	}
	endpoints, err := loadEndpoints(ctx, a.db)
	if err != nil {
		return nil, err
	}
	a.cachedMatcher = NewPathMatcher(endpoints)
	a.cachedMatcherAt = time.Now()
	return a.cachedMatcher, nil
}

// Activity function references for Temporal registration.
var NormalizeTrafficActivity = (*Activities).NormalizeTraffic

// NormalizeTrafficParams holds input for the NormalizeTraffic activity.
type NormalizeTrafficParams struct {
	// SinceMinutes limits how far back to look for new bronze rows.
	// Default: 30 minutes.
	SinceMinutes int

	// Since is the computed cutoff time. Set once by the workflow to avoid
	// drift between activities that run at different wall-clock times.
	Since time.Time
}

// NormalizeTrafficResult holds normalization statistics.
type NormalizeTrafficResult struct {
	Processed int
	Mapped    int
	Unmapped  int
}

// NormalizeTraffic reads bronze traffic counts, matches against endpoints, writes to silver.
func (a *Activities) NormalizeTraffic(ctx context.Context, params NormalizeTrafficParams) (*NormalizeTrafficResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing HTTP traffic")

	since := params.Since

	// Step 1: Load endpoints and build path matcher.
	pm, err := a.getPathMatcher(ctx)
	if err != nil {
		return nil, err
	}

	// Step 2: Read new bronze rows.
	bronzeRows, err := loadBronzeTraffic(ctx, a.db, since)
	if err != nil {
		return nil, err
	}
	logger.Info("Loaded bronze traffic rows", "count", len(bronzeRows))

	// Step 3: Compute unique_client_count from bronze IP table.
	uniqueClients, err := loadUniqueClientCounts(ctx, a.db, since)
	if err != nil {
		return nil, err
	}

	// Step 4: Load match rules and compute is_scanner_detected from bronze UA table.
	rules, err := a.matchRules.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load match rules: %w", err)
	}
	scannerKeys, err := loadScannerKeys(ctx, a.db, since, rules)
	if err != nil {
		return nil, err
	}

	// Step 4b: Detect URI attack patterns from bronze HTTP counts.
	uriAttackKeys, err := loadURIAttackKeys(ctx, a.db, since, rules)
	if err != nil {
		return nil, err
	}

	// Step 5: Match and write to silver.
	now := time.Now()
	var mapped, unmapped int

	for _, row := range bronzeRows {
		ep := pm.Match(row.uri)

		resourceID := fmt.Sprintf("%s:%s:%s:%s:%d",
			row.sourceID,
			row.windowStart.Format(time.RFC3339),
			row.uri, row.method, row.statusCode)

		tk := trafficKey{row.sourceID, row.windowStart, row.uri, row.method}

		create := a.entClient.SilverHttptrafficTraffic5m.Create().
			SetID(resourceID).
			SetSourceID(row.sourceID).
			SetWindowStart(row.windowStart).
			SetWindowEnd(row.windowEnd).
			SetURI(row.uri).
			SetStatusCode(row.statusCode).
			SetRequestCount(row.requestCount).
			SetTotalBodyBytesSent(row.totalBodyBytesSent).
			SetUniqueClientCount(uniqueClients[tk]).
			SetCollectedAt(row.collectedAt).
			SetFirstCollectedAt(row.collectedAt).
			SetNormalizedAt(now)

		if row.method != "" {
			create.SetMethod(row.method)
		}
		if row.totalRequestTime > 0 && row.requestCount > 0 {
			create.SetAvgRequestTime(row.totalRequestTime / float64(row.requestCount))
		}
		if row.maxRequestTime > 0 {
			create.SetMaxRequestTime(row.maxRequestTime)
		}
		if scannerKeys[tk] {
			create.SetIsScannerDetected(true)
		}
		if uriAttackKeys["lfi"][tk] {
			create.SetIsLfiDetected(true)
		}
		if uriAttackKeys["sqli"][tk] {
			create.SetIsSqliDetected(true)
		}
		if uriAttackKeys["rce"][tk] {
			create.SetIsRceDetected(true)
		}
		if uriAttackKeys["xss"][tk] {
			create.SetIsXSSDetected(true)
		}
		if uriAttackKeys["ssrf"][tk] {
			create.SetIsSsrfDetected(true)
		}

		if ep != nil {
			create.SetEndpointID(ep.ID).
				SetIsMapped(true)
			if ep.AccessLevel != "" {
				create.SetAccessLevel(ep.AccessLevel)
			}
			if ep.Service != "" {
				create.SetService(ep.Service)
			}
			if row.method != "" && len(ep.Methods) > 0 && !methodAllowed(row.method, ep.Methods) {
				create.SetIsMethodMismatch(true)
			}
			mapped++
		} else {
			create.SetIsMapped(false)
			unmapped++
		}

		if err := create.Exec(ctx); err != nil {
			if enthttptraffic.IsConstraintError(err) {
				continue
			}
			return nil, fmt.Errorf("create traffic_5m %s: %w", resourceID, err)
		}
	}

	logger.Info("HTTP traffic normalization complete",
		"processed", len(bronzeRows),
		"mapped", mapped,
		"unmapped", unmapped)

	return &NormalizeTrafficResult{
		Processed: len(bronzeRows),
		Mapped:    mapped,
		Unmapped:  unmapped,
	}, nil
}

type bronzeTrafficRow struct {
	sourceID           string
	windowStart        time.Time
	windowEnd          time.Time
	uri                string
	method             string
	statusCode         int
	requestCount       int64
	totalRequestTime   float64
	maxRequestTime     float64
	totalBodyBytesSent int64
	collectedAt        time.Time
}

// trafficKey is used for cross-table lookups.
type trafficKey struct {
	sourceID    string
	windowStart time.Time
	uri         string
	method      string
}

func loadEndpoints(ctx context.Context, db *sql.DB) ([]MatchEndpoint, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, uri_pattern,
			COALESCE(methods, '[]'),
			COALESCE(service, ''),
			COALESCE(access_level, '')
		FROM silver.inventory_api_endpoints
		WHERE is_active = true`)
	if err != nil {
		return nil, fmt.Errorf("query inventory_api_endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []MatchEndpoint
	for rows.Next() {
		var id, uriPattern, methodsRaw, service, accessLevel string
		if err := rows.Scan(&id, &uriPattern, &methodsRaw, &service, &accessLevel); err != nil {
			return nil, fmt.Errorf("scan endpoint: %w", err)
		}
		var methods []string
		if err := json.Unmarshal([]byte(methodsRaw), &methods); err != nil {
			methods = nil
		}
		endpoints = append(endpoints, MatchEndpoint{
			ID:          id,
			URIPattern:  uriPattern,
			Methods:     methods,
			Service:     service,
			AccessLevel: accessLevel,
		})
	}
	return endpoints, rows.Err()
}

func loadBronzeTraffic(ctx context.Context, db *sql.DB, since time.Time) ([]bronzeTrafficRow, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT source_id, window_start, window_end, uri,
			COALESCE(method, ''), status_code, request_count,
			COALESCE(total_request_time, 0), COALESCE(max_request_time, 0),
			total_body_bytes_sent, collected_at
		FROM bronze.accesslog_http_counts
		WHERE collected_at >= $1
		ORDER BY window_start`, since)
	if err != nil {
		return nil, fmt.Errorf("query accesslog_http_counts: %w", err)
	}
	defer rows.Close()

	var result []bronzeTrafficRow
	for rows.Next() {
		var r bronzeTrafficRow
		if err := rows.Scan(&r.sourceID, &r.windowStart, &r.windowEnd, &r.uri,
			&r.method, &r.statusCode, &r.requestCount,
			&r.totalRequestTime, &r.maxRequestTime,
			&r.totalBodyBytesSent, &r.collectedAt); err != nil {
			return nil, fmt.Errorf("scan bronze traffic: %w", err)
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// loadUniqueClientCounts computes unique client IP counts from bronze IP table.
func loadUniqueClientCounts(ctx context.Context, db *sql.DB, since time.Time) (map[trafficKey]int, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT source_id, window_start, uri, COALESCE(method, ''),
			COUNT(DISTINCT client_ip)
		FROM bronze.accesslog_client_ips
		WHERE collected_at >= $1
		GROUP BY source_id, window_start, uri, method`, since)
	if err != nil {
		return nil, fmt.Errorf("query unique client counts: %w", err)
	}
	defer rows.Close()

	result := make(map[trafficKey]int)
	for rows.Next() {
		var k trafficKey
		var count int
		if err := rows.Scan(&k.sourceID, &k.windowStart, &k.uri, &k.method, &count); err != nil {
			return nil, fmt.Errorf("scan unique client count: %w", err)
		}
		result[k] = count
	}
	return result, rows.Err()
}

// loadScannerKeys finds traffic keys that have scanner UAs in the bronze UA table.
func loadScannerKeys(ctx context.Context, db *sql.DB, since time.Time, rules *matchrule.RuleSet) (map[trafficKey]bool, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT source_id, window_start, uri, COALESCE(method, '')
		FROM bronze.accesslog_user_agents
		WHERE collected_at >= $1
			AND %s`, rules.ScannerLikeClause())
	rows, err := db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("query scanner keys: %w", err)
	}
	defer rows.Close()

	result := make(map[trafficKey]bool)
	for rows.Next() {
		var k trafficKey
		if err := rows.Scan(&k.sourceID, &k.windowStart, &k.uri, &k.method); err != nil {
			return nil, fmt.Errorf("scan scanner key: %w", err)
		}
		result[k] = true
	}
	return result, rows.Err()
}

// loadURIAttackKeys finds traffic keys whose URIs match attack patterns.
// Returns a map keyed by pattern type ("lfi", "sqli", "rce", "xss", "ssrf").
func loadURIAttackKeys(ctx context.Context, db *sql.DB, since time.Time, rules *matchrule.RuleSet) (map[string]map[trafficKey]bool, error) {
	result := make(map[string]map[trafficKey]bool)
	types := []string{"lfi", "sqli", "rce", "xss", "ssrf"}

	for _, pt := range types {
		clause := rules.URIAttackClause(pt)
		if clause == "false" {
			continue
		}
		query := fmt.Sprintf(`
			SELECT DISTINCT source_id, window_start, uri, COALESCE(method, '')
			FROM bronze.accesslog_http_counts
			WHERE collected_at >= $1
				AND %s`, clause)
		rows, err := db.QueryContext(ctx, query, since)
		if err != nil {
			return nil, fmt.Errorf("query %s attack keys: %w", pt, err)
		}

		keys := make(map[trafficKey]bool)
		for rows.Next() {
			var k trafficKey
			if err := rows.Scan(&k.sourceID, &k.windowStart, &k.uri, &k.method); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan %s attack key: %w", pt, err)
			}
			keys[k] = true
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate %s attack keys: %w", pt, err)
		}
		if len(keys) > 0 {
			result[pt] = keys
		}
	}
	return result, nil
}

// methodAllowed checks if method is in the allowed list (case-insensitive).
func methodAllowed(method string, allowed []string) bool {
	for _, m := range allowed {
		if strings.EqualFold(method, m) {
			return true
		}
	}
	return false
}
