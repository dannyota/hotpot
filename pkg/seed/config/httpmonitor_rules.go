package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type httpmonitorRule struct {
	ruleKey     string
	anomalyType string
	severity    string
	category    string
	name        string
	description string
	baseline    string
	thresholds  map[string]float64
	status      string // "live" or "planned"
}

// SeedHttpmonitorRules inserts all predefined detection rules.
// Planned rules are seeded with is_active=false.
func SeedHttpmonitorRules(ctx context.Context, db *sql.DB) error {
	if len(httpmonitorRules) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.httpmonitor_rules
		(rule_key, anomaly_type, severity, category, name, description, baseline, thresholds_json, status, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(httpmonitorRules)*12)
	for i, r := range httpmonitorRules {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 12
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12)

		var baseline *string
		if r.baseline != "" {
			baseline = &r.baseline
		}
		var threshJSON []byte
		if r.thresholds != nil {
			threshJSON, _ = json.Marshal(r.thresholds)
		}
		isActive := r.status == "live"

		args = append(args, r.ruleKey, r.anomalyType, r.severity, r.category,
			r.name, r.description, baseline, threshJSON,
			r.status, isActive, now, now)
	}

	b.WriteString(` ON CONFLICT (rule_key) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert httpmonitor rules (%d rules): %w", len(httpmonitorRules), err)
	}
	return nil
}

var httpmonitorRules = []httpmonitorRule{
	// --- Rate & Volume ---
	{"traffic_spike_high", "traffic_spike", "high", "rate", "Traffic Spike",
		"Z-score above 24h rolling avg triggers high severity", "24h rolling avg",
		map[string]float64{"z_score": 3}, "live"},
	{"traffic_spike_warning", "traffic_spike", "medium", "rate", "Traffic Spike",
		"Z-score above 24h rolling avg triggers medium severity", "24h rolling avg",
		map[string]float64{"z_score": 2}, "live"},
	{"traffic_drop", "traffic_drop", "high", "rate", "Traffic Drop",
		"Z-score below 24h rolling avg AND actual below fraction of avg", "24h rolling avg",
		map[string]float64{"z_score": -2, "min_pct": 0.1}, "live"},
	{"off_hours_spike", "off_hours_spike", "medium", "rate", "Off-Hours Spike",
		"Traffic exceeds multiplier of baseline during 00:00-05:00 local time", "Time-of-day windowing",
		map[string]float64{"multiplier": 3}, "live"},

	// --- Error Patterns ---
	{"5xx_burst", "5xx_burst", "high", "error", "5xx Burst",
		"5xx error rate exceeds threshold AND count exceeds minimum", "Per 5-min window",
		map[string]float64{"rate_min": 0.05, "count_min": 5}, "live"},
	{"error_burst", "error_burst", "medium", "error", "Error Burst",
		"4xx+5xx error rate exceeds threshold AND total count exceeds minimum", "Per 5-min window",
		map[string]float64{"rate_min": 0.05, "count_min": 10}, "live"},

	// --- Suspicious Patterns ---
	{"scanner_detected", "scanner_detected", "medium", "suspicious", "Scanner Detected",
		"Security scanner UA detected (sqlmap, nikto, etc.)", "Config rules",
		nil, "live"},
	{"single_ip_flood", "single_ip_flood", "medium", "suspicious", "Single IP Flood",
		"Single IP sending excessive requests in 5 min window", "Per window",
		map[string]float64{"count_min": 50}, "live"},
	{"method_mismatch_warning", "method_mismatch", "medium", "suspicious", "Method Mismatch",
		"Wrong HTTP method on endpoint, count exceeds minimum", "Endpoint schema",
		map[string]float64{"count_min": 10}, "live"},
	{"method_mismatch_high", "method_mismatch", "high", "suspicious", "Method Mismatch",
		"Wrong HTTP method on endpoint, count exceeds high threshold", "Endpoint schema",
		map[string]float64{"count_min": 50}, "live"},

	// --- Endpoint Discovery ---
	{"new_endpoint", "new_endpoint", "info", "endpoint", "New Endpoint",
		"Unmapped URI with significant traffic in 1 hour", "Endpoint inventory",
		map[string]float64{"count_min": 100}, "live"},
	{"endpoint_enumeration", "endpoint_enumeration", "high", "endpoint", "Endpoint Enumeration",
		"Many distinct URIs returning 404 per IP in 5 min", "Per source IP",
		map[string]float64{"count_min": 30}, "live"},

	// --- User-Agent Analysis ---
	{"new_user_agent_high", "new_user_agent", "high", "ua", "New User Agent",
		"New UA family with high share of current window", "7-day UA baseline",
		map[string]float64{"share_min": 0.20}, "live"},
	{"new_user_agent_warning", "new_user_agent", "low", "ua", "New User Agent",
		"New UA family with moderate share of current window", "7-day UA baseline",
		map[string]float64{"share_min": 0.05}, "live"},
	{"ua_share_shift", "ua_share_shift", "medium", "ua", "UA Share Shift",
		"UA family share changed significantly vs 7-day baseline", "7-day UA baseline",
		map[string]float64{"share_delta": 0.30}, "live"},
	{"automated_client", "automated_client", "medium", "ua", "Automated Client",
		"Library UA on protected endpoint", "Config rules",
		nil, "live"},
	{"ua_spoofing", "ua_spoofing", "medium", "ua", "UA Spoofing",
		"Claims browser UA but high request rate with no static asset requests", "Cross-field correlation",
		nil, "live"},

	// --- Client IP & Geo ---
	{"new_source_ip_high", "new_source_ip", "high", "ip_geo", "New Source IP",
		"New IP with high share of current window", "7-day IP baseline",
		map[string]float64{"share_min": 0.20}, "live"},
	{"new_source_ip_warning", "new_source_ip", "low", "ip_geo", "New Source IP",
		"New IP with moderate share of current window", "7-day IP baseline",
		map[string]float64{"share_min": 0.05}, "live"},
	{"ip_concentration", "ip_concentration", "medium", "ip_geo", "IP Concentration",
		"Single IP dominates traffic share in window", "Per window",
		map[string]float64{"share_min": 0.50}, "live"},
	{"geo_shift_new_country_high", "geo_shift", "high", "ip_geo", "Geo Shift (New Country)",
		"New country with high share of current window", "7-day country baseline",
		map[string]float64{"share_min": 0.20}, "live"},
	{"geo_shift_new_country", "geo_shift", "medium", "ip_geo", "Geo Shift (New Country)",
		"New country with moderate share of current window", "7-day country baseline",
		map[string]float64{"share_min": 0.10}, "live"},
	{"geo_shift_existing", "geo_shift", "medium", "ip_geo", "Geo Shift (Existing)",
		"Existing country share changed significantly vs 7-day baseline", "7-day country baseline",
		map[string]float64{"shift_delta": 0.20}, "live"},
	{"external_on_internal", "external_on_internal", "high", "ip_geo", "External on Internal",
		"External IP on endpoint with high internal baseline", "7-day internal ratio",
		map[string]float64{"internal_baseline_min": 0.90}, "live"},
	{"sanctioned_country", "sanctioned_country", "critical", "ip_geo", "Sanctioned Country",
		"Traffic from OFAC/sanctioned country", "Country blocklist config",
		nil, "live"},
	{"tor_exit_node", "tor_exit_node", "high", "ip_geo", "Tor Exit Node",
		"Traffic from known Tor exit node IPs", "Tor exit list feed",
		nil, "planned"},
	{"vpn_proxy_detected", "vpn_proxy_detected", "medium", "ip_geo", "VPN/Proxy Detected",
		"Traffic from known VPN/proxy provider", "VPN/proxy IP list",
		nil, "planned"},
	{"impossible_travel", "impossible_travel", "critical", "ip_geo", "Impossible Travel",
		"Same session/user from 2+ countries within time window", "Session/user tracking",
		map[string]float64{"time_window_min": 30}, "planned"},
	{"ip_rotation", "ip_rotation", "medium", "ip_geo", "IP Rotation",
		"Many distinct IPs from same /24 subnet hitting same URI in 5 min", "Subnet aggregation",
		map[string]float64{"ip_count_min": 10}, "live"},

	// --- ASN & Network ---
	{"new_asn_high", "new_asn", "high", "asn", "New ASN",
		"New ASN with high share of current window", "7-day ASN baseline",
		map[string]float64{"share_min": 0.10}, "live"},
	{"new_asn_warning", "new_asn", "low", "asn", "New ASN",
		"New ASN with moderate share of current window", "7-day ASN baseline",
		map[string]float64{"share_min": 0.03}, "live"},
	{"hosting_provider", "hosting_provider", "medium", "asn", "Hosting Provider",
		"Hosting/cloud ASN share exceeds threshold while baseline was low", "7-day hosting share",
		map[string]float64{"share_min": 0.30, "baseline_max": 0.10}, "live"},
	{"asn_concentration", "asn_concentration", "medium", "asn", "ASN Concentration",
		"Single ASN dominates traffic while baseline was lower", "7-day ASN baseline",
		map[string]float64{"share_min": 0.60, "baseline_max": 0.30}, "live"},

	// --- Authentication & Credential Abuse ---
	{"auth_failure_burst", "auth_failure_burst", "critical", "auth", "Auth Failure Burst",
		"Excessive 401/403 on auth endpoints per source IP in 5 min", "Auth endpoint keywords",
		map[string]float64{"count_min": 20}, "live"},
	{"credential_stuffing", "credential_stuffing", "critical", "auth", "Credential Stuffing",
		"Many distinct auth URIs with 401 from one IP in 5 min", "Auth endpoint keywords",
		map[string]float64{"uri_count_min": 50}, "live"},
	{"otp_brute_force", "otp_brute_force", "critical", "auth", "OTP Brute Force",
		"Excessive requests to OTP/MFA endpoints per IP in 5 min", "Auth endpoint keywords",
		map[string]float64{"count_min": 10}, "live"},
	{"privilege_escalation_probe", "privilege_escalation_probe", "high", "auth", "Privilege Escalation Probe",
		"Non-internal IP accessing admin/management paths", "Auth endpoint keywords",
		nil, "live"},
	{"password_reset_flood", "password_reset_flood", "high", "auth", "Password Reset Flood",
		"Many requests to password reset endpoints per IP in 5 min", "Auth endpoint keywords",
		map[string]float64{"count_min": 10}, "live"},
	{"registration_abuse", "registration_abuse", "medium", "auth", "Registration Abuse",
		"Many requests to registration endpoints per IP in 5 min", "Auth endpoint keywords",
		map[string]float64{"count_min": 10}, "live"},
	{"rate_limit_triggered", "rate_limit_triggered", "medium", "rate", "Rate Limit Triggered",
		"IP receiving 429 responses (rate limit hit)", "Per window",
		map[string]float64{"count_min": 5}, "live"},
	{"auth_success_after_burst", "auth_success_after_burst", "critical", "auth", "Auth Success After Burst",
		"IP has auth failures then success — possible brute force success", "Auth endpoint keywords",
		map[string]float64{"fail_min": 10}, "live"},

	// --- Data Exfiltration ---
	{"response_size_anomaly", "response_size_anomaly", "high", "data", "Response Size Anomaly",
		"Response body bytes Z-score exceeds threshold vs 24h baseline", "24h rolling avg",
		map[string]float64{"z_score": 3}, "live"},
	{"bulk_data_extraction", "bulk_data_extraction", "high", "data", "Bulk Data Extraction",
		"Single client downloading excessive data in 5 min (unique_client_count=1)", "Per window",
		map[string]float64{"bytes_min": 10485760}, "live"},
	{"pagination_scraping", "pagination_scraping", "medium", "data", "Pagination Scraping",
		"Single IP requesting many query variants of same base URI in 5 min", "URI param analysis",
		map[string]float64{"count_min": 50}, "live"},

	// --- Injection & Path Traversal ---
	{"path_traversal", "path_traversal", "critical", "injection", "Path Traversal",
		"URI contains ../, %2e%2e, or encoded traversal patterns", "URI pattern flags (CRS 930)",
		nil, "live"},
	{"sql_injection_probe", "sql_injection_probe", "critical", "injection", "SQL Injection Probe",
		"URI contains SQL injection patterns", "URI pattern flags (CRS 942)",
		nil, "live"},
	{"command_injection_probe", "command_injection_probe", "critical", "injection", "Command Injection Probe",
		"URI contains command injection patterns", "URI pattern flags (CRS 932)",
		nil, "live"},
	{"xss_probe", "xss_probe", "high", "injection", "XSS Probe",
		"URI contains cross-site scripting patterns", "URI pattern flags (CRS 941)",
		nil, "live"},
	{"ssrf_probe", "ssrf_probe", "critical", "injection", "SSRF Probe",
		"URI contains server-side request forgery patterns", "URI pattern flags (CRS 934)",
		nil, "live"},

	// --- Evasion & Fingerprinting ---
	{"tls_fingerprint_mismatch", "tls_fingerprint_mismatch", "info", "evasion", "TLS Fingerprint Mismatch",
		"UA claims mobile app but TLS fingerprint does not match", "TLS fingerprint data",
		nil, "planned"},
}
