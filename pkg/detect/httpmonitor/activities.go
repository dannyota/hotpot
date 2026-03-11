package httpmonitor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/matchrule"
	enthttpmonitor "danny.vn/hotpot/pkg/storage/ent/httpmonitor"
)

// Activities holds dependencies for httpmonitor detection activities.
type Activities struct {
	configService *config.Service
	entClient     *enthttpmonitor.Client
	db            *sql.DB
	matchRules    *matchrule.Service

	cachedRules   *Rules
	cachedRulesAt time.Time
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *enthttpmonitor.Client, db *sql.DB, matchRules *matchrule.Service) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		db:            db,
		matchRules:    matchRules,
	}
}

// getRules returns cached rules if less than 5 minutes old, otherwise reloads.
func (a *Activities) getRules(ctx context.Context) (*Rules, error) {
	if a.cachedRules != nil && time.Since(a.cachedRulesAt) < 5*time.Minute {
		return a.cachedRules, nil
	}
	r, err := loadRules(ctx, a.db)
	if err != nil {
		return nil, err
	}
	a.cachedRules = r
	a.cachedRulesAt = time.Now()
	return r, nil
}

// Activity function references for Temporal registration.
var (
	DetectRateAnomaliesActivity        = (*Activities).DetectRateAnomalies
	DetectErrorBurstsActivity          = (*Activities).DetectErrorBursts
	DetectSuspiciousPatternsActivity   = (*Activities).DetectSuspiciousPatterns
	DetectMethodMismatchActivity       = (*Activities).DetectMethodMismatch
	DetectUserAgentAnomaliesActivity   = (*Activities).DetectUserAgentAnomalies
	DetectClientIPAnomaliesActivity    = (*Activities).DetectClientIPAnomalies
	DetectASNAnomaliesActivity         = (*Activities).DetectASNAnomalies
	DetectNewEndpointsActivity         = (*Activities).DetectNewEndpoints
	DetectAuthAnomaliesActivity        = (*Activities).DetectAuthAnomalies
	CleanupStaleActivity               = (*Activities).CleanupStale
)

// --- Activity 1: DetectRateAnomalies ---

// DetectRateAnomaliesResult holds output.
type DetectRateAnomaliesResult struct {
	Spikes             int
	Drops              int
	ResponseSizeSpike  int
	OffHoursSpike      int
	BulkDataExtraction int
}

// DetectRateAnomalies compares current window counts to 24h rolling average.
func (a *Activities) DetectRateAnomalies(ctx context.Context) (*DetectRateAnomaliesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting rate anomalies")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)
	baselineStart := windowEnd.Add(-24 * time.Hour)

	// Get current window counts per (endpoint_id, source_id).
	currentRows, err := a.db.QueryContext(ctx, `
		SELECT COALESCE(endpoint_id, ''), source_id, uri, COALESCE(method, ''),
			SUM(request_count) as total_count
		FROM silver.httptraffic_traffic_5m
		WHERE window_start >= $1 AND window_start < $2
		GROUP BY endpoint_id, source_id, uri, method`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query current window: %w", err)
	}
	defer currentRows.Close()

	type currentEntry struct {
		endpointID string
		sourceID   string
		uri        string
		method     string
		count      int64
	}
	var entries []currentEntry
	for currentRows.Next() {
		var e currentEntry
		if err := currentRows.Scan(&e.endpointID, &e.sourceID, &e.uri, &e.method, &e.count); err != nil {
			return nil, fmt.Errorf("scan current: %w", err)
		}
		entries = append(entries, e)
	}
	if err := currentRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate current window rows: %w", err)
	}

	// Get 24h baseline per (endpoint_id, source_id).
	baselineRows, err := a.db.QueryContext(ctx, `
		SELECT COALESCE(endpoint_id, ''), source_id, uri, COALESCE(method, ''),
			COALESCE(AVG(request_count), 0) as avg_count,
			COALESCE(STDDEV_POP(request_count), 0) as stddev_count
		FROM silver.httptraffic_traffic_5m
		WHERE window_start >= $1 AND window_start < $2
		GROUP BY endpoint_id, source_id, uri, method`, baselineStart, windowStart)
	if err != nil {
		return nil, fmt.Errorf("query baseline: %w", err)
	}
	defer baselineRows.Close()

	type baselineKey struct {
		endpointID, sourceID, uri, method string
	}
	type baselineVal struct {
		avg    float64
		stddev float64
	}
	baselines := make(map[baselineKey]baselineVal)
	for baselineRows.Next() {
		var k baselineKey
		var v baselineVal
		if err := baselineRows.Scan(&k.endpointID, &k.sourceID, &k.uri, &k.method, &v.avg, &v.stddev); err != nil {
			return nil, fmt.Errorf("scan baseline: %w", err)
		}
		baselines[k] = v
	}
	if err := baselineRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate baseline rows: %w", err)
	}

	zHigh := r.Threshold("traffic_spike_high", "z_score", 3.0)
	zWarning := r.Threshold("traffic_spike_warning", "z_score", 2.0)
	zDrop := r.Threshold("traffic_drop", "z_score", -2.0)
	dropMinPct := r.Threshold("traffic_drop", "min_pct", 0.1)

	detectedAt := time.Now()
	var spikes, drops int

	for _, e := range entries {
		key := baselineKey{e.endpointID, e.sourceID, e.uri, e.method}
		bl, ok := baselines[key]
		if !ok || bl.avg == 0 || bl.stddev == 0 {
			continue
		}

		z := (float64(e.count) - bl.avg) / bl.stddev

		var anomalyType, severity string
		if z > zHigh {
			anomalyType = "traffic_spike"
			severity = "high"
			spikes++
		} else if z > zWarning {
			anomalyType = "traffic_spike"
			severity = "medium"
			spikes++
		} else if z < zDrop && float64(e.count) < bl.avg*dropMinPct {
			anomalyType = "traffic_drop"
			severity = "high"
			drops++
		}

		if anomalyType == "" {
			continue
		}

		resourceID := fmt.Sprintf("rate:%s:%s:%s:%s:%s",
			e.sourceID, windowStart.Format(time.RFC3339), e.uri, e.method, anomalyType)

		a.createAnomaly(ctx, resourceID, e.endpointID, e.sourceID, anomalyType, severity,
			windowStart, windowEnd, e.uri, e.method, bl.avg, float64(e.count), z,
			fmt.Sprintf("%s: z-score=%.1f, baseline=%.0f, actual=%d", anomalyType, z, bl.avg, e.count),
			detectedAt, nil)
	}

	// --- response_size_anomaly: Z-score on total_body_bytes_sent ---

	zResponseSize := r.Threshold("response_size_anomaly", "z_score", 3.0)
	var responseSizeSpikes int

	respRows, err := a.db.QueryContext(ctx, `
		SELECT COALESCE(endpoint_id, ''), source_id, uri, COALESCE(method, ''),
			SUM(total_body_bytes_sent) as total_bytes
		FROM silver.httptraffic_traffic_5m
		WHERE window_start >= $1 AND window_start < $2
			AND total_body_bytes_sent > 0
		GROUP BY endpoint_id, source_id, uri, method`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query current response sizes: %w", err)
	}
	defer respRows.Close()

	type respEntry struct {
		endpointID, sourceID, uri, method string
		totalBytes                        int64
	}
	var respEntries []respEntry
	for respRows.Next() {
		var e respEntry
		if err := respRows.Scan(&e.endpointID, &e.sourceID, &e.uri, &e.method, &e.totalBytes); err != nil {
			return nil, fmt.Errorf("scan response size: %w", err)
		}
		respEntries = append(respEntries, e)
	}
	if err := respRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate response size rows: %w", err)
	}

	respBaselineRows, err := a.db.QueryContext(ctx, `
		SELECT COALESCE(endpoint_id, ''), source_id, uri, COALESCE(method, ''),
			COALESCE(AVG(total_body_bytes_sent), 0),
			COALESCE(STDDEV_POP(total_body_bytes_sent), 0)
		FROM silver.httptraffic_traffic_5m
		WHERE window_start >= $1 AND window_start < $2
			AND total_body_bytes_sent > 0
		GROUP BY endpoint_id, source_id, uri, method`, baselineStart, windowStart)
	if err != nil {
		return nil, fmt.Errorf("query response size baseline: %w", err)
	}
	defer respBaselineRows.Close()

	respBaselines := make(map[baselineKey]baselineVal)
	for respBaselineRows.Next() {
		var k baselineKey
		var v baselineVal
		if err := respBaselineRows.Scan(&k.endpointID, &k.sourceID, &k.uri, &k.method, &v.avg, &v.stddev); err != nil {
			return nil, fmt.Errorf("scan response size baseline: %w", err)
		}
		respBaselines[k] = v
	}
	if err := respBaselineRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate response size baseline rows: %w", err)
	}

	for _, e := range respEntries {
		key := baselineKey{e.endpointID, e.sourceID, e.uri, e.method}
		bl, ok := respBaselines[key]
		if !ok || bl.avg == 0 || bl.stddev == 0 {
			continue
		}
		z := (float64(e.totalBytes) - bl.avg) / bl.stddev
		if z > zResponseSize {
			resourceID := fmt.Sprintf("respsize:%s:%s:%s:%s",
				e.sourceID, windowStart.Format(time.RFC3339), e.uri, e.method)
			a.createAnomaly(ctx, resourceID, e.endpointID, e.sourceID, "response_size_anomaly", "high",
				windowStart, windowEnd, e.uri, e.method, bl.avg, float64(e.totalBytes), z,
				fmt.Sprintf("Response body bytes z-score=%.1f, baseline=%.0f, actual=%d", z, bl.avg, e.totalBytes),
				detectedAt, nil)
			responseSizeSpikes++
		}
	}

	// --- off_hours_spike: traffic during 00:00-05:00 exceeding multiplier of baseline ---

	offHoursMultiplier := r.Threshold("off_hours_spike", "multiplier", 3.0)
	var offHoursSpikes int

	hour := windowStart.UTC().Hour()
	if hour >= 0 && hour < 5 {
		for _, e := range entries {
			key := baselineKey{e.endpointID, e.sourceID, e.uri, e.method}
			bl, ok := baselines[key]
			if !ok || bl.avg == 0 {
				continue
			}
			if float64(e.count) > bl.avg*offHoursMultiplier {
				resourceID := fmt.Sprintf("offhours:%s:%s:%s:%s",
					e.sourceID, windowStart.Format(time.RFC3339), e.uri, e.method)
				ratio := float64(e.count) / bl.avg
				a.createAnomaly(ctx, resourceID, e.endpointID, e.sourceID, "off_hours_spike", "medium",
					windowStart, windowEnd, e.uri, e.method, bl.avg, float64(e.count), ratio,
					fmt.Sprintf("Off-hours traffic %.1fx baseline (hour=%d, actual=%d, baseline=%.0f)", ratio, hour, e.count, bl.avg),
					detectedAt, nil)
				offHoursSpikes++
			}
		}
	}

	// bulk_data_extraction: single client downloading excessive data.
	// When unique_client_count=1, that one IP received all the bytes.
	bulkBytesMin := r.Threshold("bulk_data_extraction", "bytes_min", 10485760) // 10 MB
	var bulkCount int

	bulkRows, err := a.db.QueryContext(ctx, `
		SELECT source_id, COALESCE(endpoint_id, ''), uri, COALESCE(method, ''),
			total_body_bytes_sent
		FROM silver.httptraffic_traffic_5m
		WHERE window_start >= $1 AND window_start < $2
			AND unique_client_count = 1
			AND total_body_bytes_sent > $3`,
		windowStart, windowEnd, int64(bulkBytesMin))
	if err != nil {
		return nil, fmt.Errorf("query bulk data extraction: %w", err)
	}
	defer bulkRows.Close()

	for bulkRows.Next() {
		var sourceID, endpointID, uri, method string
		var bodyBytes int64
		if err := bulkRows.Scan(&sourceID, &endpointID, &uri, &method, &bodyBytes); err != nil {
			return nil, fmt.Errorf("scan bulk data extraction: %w", err)
		}
		resourceID := fmt.Sprintf("bulkdata:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), uri, method)
		mb := float64(bodyBytes) / (1024 * 1024)
		evidence, _ := json.Marshal(map[string]any{
			"uri": uri, "method": method, "bytes": bodyBytes, "mb": mb,
		})
		a.createAnomaly(ctx, resourceID, endpointID, sourceID, "bulk_data_extraction", "high",
			windowStart, windowEnd, uri, method, bulkBytesMin, float64(bodyBytes), mb,
			fmt.Sprintf("Single client downloaded %.1f MB from %s %s in 5 min", mb, method, uri),
			detectedAt, evidence)
		bulkCount++
	}
	if err := bulkRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bulk data extraction rows: %w", err)
	}

	logger.Info("Rate anomaly detection complete",
		"spikes", spikes, "drops", drops,
		"responseSizeSpikes", responseSizeSpikes, "offHoursSpikes", offHoursSpikes,
		"bulkDataExtraction", bulkCount)
	return &DetectRateAnomaliesResult{
		Spikes:             spikes,
		Drops:              drops,
		ResponseSizeSpike:  responseSizeSpikes,
		OffHoursSpike:      offHoursSpikes,
		BulkDataExtraction: bulkCount,
	}, nil
}

// --- Activity 2: DetectErrorBursts ---

// DetectErrorBurstsResult holds output.
type DetectErrorBurstsResult struct {
	ErrorBursts int
	FiveXXBursts int
}

// DetectErrorBursts detects elevated error rates.
func (a *Activities) DetectErrorBursts(ctx context.Context) (*DetectErrorBurstsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting error bursts")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)

	errorCountMin := r.ThresholdInt("error_burst", "count_min", 10)

	rows, err := a.db.QueryContext(ctx, `
		SELECT COALESCE(endpoint_id, ''), source_id, uri, COALESCE(method, ''),
			SUM(request_count) as total,
			SUM(CASE WHEN status_code >= 400 THEN request_count ELSE 0 END) as errors,
			SUM(CASE WHEN status_code >= 500 THEN request_count ELSE 0 END) as server_errors
		FROM silver.httptraffic_traffic_5m
		WHERE window_start >= $1 AND window_start < $2
		GROUP BY endpoint_id, source_id, uri, method
		HAVING SUM(request_count) > $3`, windowStart, windowEnd, errorCountMin)
	if err != nil {
		return nil, fmt.Errorf("query error rates: %w", err)
	}
	defer rows.Close()

	fiveXXRateMin := r.Threshold("5xx_burst", "rate_min", 0.05)
	fiveXXCountMin := r.ThresholdInt("5xx_burst", "count_min", 5)
	errorRateMin := r.Threshold("error_burst", "rate_min", 0.05)

	detectedAt := time.Now()
	var errorBursts, fiveXXBursts int

	for rows.Next() {
		var endpointID, sourceID, uri, method string
		var total, errors, serverErrors int64
		if err := rows.Scan(&endpointID, &sourceID, &uri, &method, &total, &errors, &serverErrors); err != nil {
			return nil, fmt.Errorf("scan error rates: %w", err)
		}

		errorRate := float64(errors) / float64(total)
		fiveXXRate := float64(serverErrors) / float64(total)

		if fiveXXRate > fiveXXRateMin && serverErrors > fiveXXCountMin {
			resourceID := fmt.Sprintf("5xx:%s:%s:%s:%s",
				sourceID, windowStart.Format(time.RFC3339), uri, method)
			a.createAnomaly(ctx, resourceID, endpointID, sourceID, "5xx_burst", "high",
				windowStart, windowEnd, uri, method, 0, fiveXXRate, 0,
				fmt.Sprintf("5xx rate=%.1f%%, count=%d/%d", fiveXXRate*100, serverErrors, total),
				detectedAt, nil)
			fiveXXBursts++
		} else if errorRate > errorRateMin && errors > errorCountMin {
			resourceID := fmt.Sprintf("err:%s:%s:%s:%s",
				sourceID, windowStart.Format(time.RFC3339), uri, method)
			a.createAnomaly(ctx, resourceID, endpointID, sourceID, "error_burst", "medium",
				windowStart, windowEnd, uri, method, 0, errorRate, 0,
				fmt.Sprintf("error rate=%.1f%%, count=%d/%d", errorRate*100, errors, total),
				detectedAt, nil)
			errorBursts++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate error rate rows: %w", err)
	}

	logger.Info("Error burst detection complete", "errorBursts", errorBursts, "5xxBursts", fiveXXBursts)
	return &DetectErrorBurstsResult{ErrorBursts: errorBursts, FiveXXBursts: fiveXXBursts}, nil
}

// --- Activity 3: DetectSuspiciousPatterns ---

// DetectSuspiciousPatternsResult holds output.
type DetectSuspiciousPatternsResult struct {
	ScannerDetected      int
	SingleIPFlood        int
	EndpointEnumeration  int
	PathTraversal        int
	SQLInjection         int
	CommandInjection     int
	XSSProbe             int
	SSRFProbe            int
	PaginationScraping   int
}

// DetectSuspiciousPatterns detects scanner UAs and single-IP floods.
func (a *Activities) DetectSuspiciousPatterns(ctx context.Context) (*DetectSuspiciousPatternsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting suspicious patterns")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)
	detectedAt := time.Now()
	var scannerCount, ipFloodCount int

	// Scanner detection via is_scanner_detected boolean flag.
	scannerRows, err := a.db.QueryContext(ctx, `
		SELECT source_id, uri, COALESCE(method, ''), SUM(request_count)
		FROM silver.httptraffic_traffic_5m
		WHERE is_scanner_detected = true
			AND window_start >= $1 AND window_start < $2
		GROUP BY source_id, uri, method`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query scanner traffic: %w", err)
	}
	defer scannerRows.Close()

	for scannerRows.Next() {
		var sourceID, uri, method string
		var reqCount int64
		if err := scannerRows.Scan(&sourceID, &uri, &method, &reqCount); err != nil {
			return nil, fmt.Errorf("scan scanner: %w", err)
		}
		resourceID := fmt.Sprintf("scanner:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), uri, method)
		a.createAnomaly(ctx, resourceID, "", sourceID, "scanner_detected", "medium",
			windowStart, windowEnd, uri, method, 0, float64(reqCount), 0,
			fmt.Sprintf("scanner UA detected, %d requests", reqCount),
			detectedAt, nil)
		scannerCount++
	}
	if err := scannerRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scanner rows: %w", err)
	}

	// Single-IP flood detection.
	ipFloodMin := r.ThresholdInt("single_ip_flood", "count_min", 50)

	floodRows, err := a.db.QueryContext(ctx, `
		SELECT source_id, uri, COALESCE(method, ''), request_count
		FROM silver.httptraffic_traffic_5m
		WHERE unique_client_count = 1 AND request_count > $3
			AND window_start >= $1 AND window_start < $2`, windowStart, windowEnd, ipFloodMin)
	if err != nil {
		return nil, fmt.Errorf("query IP floods: %w", err)
	}
	defer floodRows.Close()

	for floodRows.Next() {
		var sourceID, uri, method string
		var reqCount int64
		if err := floodRows.Scan(&sourceID, &uri, &method, &reqCount); err != nil {
			return nil, fmt.Errorf("scan flood: %w", err)
		}
		resourceID := fmt.Sprintf("ipflood:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), uri, method)
		a.createAnomaly(ctx, resourceID, "", sourceID, "single_ip_flood", "medium",
			windowStart, windowEnd, uri, method, 0, float64(reqCount), 0,
			fmt.Sprintf("single IP sent %d requests", reqCount),
			detectedAt, nil)
		ipFloodCount++
	}
	if err := floodRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate flood rows: %w", err)
	}

	// Endpoint enumeration: many distinct 404 URIs per source IP.
	enumMin := r.ThresholdInt("endpoint_enumeration", "count_min", 30)
	var enumCount int

	enumRows, err := a.db.QueryContext(ctx, `
		SELECT c.source_id, c.client_ip, COUNT(DISTINCT t.uri) as uri_count
		FROM silver.httptraffic_client_ip_5m c
		JOIN silver.httptraffic_traffic_5m t
			ON c.source_id = t.source_id
			AND c.window_start = t.window_start
			AND c.uri = t.uri
		WHERE c.window_start >= $1 AND c.window_start < $2
			AND t.status_code = 404
		GROUP BY c.source_id, c.client_ip
		HAVING COUNT(DISTINCT t.uri) > $3`, windowStart, windowEnd, enumMin)
	if err != nil {
		return nil, fmt.Errorf("query endpoint enumeration: %w", err)
	}
	defer enumRows.Close()

	for enumRows.Next() {
		var sourceID, clientIP string
		var uriCount int64
		if err := enumRows.Scan(&sourceID, &clientIP, &uriCount); err != nil {
			return nil, fmt.Errorf("scan endpoint enumeration: %w", err)
		}
		resourceID := fmt.Sprintf("enum:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), clientIP)
		evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "uri_count": uriCount})
		a.createAnomaly(ctx, resourceID, "", sourceID, "endpoint_enumeration", "high",
			windowStart, windowEnd, "", "", 0, float64(uriCount), 0,
			fmt.Sprintf("IP %s probed %d distinct URIs returning 404", clientIP, uriCount),
			detectedAt, evidence)
		enumCount++
	}
	if err := enumRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate endpoint enumeration rows: %w", err)
	}

	// URI attack pattern detection via boolean flags set during normalization.
	type attackDetection struct {
		flag        string
		anomalyType string
		severity    string
		prefix      string
		label       string
	}
	attacks := []attackDetection{
		{"is_lfi_detected", "path_traversal", "critical", "lfi", "path traversal / LFI"},
		{"is_sqli_detected", "sql_injection_probe", "critical", "sqli", "SQL injection"},
		{"is_rce_detected", "command_injection_probe", "critical", "rce", "command injection / RCE"},
		{"is_xss_detected", "xss_probe", "high", "xss", "XSS"},
		{"is_ssrf_detected", "ssrf_probe", "critical", "ssrf", "SSRF"},
	}
	var lfiCount, sqliCount, rceCount, xssCount, ssrfCount int
	attackCounts := []*int{&lfiCount, &sqliCount, &rceCount, &xssCount, &ssrfCount}

	for i, atk := range attacks {
		atkRows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT source_id, uri, COALESCE(method, ''), SUM(request_count)
			FROM silver.httptraffic_traffic_5m
			WHERE %s = true
				AND window_start >= $1 AND window_start < $2
			GROUP BY source_id, uri, method`, atk.flag), windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query %s traffic: %w", atk.prefix, err)
		}

		for atkRows.Next() {
			var sourceID, uri, method string
			var reqCount int64
			if err := atkRows.Scan(&sourceID, &uri, &method, &reqCount); err != nil {
				atkRows.Close()
				return nil, fmt.Errorf("scan %s: %w", atk.prefix, err)
			}
			resourceID := fmt.Sprintf("%s:%s:%s:%s:%s",
				atk.prefix, sourceID, windowStart.Format(time.RFC3339), uri, method)
			a.createAnomaly(ctx, resourceID, "", sourceID, atk.anomalyType, atk.severity,
				windowStart, windowEnd, uri, method, 0, float64(reqCount), 0,
				fmt.Sprintf("%s detected, %d requests", atk.label, reqCount),
				detectedAt, nil)
			*attackCounts[i]++
		}
		if err := atkRows.Err(); err != nil {
			atkRows.Close()
			return nil, fmt.Errorf("iterate %s rows: %w", atk.prefix, err)
		}
		atkRows.Close()
	}

	// pagination_scraping: single IP making many requests to same base URI with different query params.
	// Strip query string, group by (IP, base_path), count distinct full URIs.
	pagMin := r.ThresholdInt("pagination_scraping", "count_min", 50)
	var pagCount int

	pagRows, err := a.db.QueryContext(ctx, `
		SELECT c.source_id, c.client_ip,
			SPLIT_PART(c.uri, '?', 1) AS base_path,
			COUNT(DISTINCT c.uri) AS uri_count,
			SUM(c.request_count) AS req_count
		FROM silver.httptraffic_client_ip_5m c
		WHERE c.window_start >= $1 AND c.window_start < $2
			AND c.uri LIKE '%?%'
		GROUP BY c.source_id, c.client_ip, SPLIT_PART(c.uri, '?', 1)
		HAVING COUNT(DISTINCT c.uri) > $3`, windowStart, windowEnd, pagMin)
	if err != nil {
		return nil, fmt.Errorf("query pagination scraping: %w", err)
	}
	defer pagRows.Close()

	for pagRows.Next() {
		var sourceID, clientIP, basePath string
		var uriCount, reqCount int64
		if err := pagRows.Scan(&sourceID, &clientIP, &basePath, &uriCount, &reqCount); err != nil {
			return nil, fmt.Errorf("scan pagination scraping: %w", err)
		}
		resourceID := fmt.Sprintf("pagscrape:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), clientIP, basePath)
		evidence, _ := json.Marshal(map[string]any{
			"ip": clientIP, "base_path": basePath,
			"distinct_uris": uriCount, "total_requests": reqCount,
		})
		a.createAnomaly(ctx, resourceID, "", sourceID, "pagination_scraping", "medium",
			windowStart, windowEnd, basePath, "", 0, float64(uriCount), 0,
			fmt.Sprintf("IP %s scraped %d distinct query variants of %s (%d requests)",
				clientIP, uriCount, basePath, reqCount),
			detectedAt, evidence)
		pagCount++
	}
	if err := pagRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pagination scraping rows: %w", err)
	}

	logger.Info("Suspicious pattern detection complete",
		"scanners", scannerCount, "ipFloods", ipFloodCount, "endpointEnum", enumCount,
		"lfi", lfiCount, "sqli", sqliCount, "rce", rceCount, "xss", xssCount, "ssrf", ssrfCount,
		"pagScraping", pagCount)
	return &DetectSuspiciousPatternsResult{
		ScannerDetected:    scannerCount,
		SingleIPFlood:      ipFloodCount,
		EndpointEnumeration: enumCount,
		PathTraversal:      lfiCount,
		SQLInjection:       sqliCount,
		CommandInjection:   rceCount,
		XSSProbe:           xssCount,
		SSRFProbe:          ssrfCount,
		PaginationScraping: pagCount,
	}, nil
}

// --- Activity 4: DetectMethodMismatch ---

// DetectMethodMismatchResult holds output.
type DetectMethodMismatchResult struct {
	Detected int
}

// DetectMethodMismatch detects traffic using HTTP methods not in the endpoint's allowed list.
func (a *Activities) DetectMethodMismatch(ctx context.Context) (*DetectMethodMismatchResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting method mismatches")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)

	countMin := r.ThresholdInt("method_mismatch_warning", "count_min", 10)
	countHigh := r.ThresholdInt("method_mismatch_high", "count_min", 50)

	rows, err := a.db.QueryContext(ctx, `
		SELECT COALESCE(endpoint_id, ''), source_id, uri, COALESCE(method, ''),
			SUM(request_count) as total_count
		FROM silver.httptraffic_traffic_5m
		WHERE is_method_mismatch = true
			AND window_start >= $1 AND window_start < $2
		GROUP BY endpoint_id, source_id, uri, method
		HAVING SUM(request_count) > $3`, windowStart, windowEnd, countMin)
	if err != nil {
		return nil, fmt.Errorf("query method mismatches: %w", err)
	}
	defer rows.Close()

	detectedAt := time.Now()
	var count int

	for rows.Next() {
		var endpointID, sourceID, uri, method string
		var totalCount int64
		if err := rows.Scan(&endpointID, &sourceID, &uri, &method, &totalCount); err != nil {
			return nil, fmt.Errorf("scan method mismatch: %w", err)
		}

		severity := "medium"
		if totalCount > countHigh {
			severity = "high"
		}

		resourceID := fmt.Sprintf("method:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), uri, method)
		a.createAnomaly(ctx, resourceID, endpointID, sourceID, "method_mismatch", severity,
			windowStart, windowEnd, uri, method, 0, float64(totalCount), 0,
			fmt.Sprintf("%s on endpoint that does not allow it, %d requests", method, totalCount),
			detectedAt, nil)
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate method mismatch rows: %w", err)
	}

	logger.Info("Method mismatch detection complete", "detected", count)
	return &DetectMethodMismatchResult{Detected: count}, nil
}

// --- Activity 5: DetectNewEndpoints ---

// DetectNewEndpointsResult holds output.
type DetectNewEndpointsResult struct {
	NewEndpoints int
}

// DetectNewEndpoints finds unmapped URIs with significant traffic.
func (a *Activities) DetectNewEndpoints(ctx context.Context) (*DetectNewEndpointsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting new endpoints")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	lookback := windowEnd.Add(-1 * time.Hour)

	newEndpointMin := r.ThresholdInt("new_endpoint", "count_min", 100)

	rows, err := a.db.QueryContext(ctx, `
		SELECT source_id, uri, COALESCE(method, ''),
			SUM(request_count) as total_count
		FROM silver.httptraffic_traffic_5m
		WHERE is_mapped = false
			AND window_start >= $1 AND window_start < $2
		GROUP BY source_id, uri, method
		HAVING SUM(request_count) > $3
		ORDER BY total_count DESC
		LIMIT 50`, lookback, windowEnd, newEndpointMin)
	if err != nil {
		return nil, fmt.Errorf("query unmapped endpoints: %w", err)
	}
	defer rows.Close()

	detectedAt := time.Now()
	var count int

	for rows.Next() {
		var sourceID, uri, method string
		var totalCount int64
		if err := rows.Scan(&sourceID, &uri, &method, &totalCount); err != nil {
			return nil, fmt.Errorf("scan unmapped: %w", err)
		}

		resourceID := fmt.Sprintf("newep:%s:%s:%s:%s",
			sourceID, windowEnd.Format(time.RFC3339), uri, method)
		a.createAnomaly(ctx, resourceID, "", sourceID, "new_endpoint", "info",
			lookback, windowEnd, uri, method, 0, float64(totalCount), 0,
			fmt.Sprintf("unmapped URI with %d requests in 1h", totalCount),
			detectedAt, nil)
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unmapped endpoint rows: %w", err)
	}

	logger.Info("New endpoint detection complete", "found", count)
	return &DetectNewEndpointsResult{NewEndpoints: count}, nil
}

// --- Activity 5: DetectUserAgentAnomalies ---

// DetectUserAgentAnomaliesResult holds output.
type DetectUserAgentAnomaliesResult struct {
	NewUA           int
	ShareShift      int
	AutomatedClient int
	UASpoofing      int
}

// DetectUserAgentAnomalies detects unusual UA family patterns against 7-day baseline.
func (a *Activities) DetectUserAgentAnomalies(ctx context.Context) (*DetectUserAgentAnomaliesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting user agent anomalies")

	matchRules, err := a.matchRules.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load match rules: %w", err)
	}

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)
	baselineStart := windowEnd.Add(-7 * 24 * time.Hour)
	detectedAt := time.Now()

	result := &DetectUserAgentAnomaliesResult{}

	// Get endpoints with UA traffic in current window.
	epRows, err := a.db.QueryContext(ctx, `
		SELECT DISTINCT COALESCE(endpoint_id, ''), uri
		FROM silver.httptraffic_user_agent_5m
		WHERE window_start >= $1 AND window_start < $2`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query UA endpoints: %w", err)
	}
	defer epRows.Close()

	type epKey struct{ endpointID, uri string }
	var endpoints []epKey
	for epRows.Next() {
		var k epKey
		if err := epRows.Scan(&k.endpointID, &k.uri); err != nil {
			return nil, fmt.Errorf("scan UA endpoint: %w", err)
		}
		endpoints = append(endpoints, k)
	}
	if err := epRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate UA endpoint rows: %w", err)
	}

	// Load endpoint access levels for automated_client detection.
	accessLevels := make(map[string]string)
	alRows, err := a.db.QueryContext(ctx, `
		SELECT resource_id, COALESCE(access_level, '')
		FROM silver.inventory_api_endpoints WHERE is_active = true`)
	if err == nil {
		defer alRows.Close()
		for alRows.Next() {
			var id, al string
			if err := alRows.Scan(&id, &al); err != nil {
				return nil, fmt.Errorf("scan access level: %w", err)
			}
			accessLevels[id] = al
		}
		if err := alRows.Err(); err != nil {
			return nil, fmt.Errorf("iterate access level rows: %w", err)
		}
	}

	for _, ep := range endpoints {
		// 7-day baseline per ua_family.
		baselineRows, err := a.db.QueryContext(ctx, `
			SELECT COALESCE(ua_family, ''), SUM(request_count)
			FROM silver.httptraffic_user_agent_5m
			WHERE COALESCE(endpoint_id, '') = $1 AND uri = $2
				AND window_start >= $3 AND window_start < $4
			GROUP BY ua_family`, ep.endpointID, ep.uri, baselineStart, windowStart)
		if err != nil {
			return nil, fmt.Errorf("query UA baseline: %w", err)
		}
		baseline := make(map[string]int64)
		var baselineTotal int64
		for baselineRows.Next() {
			var family string
			var total int64
			if err := baselineRows.Scan(&family, &total); err != nil {
				baselineRows.Close()
				return nil, fmt.Errorf("scan UA baseline: %w", err)
			}
			baseline[family] = total
			baselineTotal += total
		}
		if err := baselineRows.Err(); err != nil {
			baselineRows.Close()
			return nil, fmt.Errorf("iterate UA baseline rows: %w", err)
		}
		baselineRows.Close()

		// Current window per ua_family.
		currentRows, err := a.db.QueryContext(ctx, `
			SELECT COALESCE(ua_family, ''), SUM(request_count)
			FROM silver.httptraffic_user_agent_5m
			WHERE COALESCE(endpoint_id, '') = $1 AND uri = $2
				AND window_start >= $3 AND window_start < $4
			GROUP BY ua_family`, ep.endpointID, ep.uri, windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query current UA: %w", err)
		}
		current := make(map[string]int64)
		var currentTotal int64
		for currentRows.Next() {
			var family string
			var count int64
			if err := currentRows.Scan(&family, &count); err != nil {
				currentRows.Close()
				return nil, fmt.Errorf("scan current UA: %w", err)
			}
			current[family] = count
			currentTotal += count
		}
		if err := currentRows.Err(); err != nil {
			currentRows.Close()
			return nil, fmt.Errorf("iterate current UA rows: %w", err)
		}
		currentRows.Close()

		if currentTotal == 0 {
			continue
		}

		newUAShareHigh := r.Threshold("new_user_agent_high", "share_min", 0.20)
		newUAShareWarning := r.Threshold("new_user_agent_warning", "share_min", 0.05)
		uaShareShiftDelta := r.Threshold("ua_share_shift", "share_delta", 0.30)

		for family, count := range current {
			currentShare := float64(count) / float64(currentTotal)

			// new_user_agent: family not in baseline.
			if _, inBaseline := baseline[family]; !inBaseline {
				if currentShare > newUAShareHigh {
					resourceID := fmt.Sprintf("newua:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), family)
					evidence, _ := json.Marshal(map[string]any{"ua_family": family, "share": currentShare, "count": count})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "new_user_agent", "high",
						windowStart, windowEnd, ep.uri, "", 0, currentShare, 0,
						fmt.Sprintf("New UA family '%s' appeared with %.0f%% share (%d/%d requests)", family, currentShare*100, count, currentTotal),
						detectedAt, evidence)
					result.NewUA++
				} else if currentShare > newUAShareWarning {
					resourceID := fmt.Sprintf("newua:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), family)
					evidence, _ := json.Marshal(map[string]any{"ua_family": family, "share": currentShare, "count": count})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "new_user_agent", "low",
						windowStart, windowEnd, ep.uri, "", 0, currentShare, 0,
						fmt.Sprintf("New UA family '%s' appeared with %.0f%% share", family, currentShare*100),
						detectedAt, evidence)
					result.NewUA++
				}
			}

			// ua_share_shift: share changed significantly vs baseline.
			if baselineTotal > 0 {
				baselineShare := float64(baseline[family]) / float64(baselineTotal)
				delta := currentShare - baselineShare
				if delta > uaShareShiftDelta || delta < -uaShareShiftDelta {
					resourceID := fmt.Sprintf("uashift:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), family)
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "ua_share_shift", "medium",
						windowStart, windowEnd, ep.uri, "", baselineShare, currentShare, delta,
						fmt.Sprintf("UA family '%s' share changed from %.0f%% to %.0f%%", family, baselineShare*100, currentShare*100),
						detectedAt, nil)
					result.ShareShift++
				}
			}

			// automated_client: library UA on protected endpoint.
			if matchRules.IsLibraryUA(family) && ep.endpointID != "" {
				if accessLevels[ep.endpointID] == "protected" {
					resourceID := fmt.Sprintf("autoua:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), family)
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "automated_client", "medium",
						windowStart, windowEnd, ep.uri, "", 0, float64(count), 0,
						fmt.Sprintf("Library UA '%s' on protected endpoint %s", family, ep.uri),
						detectedAt, nil)
					result.AutomatedClient++
				}
			}
		}
	}

	// ua_spoofing: browser UA + high request rate + no static asset requests.
	// A real browser loads CSS/JS/images; a bot claiming to be a browser typically only hits API paths.
	browserFamilies := map[string]bool{"chrome": true, "firefox": true, "safari": true, "edge": true}

	// Find IPs claiming browser UA with high volume, then check for static asset requests.
	spoofRows, err := a.db.QueryContext(ctx, `
		SELECT u.source_id, c.client_ip, u.ua_family, SUM(c.request_count) as req_count
		FROM silver.httptraffic_user_agent_5m u
		JOIN silver.httptraffic_client_ip_5m c
			ON u.source_id = c.source_id AND u.window_start = c.window_start
			AND u.uri = c.uri AND COALESCE(u.method, '') = COALESCE(c.method, '')
		WHERE u.window_start >= $1 AND u.window_start < $2
			AND u.ua_family IN ('chrome', 'firefox', 'safari', 'edge')
		GROUP BY u.source_id, c.client_ip, u.ua_family
		HAVING SUM(c.request_count) > 50`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query ua spoofing candidates: %w", err)
	}
	defer spoofRows.Close()

	type spoofCandidate struct {
		sourceID, clientIP, uaFamily string
		reqCount                     int64
	}
	var candidates []spoofCandidate
	for spoofRows.Next() {
		var sc spoofCandidate
		if err := spoofRows.Scan(&sc.sourceID, &sc.clientIP, &sc.uaFamily, &sc.reqCount); err != nil {
			return nil, fmt.Errorf("scan ua spoofing: %w", err)
		}
		if browserFamilies[sc.uaFamily] {
			candidates = append(candidates, sc)
		}
	}
	if err := spoofRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate UA spoofing candidate rows: %w", err)
	}

	// Check each candidate for static asset requests (real browsers load .js, .css, .png, etc.)
	for _, sc := range candidates {
		var staticCount int64
		err := a.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM silver.httptraffic_client_ip_5m
			WHERE source_id = $1 AND client_ip = $2
				AND window_start >= $3 AND window_start < $4
				AND (uri LIKE '%.js' OR uri LIKE '%.css' OR uri LIKE '%.png'
					OR uri LIKE '%.jpg' OR uri LIKE '%.gif' OR uri LIKE '%.svg'
					OR uri LIKE '%.woff%' OR uri LIKE '%.ico')`,
			sc.sourceID, sc.clientIP, windowStart, windowEnd).Scan(&staticCount)
		if err != nil {
			return nil, fmt.Errorf("query static assets for %s: %w", sc.clientIP, err)
		}
		if staticCount == 0 {
			resourceID := fmt.Sprintf("uaspoof:%s:%s:%s:%s",
				sc.sourceID, windowStart.Format(time.RFC3339), sc.clientIP, sc.uaFamily)
			evidence, _ := json.Marshal(map[string]any{
				"ip": sc.clientIP, "ua_family": sc.uaFamily,
				"request_count": sc.reqCount, "static_assets": 0,
			})
			a.createAnomaly(ctx, resourceID, "", sc.sourceID, "ua_spoofing", "medium",
				windowStart, windowEnd, "", "", 0, float64(sc.reqCount), 0,
				fmt.Sprintf("IP %s claims %s UA but made %d requests with no static assets",
					sc.clientIP, sc.uaFamily, sc.reqCount),
				detectedAt, evidence)
			result.UASpoofing++
		}
	}

	logger.Info("User agent anomaly detection complete",
		"newUA", result.NewUA, "shareShift", result.ShareShift,
		"automated", result.AutomatedClient, "uaSpoofing", result.UASpoofing)
	return result, nil
}

// --- Activity 6: DetectClientIPAnomalies ---

// DetectClientIPAnomaliesResult holds output.
type DetectClientIPAnomaliesResult struct {
	NewSourceIP        int
	GeoShift           int
	ExternalOnInternal int
	IPConcentration    int
	SanctionedCountry  int
	IPRotation         int
}

// DetectClientIPAnomalies detects unusual IP patterns against 7-day baseline.
func (a *Activities) DetectClientIPAnomalies(ctx context.Context) (*DetectClientIPAnomaliesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting client IP anomalies")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)
	baselineStart := windowEnd.Add(-7 * 24 * time.Hour)
	detectedAt := time.Now()

	result := &DetectClientIPAnomaliesResult{}

	// sanctioned_country: traffic from OFAC/sanctioned countries.
	sanctionedRows, err := a.db.QueryContext(ctx, `
		SELECT c.source_id, c.client_ip, c.country_code, COALESCE(c.country_name, ''),
			SUM(c.request_count)
		FROM silver.httptraffic_client_ip_5m c
		JOIN config.sanctioned_countries s ON s.country_code = c.country_code AND s.is_active = true
		WHERE c.window_start >= $1 AND c.window_start < $2
			AND c.country_code IS NOT NULL
		GROUP BY c.source_id, c.client_ip, c.country_code, c.country_name`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query sanctioned countries: %w", err)
	}
	defer sanctionedRows.Close()

	for sanctionedRows.Next() {
		var sourceID, clientIP, countryCode, countryName string
		var reqCount int64
		if err := sanctionedRows.Scan(&sourceID, &clientIP, &countryCode, &countryName, &reqCount); err != nil {
			return nil, fmt.Errorf("scan sanctioned: %w", err)
		}
		resourceID := fmt.Sprintf("sanctioned:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), clientIP, countryCode)
		evidence, _ := json.Marshal(map[string]any{
			"ip": clientIP, "country_code": countryCode, "country_name": countryName, "count": reqCount,
		})
		desc := fmt.Sprintf("Traffic from sanctioned country %s (%s), IP %s, %d requests",
			countryCode, countryName, clientIP, reqCount)
		a.createAnomaly(ctx, resourceID, "", sourceID, "sanctioned_country", "critical",
			windowStart, windowEnd, "", "", 0, float64(reqCount), 0, desc,
			detectedAt, evidence)
		result.SanctionedCountry++
	}
	if err := sanctionedRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sanctioned country rows: %w", err)
	}

	// Get endpoints with IP traffic in current window.
	epRows, err := a.db.QueryContext(ctx, `
		SELECT DISTINCT COALESCE(endpoint_id, ''), uri
		FROM silver.httptraffic_client_ip_5m
		WHERE window_start >= $1 AND window_start < $2`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query IP endpoints: %w", err)
	}
	defer epRows.Close()

	type epKey struct{ endpointID, uri string }
	var endpoints []epKey
	for epRows.Next() {
		var k epKey
		if err := epRows.Scan(&k.endpointID, &k.uri); err != nil {
			return nil, fmt.Errorf("scan IP endpoint: %w", err)
		}
		endpoints = append(endpoints, k)
	}
	if err := epRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate IP endpoint rows: %w", err)
	}

	for _, ep := range endpoints {
		// 7-day baseline: distinct IPs, country distribution, % internal.
		baselineIPRows, err := a.db.QueryContext(ctx, `
			SELECT client_ip, COALESCE(country_code, ''), is_internal, SUM(request_count)
			FROM silver.httptraffic_client_ip_5m
			WHERE COALESCE(endpoint_id, '') = $1 AND uri = $2
				AND window_start >= $3 AND window_start < $4
			GROUP BY client_ip, country_code, is_internal`, ep.endpointID, ep.uri, baselineStart, windowStart)
		if err != nil {
			return nil, fmt.Errorf("query IP baseline: %w", err)
		}
		baselineIPs := make(map[string]bool)
		baselineCountry := make(map[string]int64)
		var baselineTotal, baselineInternalTotal int64
		for baselineIPRows.Next() {
			var ip, country string
			var isInternal bool
			var total int64
			if err := baselineIPRows.Scan(&ip, &country, &isInternal, &total); err != nil {
				baselineIPRows.Close()
				return nil, fmt.Errorf("scan IP baseline: %w", err)
			}
			baselineIPs[ip] = true
			if country != "" {
				baselineCountry[country] += total
			}
			baselineTotal += total
			if isInternal {
				baselineInternalTotal += total
			}
		}
		if err := baselineIPRows.Err(); err != nil {
			baselineIPRows.Close()
			return nil, fmt.Errorf("iterate IP baseline rows: %w", err)
		}
		baselineIPRows.Close()

		// Current window IPs.
		currentIPRows, err := a.db.QueryContext(ctx, `
			SELECT client_ip, COALESCE(country_code, ''), is_internal, SUM(request_count)
			FROM silver.httptraffic_client_ip_5m
			WHERE COALESCE(endpoint_id, '') = $1 AND uri = $2
				AND window_start >= $3 AND window_start < $4
			GROUP BY client_ip, country_code, is_internal`, ep.endpointID, ep.uri, windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query current IPs: %w", err)
		}
		type ipEntry struct {
			ip, country string
			isInternal  bool
			count       int64
		}
		var currentEntries []ipEntry
		currentCountry := make(map[string]int64)
		var currentTotal int64
		for currentIPRows.Next() {
			var e ipEntry
			if err := currentIPRows.Scan(&e.ip, &e.country, &e.isInternal, &e.count); err != nil {
				currentIPRows.Close()
				return nil, fmt.Errorf("scan current IP: %w", err)
			}
			currentEntries = append(currentEntries, e)
			if e.country != "" {
				currentCountry[e.country] += e.count
			}
			currentTotal += e.count
		}
		if err := currentIPRows.Err(); err != nil {
			currentIPRows.Close()
			return nil, fmt.Errorf("iterate current IP rows: %w", err)
		}
		currentIPRows.Close()

		if currentTotal == 0 {
			continue
		}

		newIPShareHigh := r.Threshold("new_source_ip_high", "share_min", 0.20)
		newIPShareWarning := r.Threshold("new_source_ip_warning", "share_min", 0.05)
		ipConcentrationMin := r.Threshold("ip_concentration", "share_min", 0.50)

		for _, e := range currentEntries {
			share := float64(e.count) / float64(currentTotal)

			// new_source_ip: IP not in 7d baseline.
			if !baselineIPs[e.ip] {
				if share > newIPShareHigh {
					resourceID := fmt.Sprintf("newip:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), e.ip)
					evidence, _ := json.Marshal(map[string]any{"ip": e.ip, "country": e.country, "share": share})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "new_source_ip", "high",
						windowStart, windowEnd, ep.uri, "", 0, share, 0,
						fmt.Sprintf("New IP %s appeared with %.0f%% of traffic", e.ip, share*100),
						detectedAt, evidence)
					result.NewSourceIP++
				} else if share > newIPShareWarning {
					resourceID := fmt.Sprintf("newip:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), e.ip)
					evidence, _ := json.Marshal(map[string]any{"ip": e.ip, "country": e.country, "share": share})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "new_source_ip", "low",
						windowStart, windowEnd, ep.uri, "", 0, share, 0,
						fmt.Sprintf("New IP %s appeared with %.0f%% of traffic", e.ip, share*100),
						detectedAt, evidence)
					result.NewSourceIP++
				}
			}

			// ip_concentration: single IP dominates traffic.
			if share > ipConcentrationMin {
				resourceID := fmt.Sprintf("ipconc:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), e.ip)
				a.createAnomaly(ctx, resourceID, ep.endpointID, "", "ip_concentration", "medium",
					windowStart, windowEnd, ep.uri, "", 0, share, 0,
					fmt.Sprintf("Single IP %s sent %.0f%% of traffic (%d/%d)", e.ip, share*100, e.count, currentTotal),
					detectedAt, nil)
				result.IPConcentration++
			}
		}

		// geo_shift: new country with significant share.
		geoNewCountryWarning := r.Threshold("geo_shift_new_country", "share_min", 0.10)
		geoNewCountryHigh := r.Threshold("geo_shift_new_country_high", "share_min", 0.20)
		geoShiftDelta := r.Threshold("geo_shift_existing", "shift_delta", 0.20)

		for country, count := range currentCountry {
			if country == "" {
				continue
			}
			share := float64(count) / float64(currentTotal)
			_, inBaseline := baselineCountry[country]
			if !inBaseline && share > geoNewCountryWarning {
				severity := "medium"
				if share > geoNewCountryHigh {
					severity = "high"
				}
				resourceID := fmt.Sprintf("geo:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), country)
				a.createAnomaly(ctx, resourceID, ep.endpointID, "", "geo_shift", severity,
					windowStart, windowEnd, ep.uri, "", 0, share, 0,
					fmt.Sprintf("New country %s appeared with %.0f%% share", country, share*100),
					detectedAt, nil)
				result.GeoShift++
			} else if inBaseline && baselineTotal > 0 {
				baselineShare := float64(baselineCountry[country]) / float64(baselineTotal)
				delta := share - baselineShare
				if delta > geoShiftDelta || delta < -geoShiftDelta {
					resourceID := fmt.Sprintf("geo:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), country)
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "geo_shift", "medium",
						windowStart, windowEnd, ep.uri, "", baselineShare, share, delta,
						fmt.Sprintf("Country %s share changed from %.0f%% to %.0f%%", country, baselineShare*100, share*100),
						detectedAt, nil)
					result.GeoShift++
				}
			}
		}

		// external_on_internal: external traffic on baseline-internal endpoint.
		extOnIntBaseline := r.Threshold("external_on_internal", "internal_baseline_min", 0.90)
		if baselineTotal > 0 {
			baselineInternalPct := float64(baselineInternalTotal) / float64(baselineTotal)
			if baselineInternalPct > extOnIntBaseline {
				// Check if current has external traffic.
				for _, e := range currentEntries {
					if !e.isInternal {
						resourceID := fmt.Sprintf("extint:%s:%s:%s", ep.endpointID, windowStart.Format(time.RFC3339), e.ip)
						a.createAnomaly(ctx, resourceID, ep.endpointID, "", "external_on_internal", "high",
							windowStart, windowEnd, ep.uri, "", baselineInternalPct, float64(e.count), 0,
							fmt.Sprintf("External IP %s on endpoint with %.0f%% internal baseline", e.ip, baselineInternalPct*100),
							detectedAt, nil)
						result.ExternalOnInternal++
						break
					}
				}
			}
		}
	}

	// ip_rotation: many distinct IPs from same /24 subnet hitting same URI.
	ipRotationMin := r.ThresholdInt("ip_rotation", "ip_count_min", 10)
	rotationRows, err := a.db.QueryContext(ctx, `
		SELECT source_id, uri, COALESCE(method, ''),
			SUBSTRING(client_ip FROM '^(\d+\.\d+\.\d+)\.')  AS subnet24,
			COUNT(DISTINCT client_ip) AS ip_count
		FROM silver.httptraffic_client_ip_5m
		WHERE window_start >= $1 AND window_start < $2
			AND client_ip ~ '^\d+\.\d+\.\d+\.\d+$'
		GROUP BY source_id, uri, method, subnet24
		HAVING COUNT(DISTINCT client_ip) > $3`, windowStart, windowEnd, ipRotationMin)
	if err != nil {
		return nil, fmt.Errorf("query ip rotation: %w", err)
	}
	defer rotationRows.Close()

	for rotationRows.Next() {
		var sourceID, uri, method, subnet string
		var ipCount int64
		if err := rotationRows.Scan(&sourceID, &uri, &method, &subnet, &ipCount); err != nil {
			return nil, fmt.Errorf("scan ip rotation: %w", err)
		}
		resourceID := fmt.Sprintf("iprotation:%s:%s:%s.0/24:%s:%s",
			sourceID, windowStart.Format(time.RFC3339), subnet, uri, method)
		evidence, _ := json.Marshal(map[string]any{
			"subnet": subnet + ".0/24", "distinct_ips": ipCount, "uri": uri, "method": method,
		})
		a.createAnomaly(ctx, resourceID, "", sourceID, "ip_rotation", "medium",
			windowStart, windowEnd, uri, method, 0, float64(ipCount), 0,
			fmt.Sprintf("%d distinct IPs from %s.0/24 hitting %s %s", ipCount, subnet, method, uri),
			detectedAt, evidence)
		result.IPRotation++
	}
	if err := rotationRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate IP rotation rows: %w", err)
	}

	logger.Info("Client IP anomaly detection complete",
		"newSourceIP", result.NewSourceIP, "geoShift", result.GeoShift,
		"externalOnInternal", result.ExternalOnInternal, "ipConcentration", result.IPConcentration,
		"ipRotation", result.IPRotation)
	return result, nil
}

// --- Activity 7: DetectASNAnomalies ---

// DetectASNAnomaliesResult holds output.
type DetectASNAnomaliesResult struct {
	NewASN           int
	HostingProvider  int
	ASNConcentration int
}

// DetectASNAnomalies detects unusual ASN patterns against 7-day baseline.
func (a *Activities) DetectASNAnomalies(ctx context.Context) (*DetectASNAnomaliesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting ASN anomalies")

	matchRules, err := a.matchRules.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load match rules: %w", err)
	}

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)
	baselineStart := windowEnd.Add(-7 * 24 * time.Hour)
	detectedAt := time.Now()

	result := &DetectASNAnomaliesResult{}

	// Get endpoints with IP traffic in current window.
	epRows, err := a.db.QueryContext(ctx, `
		SELECT DISTINCT COALESCE(endpoint_id, ''), uri
		FROM silver.httptraffic_client_ip_5m
		WHERE window_start >= $1 AND window_start < $2
			AND asn IS NOT NULL`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query ASN endpoints: %w", err)
	}
	defer epRows.Close()

	type epKey struct{ endpointID, uri string }
	var endpoints []epKey
	for epRows.Next() {
		var k epKey
		if err := epRows.Scan(&k.endpointID, &k.uri); err != nil {
			return nil, fmt.Errorf("scan ASN endpoint: %w", err)
		}
		endpoints = append(endpoints, k)
	}
	if err := epRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ASN endpoint rows: %w", err)
	}

	for _, ep := range endpoints {
		// 7-day baseline per ASN.
		baselineRows, err := a.db.QueryContext(ctx, `
			SELECT asn, COALESCE(org_name, ''), COALESCE(as_domain, ''), COALESCE(asn_type, ''),
				SUM(request_count)
			FROM silver.httptraffic_client_ip_5m
			WHERE COALESCE(endpoint_id, '') = $1 AND uri = $2
				AND window_start >= $3 AND window_start < $4
				AND asn IS NOT NULL
			GROUP BY asn, org_name, as_domain, asn_type`, ep.endpointID, ep.uri, baselineStart, windowStart)
		if err != nil {
			return nil, fmt.Errorf("query ASN baseline: %w", err)
		}
		baselineASN := make(map[int]int64)
		var baselineTotal int64
		var baselineHostingTotal int64
		for baselineRows.Next() {
			var asn int
			var orgName, asDomain, asnType string
			var total int64
			if err := baselineRows.Scan(&asn, &orgName, &asDomain, &asnType, &total); err != nil {
				baselineRows.Close()
				return nil, fmt.Errorf("scan ASN baseline: %w", err)
			}
			baselineASN[asn] = total
			baselineTotal += total
			if matchRules.IsHostingDomain(asDomain, asnType) {
				baselineHostingTotal += total
			}
		}
		if err := baselineRows.Err(); err != nil {
			baselineRows.Close()
			return nil, fmt.Errorf("iterate ASN baseline rows: %w", err)
		}
		baselineRows.Close()

		// Current window per ASN.
		currentRows, err := a.db.QueryContext(ctx, `
			SELECT asn, COALESCE(org_name, ''), COALESCE(as_domain, ''), COALESCE(asn_type, ''),
				SUM(request_count)
			FROM silver.httptraffic_client_ip_5m
			WHERE COALESCE(endpoint_id, '') = $1 AND uri = $2
				AND window_start >= $3 AND window_start < $4
				AND asn IS NOT NULL
			GROUP BY asn, org_name, as_domain, asn_type`, ep.endpointID, ep.uri, windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query current ASN: %w", err)
		}
		type asnEntry struct {
			asn      int
			orgName  string
			asDomain string
			asnType  string
			count    int64
		}
		var currentEntries []asnEntry
		var currentTotal int64
		var currentHostingTotal int64
		for currentRows.Next() {
			var e asnEntry
			if err := currentRows.Scan(&e.asn, &e.orgName, &e.asDomain, &e.asnType, &e.count); err != nil {
				currentRows.Close()
				return nil, fmt.Errorf("scan current ASN: %w", err)
			}
			currentEntries = append(currentEntries, e)
			currentTotal += e.count
			if matchRules.IsHostingDomain(e.asDomain, e.asnType) {
				currentHostingTotal += e.count
			}
		}
		if err := currentRows.Err(); err != nil {
			currentRows.Close()
			return nil, fmt.Errorf("iterate current ASN rows: %w", err)
		}
		currentRows.Close()

		if currentTotal == 0 {
			continue
		}

		baselineHostingShare := float64(0)
		if baselineTotal > 0 {
			baselineHostingShare = float64(baselineHostingTotal) / float64(baselineTotal)
		}

		newASNShareHigh := r.Threshold("new_asn_high", "share_min", 0.10)
		newASNShareWarning := r.Threshold("new_asn_warning", "share_min", 0.03)
		hostingShareMin := r.Threshold("hosting_provider", "share_min", 0.30)
		hostingBaselineMax := r.Threshold("hosting_provider", "baseline_max", 0.10)

		for _, e := range currentEntries {
			share := float64(e.count) / float64(currentTotal)

			// new_asn: ASN not in 7d baseline.
			if _, inBaseline := baselineASN[e.asn]; !inBaseline {
				if share > newASNShareHigh {
					resourceID := fmt.Sprintf("newasn:%s:%s:%d", ep.endpointID, windowStart.Format(time.RFC3339), e.asn)
					evidence, _ := json.Marshal(map[string]any{"asn": e.asn, "org": e.orgName, "domain": e.asDomain, "share": share, "count": e.count})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "new_asn", "high",
						windowStart, windowEnd, ep.uri, "", 0, share, 0,
						fmt.Sprintf("New ASN %d (%s) appeared with %.0f%% share", e.asn, e.orgName, share*100),
						detectedAt, evidence)
					result.NewASN++
				} else if share > newASNShareWarning {
					resourceID := fmt.Sprintf("newasn:%s:%s:%d", ep.endpointID, windowStart.Format(time.RFC3339), e.asn)
					evidence, _ := json.Marshal(map[string]any{"asn": e.asn, "org": e.orgName, "domain": e.asDomain, "share": share})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "new_asn", "low",
						windowStart, windowEnd, ep.uri, "", 0, share, 0,
						fmt.Sprintf("New ASN %d (%s) appeared with %.0f%% share", e.asn, e.orgName, share*100),
						detectedAt, evidence)
					result.NewASN++
				}
			}

			// hosting_provider: traffic from hosting domain/type, high share, baseline was low.
			if matchRules.IsHostingDomain(e.asDomain, e.asnType) {
				if share > hostingShareMin && baselineHostingShare < hostingBaselineMax {
					resourceID := fmt.Sprintf("hosting:%s:%s:%d", ep.endpointID, windowStart.Format(time.RFC3339), e.asn)
					evidence, _ := json.Marshal(map[string]any{"asn": e.asn, "org": e.orgName, "domain": e.asDomain, "share": share, "baseline_hosting_share": baselineHostingShare})
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "hosting_provider", "medium",
						windowStart, windowEnd, ep.uri, "", baselineHostingShare, share, 0,
						fmt.Sprintf("Hosting provider %s (ASN %d, %s) now %.0f%% of traffic (baseline: %.0f%%)", e.orgName, e.asn, e.asDomain, share*100, baselineHostingShare*100),
						detectedAt, evidence)
					result.HostingProvider++
				}
			}

			// asn_concentration: single ASN dominates, baseline was lower.
			asnConcentrationMin := r.Threshold("asn_concentration", "share_min", 0.60)
			asnConcentrationBaselineMax := r.Threshold("asn_concentration", "baseline_max", 0.30)
			if share > asnConcentrationMin && baselineTotal > 0 {
				baselineShare := float64(baselineASN[e.asn]) / float64(baselineTotal)
				if baselineShare < asnConcentrationBaselineMax {
					resourceID := fmt.Sprintf("asnconc:%s:%s:%d", ep.endpointID, windowStart.Format(time.RFC3339), e.asn)
					a.createAnomaly(ctx, resourceID, ep.endpointID, "", "asn_concentration", "medium",
						windowStart, windowEnd, ep.uri, "", baselineShare, share, 0,
						fmt.Sprintf("Single ASN %d (%s) sends %.0f%% of traffic", e.asn, e.orgName, share*100),
						detectedAt, nil)
					result.ASNConcentration++
				}
			}
		}
	}

	logger.Info("ASN anomaly detection complete",
		"newASN", result.NewASN, "hostingProvider", result.HostingProvider,
		"asnConcentration", result.ASNConcentration)
	return result, nil
}

// --- Activity 8: DetectAuthAnomalies ---

// DetectAuthAnomaliesResult holds output.
type DetectAuthAnomaliesResult struct {
	AuthFailureBurst      int
	CredentialStuffing    int
	OTPBruteForce         int
	PrivilegeEscalation   int
	PasswordResetFlood    int
	RegistrationAbuse     int
	RateLimitTriggered    int
	AuthSuccessAfterBurst int
}

// DetectAuthAnomalies detects authentication-related attacks using keyword-based
// URI matching against config.auth_endpoint_patterns.
func (a *Activities) DetectAuthAnomalies(ctx context.Context) (*DetectAuthAnomaliesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Detecting auth anomalies")

	r, err := a.getRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}

	rules, err := a.matchRules.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load match rules: %w", err)
	}

	now := time.Now()
	windowEnd := now.Truncate(5 * time.Minute)
	windowStart := windowEnd.Add(-5 * time.Minute)
	detectedAt := time.Now()
	result := &DetectAuthAnomaliesResult{}

	loginClause := rules.AuthPatternClause("login")
	tokenClause := rules.AuthPatternClause("token")
	otpClause := rules.AuthPatternClause("otp")
	adminClause := rules.AuthPatternClause("admin")
	resetClause := rules.AuthPatternClause("password_reset")
	registerClause := rules.AuthPatternClause("register")

	// Combine login + token for auth failure detection.
	authClause := "false"
	if loginClause != "false" && tokenClause != "false" {
		authClause = loginClause[:len(loginClause)-1] + " OR " + tokenClause[1:]
	} else if loginClause != "false" {
		authClause = loginClause
	} else if tokenClause != "false" {
		authClause = tokenClause
	}

	// 1. auth_failure_burst: Many 401/403 on login/token endpoints per IP.
	if authClause != "false" {
		authFailMin := r.ThresholdInt("auth_failure_burst", "count_min", 20)
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip, SUM(c.request_count) as fail_count
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND t.status_code IN (401, 403)
				AND %s
			GROUP BY c.source_id, c.client_ip
			HAVING SUM(c.request_count) > $3`, authClause), windowStart, windowEnd, authFailMin)
		if err != nil {
			return nil, fmt.Errorf("query auth failure burst: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP string
			var failCount int64
			if err := rows.Scan(&sourceID, &clientIP, &failCount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan auth failure burst: %w", err)
			}
			resourceID := fmt.Sprintf("authfail:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "fail_count": failCount})
			a.createAnomaly(ctx, resourceID, "", sourceID, "auth_failure_burst", "critical",
				windowStart, windowEnd, "", "", 0, float64(failCount), 0,
				fmt.Sprintf("IP %s: %d auth failures (401/403) on login endpoints", clientIP, failCount),
				detectedAt, evidence)
			result.AuthFailureBurst++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate auth failure burst rows: %w", err)
		}
		rows.Close()
	}

	// 2. credential_stuffing: Many DISTINCT login URIs with 401 per IP.
	if loginClause != "false" {
		credMin := r.ThresholdInt("credential_stuffing", "uri_count_min", 50)
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip, COUNT(DISTINCT t.uri) as uri_count
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND t.status_code = 401
				AND %s
			GROUP BY c.source_id, c.client_ip
			HAVING COUNT(DISTINCT t.uri) > $3`, loginClause), windowStart, windowEnd, credMin)
		if err != nil {
			return nil, fmt.Errorf("query credential stuffing: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP string
			var uriCount int64
			if err := rows.Scan(&sourceID, &clientIP, &uriCount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan credential stuffing: %w", err)
			}
			resourceID := fmt.Sprintf("credstuff:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "distinct_uris": uriCount})
			a.createAnomaly(ctx, resourceID, "", sourceID, "credential_stuffing", "critical",
				windowStart, windowEnd, "", "", 0, float64(uriCount), 0,
				fmt.Sprintf("IP %s: 401 on %d distinct login URIs (credential stuffing)", clientIP, uriCount),
				detectedAt, evidence)
			result.CredentialStuffing++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate credential stuffing rows: %w", err)
		}
		rows.Close()
	}

	// 3. otp_brute_force: Many requests to OTP/MFA endpoints per IP.
	if otpClause != "false" {
		otpMin := r.ThresholdInt("otp_brute_force", "count_min", 10)
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip, SUM(c.request_count) as req_count
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND %s
			GROUP BY c.source_id, c.client_ip
			HAVING SUM(c.request_count) > $3`, otpClause), windowStart, windowEnd, otpMin)
		if err != nil {
			return nil, fmt.Errorf("query otp brute force: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP string
			var reqCount int64
			if err := rows.Scan(&sourceID, &clientIP, &reqCount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan otp brute force: %w", err)
			}
			resourceID := fmt.Sprintf("otpbrute:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "count": reqCount})
			a.createAnomaly(ctx, resourceID, "", sourceID, "otp_brute_force", "critical",
				windowStart, windowEnd, "", "", 0, float64(reqCount), 0,
				fmt.Sprintf("IP %s: %d requests to OTP/MFA endpoints", clientIP, reqCount),
				detectedAt, evidence)
			result.OTPBruteForce++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate OTP brute force rows: %w", err)
		}
		rows.Close()
	}

	// 4. privilege_escalation_probe: External IP on admin/management paths.
	if adminClause != "false" {
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip, t.uri, SUM(c.request_count) as req_count
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND c.is_internal = false
				AND %s
			GROUP BY c.source_id, c.client_ip, t.uri`, adminClause), windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query privilege escalation: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP, uri string
			var reqCount int64
			if err := rows.Scan(&sourceID, &clientIP, &uri, &reqCount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan privilege escalation: %w", err)
			}
			resourceID := fmt.Sprintf("privesc:%s:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP, uri)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "uri": uri, "count": reqCount})
			a.createAnomaly(ctx, resourceID, "", sourceID, "privilege_escalation_probe", "high",
				windowStart, windowEnd, uri, "", 0, float64(reqCount), 0,
				fmt.Sprintf("External IP %s accessing admin path %s (%d requests)", clientIP, uri, reqCount),
				detectedAt, evidence)
			result.PrivilegeEscalation++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate privilege escalation rows: %w", err)
		}
		rows.Close()
	}

	// 5. password_reset_flood: Many requests to reset endpoints per IP.
	if resetClause != "false" {
		resetMin := r.ThresholdInt("password_reset_flood", "count_min", 10)
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip, SUM(c.request_count) as req_count
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND %s
			GROUP BY c.source_id, c.client_ip
			HAVING SUM(c.request_count) > $3`, resetClause), windowStart, windowEnd, resetMin)
		if err != nil {
			return nil, fmt.Errorf("query password reset flood: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP string
			var reqCount int64
			if err := rows.Scan(&sourceID, &clientIP, &reqCount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan password reset flood: %w", err)
			}
			resourceID := fmt.Sprintf("resetflood:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "count": reqCount})
			a.createAnomaly(ctx, resourceID, "", sourceID, "password_reset_flood", "high",
				windowStart, windowEnd, "", "", 0, float64(reqCount), 0,
				fmt.Sprintf("IP %s: %d requests to password reset endpoints", clientIP, reqCount),
				detectedAt, evidence)
			result.PasswordResetFlood++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate password reset flood rows: %w", err)
		}
		rows.Close()
	}

	// 6. registration_abuse: Many requests to register endpoints per IP.
	if registerClause != "false" {
		regMin := r.ThresholdInt("registration_abuse", "count_min", 10)
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip, SUM(c.request_count) as req_count
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND %s
			GROUP BY c.source_id, c.client_ip
			HAVING SUM(c.request_count) > $3`, registerClause), windowStart, windowEnd, regMin)
		if err != nil {
			return nil, fmt.Errorf("query registration abuse: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP string
			var reqCount int64
			if err := rows.Scan(&sourceID, &clientIP, &reqCount); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan registration abuse: %w", err)
			}
			resourceID := fmt.Sprintf("regabuse:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "count": reqCount})
			a.createAnomaly(ctx, resourceID, "", sourceID, "registration_abuse", "medium",
				windowStart, windowEnd, "", "", 0, float64(reqCount), 0,
				fmt.Sprintf("IP %s: %d requests to registration endpoints", clientIP, reqCount),
				detectedAt, evidence)
			result.RegistrationAbuse++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate registration abuse rows: %w", err)
		}
		rows.Close()
	}

	// 7. rate_limit_triggered: 429 responses per source IP (no auth pattern needed).
	rateLimitMin := r.ThresholdInt("rate_limit_triggered", "count_min", 5)
	rlRows, err := a.db.QueryContext(ctx, `
		SELECT c.source_id, c.client_ip, SUM(c.request_count) as count_429
		FROM silver.httptraffic_client_ip_5m c
		JOIN silver.httptraffic_traffic_5m t
			ON c.source_id = t.source_id AND c.window_start = t.window_start
			AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
		WHERE t.window_start >= $1 AND t.window_start < $2
			AND t.status_code = 429
		GROUP BY c.source_id, c.client_ip
		HAVING SUM(c.request_count) > $3`, windowStart, windowEnd, rateLimitMin)
	if err != nil {
		return nil, fmt.Errorf("query rate limit triggered: %w", err)
	}
	for rlRows.Next() {
		var sourceID, clientIP string
		var count429 int64
		if err := rlRows.Scan(&sourceID, &clientIP, &count429); err != nil {
			rlRows.Close()
			return nil, fmt.Errorf("scan rate limit triggered: %w", err)
		}
		resourceID := fmt.Sprintf("ratelimit:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
		evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "count_429": count429})
		a.createAnomaly(ctx, resourceID, "", sourceID, "rate_limit_triggered", "medium",
			windowStart, windowEnd, "", "", 0, float64(count429), 0,
			fmt.Sprintf("IP %s: %d rate-limited responses (429)", clientIP, count429),
			detectedAt, evidence)
		result.RateLimitTriggered++
	}
	if err := rlRows.Err(); err != nil {
		rlRows.Close()
		return nil, fmt.Errorf("iterate rate limit triggered rows: %w", err)
	}
	rlRows.Close()

	// 8. auth_success_after_burst: IP has both 401/403 AND 200 on auth endpoints.
	if authClause != "false" {
		authSuccessFailMin := r.ThresholdInt("auth_success_after_burst", "fail_min", 10)
		rows, err := a.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT c.source_id, c.client_ip,
				SUM(CASE WHEN t.status_code IN (401, 403) THEN c.request_count ELSE 0 END) as fails,
				SUM(CASE WHEN t.status_code BETWEEN 200 AND 299 THEN c.request_count ELSE 0 END) as successes
			FROM silver.httptraffic_client_ip_5m c
			JOIN silver.httptraffic_traffic_5m t
				ON c.source_id = t.source_id AND c.window_start = t.window_start
				AND c.uri = t.uri AND COALESCE(c.method, '') = COALESCE(t.method, '')
			WHERE t.window_start >= $1 AND t.window_start < $2
				AND %s
			GROUP BY c.source_id, c.client_ip
			HAVING SUM(CASE WHEN t.status_code IN (401, 403) THEN c.request_count ELSE 0 END) > $3
				AND SUM(CASE WHEN t.status_code BETWEEN 200 AND 299 THEN c.request_count ELSE 0 END) > 0`,
			authClause), windowStart, windowEnd, authSuccessFailMin)
		if err != nil {
			return nil, fmt.Errorf("query auth success after burst: %w", err)
		}
		for rows.Next() {
			var sourceID, clientIP string
			var fails, successes int64
			if err := rows.Scan(&sourceID, &clientIP, &fails, &successes); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan auth success after burst: %w", err)
			}
			resourceID := fmt.Sprintf("authsuccess:%s:%s:%s", sourceID, windowStart.Format(time.RFC3339), clientIP)
			evidence, _ := json.Marshal(map[string]any{"ip": clientIP, "fails": fails, "successes": successes})
			a.createAnomaly(ctx, resourceID, "", sourceID, "auth_success_after_burst", "critical",
				windowStart, windowEnd, "", "", float64(fails), float64(successes), 0,
				fmt.Sprintf("Possible brute force success: IP %s had %d auth failures then %d successes", clientIP, fails, successes),
				detectedAt, evidence)
			result.AuthSuccessAfterBurst++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterate auth success after burst rows: %w", err)
		}
		rows.Close()
	}

	logger.Info("Auth anomaly detection complete",
		"authFailure", result.AuthFailureBurst,
		"credStuffing", result.CredentialStuffing,
		"otpBrute", result.OTPBruteForce,
		"privEsc", result.PrivilegeEscalation,
		"resetFlood", result.PasswordResetFlood,
		"regAbuse", result.RegistrationAbuse,
		"rateLimit", result.RateLimitTriggered,
		"authSuccess", result.AuthSuccessAfterBurst)
	return result, nil
}

// --- Activity: CleanupStale ---

// CleanupStaleResult holds cleanup statistics.
type CleanupStaleResult struct {
	AnomaliesDeleted     int
	SilverTrafficDeleted int
	SilverUADeleted      int
	SilverIPDeleted      int
}

// CleanupStale removes old anomalies and silver traffic data beyond retention period.
func (a *Activities) CleanupStale(ctx context.Context) (*CleanupStaleResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Cleaning up stale data")

	retentionDays := a.configService.AccessLogRetentionDays()
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	// Delete old anomalies.
	anomalyResult, err := a.db.ExecContext(ctx,
		`DELETE FROM gold.httpmonitor_anomalies WHERE window_start < $1`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("delete old anomalies: %w", err)
	}
	anomaliesDeleted, _ := anomalyResult.RowsAffected()

	// Delete old silver traffic data.
	silverResult, err := a.db.ExecContext(ctx,
		`DELETE FROM silver.httptraffic_traffic_5m WHERE window_start < $1`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("delete old silver traffic: %w", err)
	}
	silverDeleted, _ := silverResult.RowsAffected()

	// Delete old silver user agent data.
	uaResult, err := a.db.ExecContext(ctx,
		`DELETE FROM silver.httptraffic_user_agent_5m WHERE window_start < $1`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("delete old silver user agents: %w", err)
	}
	uaDeleted, _ := uaResult.RowsAffected()

	// Delete old silver client IP data.
	ipResult, err := a.db.ExecContext(ctx,
		`DELETE FROM silver.httptraffic_client_ip_5m WHERE window_start < $1`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("delete old silver client IPs: %w", err)
	}
	ipDeleted, _ := ipResult.RowsAffected()

	logger.Info("Cleanup complete",
		"anomaliesDeleted", anomaliesDeleted,
		"silverTrafficDeleted", silverDeleted,
		"silverUADeleted", uaDeleted,
		"silverIPDeleted", ipDeleted)

	return &CleanupStaleResult{
		AnomaliesDeleted:     int(anomaliesDeleted),
		SilverTrafficDeleted: int(silverDeleted),
		SilverUADeleted:      int(uaDeleted),
		SilverIPDeleted:      int(ipDeleted),
	}, nil
}

// --- Helpers ---

func (a *Activities) createAnomaly(ctx context.Context, resourceID, endpointID, sourceID, anomalyType, severity string,
	windowStart, windowEnd time.Time, uri, method string,
	baselineValue, actualValue, deviation float64, description string, detectedAt time.Time,
	evidenceJSON json.RawMessage) {

	create := a.entClient.GoldHttpmonitorAnomaly.Create().
		SetID(resourceID).
		SetSourceID(sourceID).
		SetAnomalyType(anomalyType).
		SetSeverity(severity).
		SetWindowStart(windowStart).
		SetWindowEnd(windowEnd).
		SetDetectedAt(detectedAt).
		SetFirstDetectedAt(detectedAt)

	if endpointID != "" {
		create.SetEndpointID(endpointID)
	}
	if uri != "" {
		create.SetURI(uri)
	}
	if method != "" {
		create.SetMethod(method)
	}
	create.SetBaselineValue(baselineValue)
	create.SetActualValue(actualValue)
	if !math.IsNaN(deviation) && !math.IsInf(deviation, 0) {
		create.SetDeviation(deviation)
	}
	if description != "" {
		create.SetDescription(description)
	}
	if len(evidenceJSON) > 0 {
		create.SetEvidenceJSON(evidenceJSON)
	}

	if err := create.Exec(ctx); err != nil {
		// Skip duplicates.
		if !enthttpmonitor.IsConstraintError(err) {
			activity.GetLogger(ctx).Warn("Failed to create anomaly",
				"resourceID", resourceID, "error", err)
		}
	}
}
