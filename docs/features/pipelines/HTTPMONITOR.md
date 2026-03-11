# HTTP Monitor

Real-time API traffic anomaly detection for banking production systems.

## 🎯 Overview

```
  Traffic Types          Log Sources (role)              Pipeline (every 5 min)
  =============          ==================              ======================

                         PRIMARY sources
                         (count traffic, own the numbers)
                         ┌─────────────────────────┐
  Mobile Apps ──┐        │  Nginx, HAProxy, etc.    │
  Web Pages ────┤        │  Real client IP, UA      │     ┌──────────────────────┐
  Partner APIs ─┤ ──────►│  Full request details    │────►│  Bronze: accesslog   │
  Protected ────┘        └─────────────────────────┘     │  5-min aggregates    │
  APIs                                                    │  counts, UAs, IPs    │
                         ENRICHMENT sources               └──────────┬───────────┘
                         (add fields, don't count)                   │
                         ┌─────────────────────────┐        normalize & merge
                         │  Kong, API gateways      │                │
                         │  Partner ID, auth info   │     ┌──────────▼───────────┐
                         │  No duplicate counting   │────►│  Silver: httptraffic  │
                         └─────────────────────────┘     │  + GeoIP & ASN        │
                                                          │  + endpoint mapping   │
                           ~1M+ logs/day                  │  + scanner flags      │
                           per environment                │  + UA classification  │
                           multi-cloud                    │  + partner ID         │
                                                          │  (merged from         │
                                                          │   enrichment sources) │
                                                          └──────────┬───────────┘
                                                                     │
                                                           50 detection rules
                                                          46 live, 4 planned
                                                           7-day rolling baseline
                                                                     │
                                                          ┌──────────▼───────────┐
                                                          │  Gold: httpmonitor    │
                                                          │                       │
  ~1M logs ──► ~5K anomalies ──► ~50 alerts               │  Anomalies ── Alerts  │
  (raw)        (signals)          (actionable)            │  (per rule)   (per    │
                                                          │               actor)  │
                                                          └──────────┬───────────┘
                                                                     │
                                                                     ▼
                                                           Analyst Alert Queue
                                                           ====================
                                                           One alert per attacker
                                                           Auto-updates while
                                                           attack continues.
                                                           Auto-resolves when
                                                           it stops.
```

What it detects (46 rules live, 4 planned):

| Category | Detects |
|----------|---------|
| Rate & Errors | Traffic spikes/drops, 5xx bursts, error surges |
| Suspicious Actors | Scanners, single-IP floods, hosting providers, Tor/VPN |
| Identity Shifts | New IPs, new ASNs, geo shifts, UA changes |
| Auth Abuse | Credential stuffing, OTP brute force, privilege probing |
| Data Protection | Response size anomalies, bulk extraction, scraping |
| Injection | Path traversal, SQL injection, command injection |

## 🔀 Multi-Source Support (planned — Phase 1)

### Source Roles

Each gateway instance is a **log source** (`accesslog_log_sources`). Each source will have a **role** that determines how it contributes to the pipeline.

| Role | Counts Traffic? | Purpose | Example |
|------|:---------------:|---------|---------|
| `primary` | ✅ Yes | Owns request counts, client IPs, UAs | Nginx, HAProxy |
| `enrichment` | ❌ No | Adds fields only, no duplicate counting | Kong, API gateways |

When a request passes through multiple gateways, each gateway logs it. Without roles, the same request would be counted multiple times. With roles:

- **Primary** sources own the traffic numbers — request counts, client IPs, user agents, response times
- **Enrichment** sources contribute extra fields (partner ID, auth context) that are **merged** into silver records by matching on `(window, uri, method)`

```
  Primary source (nginx)               Enrichment source (kong)
  ══════════════════════               ═════════════════════════
  request_count: 150                   partner_id: "fintech-abc"
  client_ip: 103.28.x.x               auth_context: "cert-xyz"
  user_agent: "Dart/3.7"
  status: 200
           │                                      │
           └──────────────┬───────────────────────┘
                          │  merge by (window, uri, method)
                          ▼
              Silver record
              ═════════════
              request_count: 150        ← from primary
              client_ip: 103.28.x.x    ← from primary
              partner_id: "fintech-abc" ← from enrichment
              country: Vietnam          ← GeoIP on primary IP
```

Each source has its own ingestion cursor, field mapping, and collection interval. Adding a new cloud, region, or gateway is one config row.

### Cross-Source Correlation

Detection runs across all sources per environment. Silver normalization produces a unified schema regardless of source type or cloud. Alert correlation groups anomalies by **client IP** across sources — an attack spanning multiple gateways or clouds still produces one alert per attacker.

## 🏗️ Data Model

Two-tier detection model. All tables in `gold` schema.

```
  ┌─────────────────────────────────────────────────────────────────┐
  │ gold.httpmonitor_anomalies              gold.httpmonitor_alerts  │
  │ ══════════════════════════              ══════════════════════   │
  │                                                                 │
  │  anomaly 1 ─┐                      ┌─ Alert: "Credential       │
  │  anomaly 2 ─┤  grouped by          │   stuffing from           │
  │  anomaly 3 ─┤  actor (IP,  ───────►│   45.33.12.8"             │
  │  anomaly 4 ─┤  partner,            │                           │
  │  anomaly 5 ─┘  subnet)             │  types: [auth_failure,    │
  │                                     │    new_source_ip,         │
  │  Per detection type                 │    hosting_provider,      │
  │  Per 5-min window                   │    geo_shift]             │
  │  High volume                        │  severity: critical       │
  │  (~5K/day)                          │  status: open             │
  │                                     │  living document          │
  │                                     │  (~50/day)                │
  │                                     └──────────────────────     │
  └─────────────────────────────────────────────────────────────────┘
```

### Anomalies (✅ exists)

Table: `gold.httpmonitor_anomalies`

One row per detection per 5-min window. High volume, many are noise.

| Field | Type | Purpose |
|-------|------|---------|
| `resource_id` | PK | Deterministic: `{type}:{source}:{window}:{uri}:{method}` |
| `alert_id` | FK, optional, planned | Link to parent alert (null until correlated) |
| `anomaly_type` | enum | One of 50 detection types (46 live) |
| `severity` | enum | critical / high / medium / low / info |
| `source_id` | string | Log source identifier |
| `endpoint_id` | string, optional | Matched API endpoint |
| `window_start/end` | timestamp | 5-min detection window |
| `uri`, `method` | string, optional | HTTP request details |
| `baseline_value` | float, optional | Historical baseline |
| `actual_value` | float, optional | Current observed value |
| `deviation` | float, optional | Z-score or ratio |
| `description` | string | Human-readable explanation |
| `evidence_json` | JSON, optional | Structured evidence |

### Alerts (planned)

Table: `gold.httpmonitor_alerts`

One alert per **actor** — groups all anomaly types from the same source. Living document that keeps updating as the attack continues.

| Field | Type | Purpose |
|-------|------|---------|
| `resource_id` | PK | Auto-generated |
| `group_key` | string, unique | Actor key (see scopes below) |
| `scope` | enum | `ip`, `subnet`, `partner`, `endpoint` |
| `severity` | enum | Max of linked anomalies, auto-escalates |
| `status` | enum | `open` → `acknowledged` → `resolved` / `false_positive` / `suppressed` |
| `source_id` | string | Log source |
| `client_ip` | string, optional | Attacker IP (for ip/subnet scope) |
| `partner_id` | string, optional | Partner identifier (for partner scope) |
| `endpoint_id` | string, optional | Target endpoint (for endpoint scope) |
| `anomaly_types` | string[] | All distinct anomaly types seen |
| `anomaly_count` | int | Total anomalies linked |
| `affected_endpoints` | string[] | All endpoints involved |
| `affected_ips` | string[] | All IPs involved (for subnet scope) |
| `countries` | string[] | All countries seen |
| `asns` | string[] | All ASNs seen |
| `first_seen` | timestamp | Earliest anomaly window |
| `last_seen` | timestamp | Latest anomaly window |
| `title` | string | Auto-generated, updates as new types appear |
| `description` | string | Aggregated context |
| `timeline_json` | JSON | Chronological event log (see below) |

**Alert scopes** — group key by actor type:

| Scope | Group Key | Example |
|-------|-----------|---------|
| `ip` | `ip:{source_id}:{client_ip}` | All anomalies from one attacker IP |
| `subnet` | `net:{source_id}:{ip/24}` | Distributed attack from same /24 range |
| `partner` | `partner:{partner_id}` | All anomalies from one partner |
| `endpoint` | `ep:{source_id}:{endpoint_id}` | Aggregate anomalies (rate, errors) without specific IP |

**Living updates** — every 5-min cycle, `CorrelateAlerts` does:

1. Find new anomalies without `alert_id`
2. Match to existing open alert by `group_key` — if found, **update**:
   - Append new types to `anomaly_types`
   - Update `last_seen`
   - Recalculate `severity` (max of all linked anomalies)
   - Add entry to `timeline_json`
   - Update `anomaly_count`, `affected_endpoints`, `countries`, `asns`
   - Regenerate `title` and `description`
3. If no matching open alert — create new alert
4. Close stale alerts: `last_seen` > 30 min ago → status = `resolved` (auto)

**Alert lifecycle**:

```
                   new anomaly,                     analyst
                   no matching alert                picks up
                         │                             │
                         ▼                             ▼
                  ┌────────────┐              ┌──────────────┐
  new anomalies   │            │              │              │   new anomalies
  added (living ──│    open    │─────────────►│ acknowledged │── added (living
  update)         │            │              │              │   update)
                  └─────┬──────┘              └──┬───┬───┬───┘
                        │                        │   │   │
              no new anomalies            analyst │   │   │ analyst
              for 30 min (auto)           closes  │   │   │ marks FP
                        │                         │   │   │
                        ▼                         ▼   │   ▼
                  ┌────────────┐                      │  ┌───────────────┐
                  │  resolved  │◄─────────────────────┘  │ false_positive│
                  └────────────┘                         └───────────────┘
                                                         ┌───────────────┐
                                           analyst ─────►│  suppressed   │
                                           suppresses    └───────────────┘
```

**Auto-escalation** — severity bumps as the attack evolves:

| Condition | Severity |
|-----------|:--------:|
| Single anomaly type | Original severity |
| 2+ anomaly types | max(severities) |
| 3+ types including any high | critical |
| Any critical anomaly | critical |

**Example — Credential Stuffing Attack evolving over 30 min**:

```
  Time   Event                                          Severity   Anomaly Count
  ─────  ─────────────────────────────────────────────  ─────────  ─────────────
  02:00  Alert CREATED — auth_failure_burst             critical          1
         + new_source_ip — IP 45.33.12.8 first seen                      2
         + hosting_provider — DigitalOcean                                3
  02:05  + geo_shift — Romania at 40% share             critical          6
         + automated_client — python-requests on /auth                    8
  02:10  + credential_stuffing — 80 distinct auth URIs  critical         12
  02:15  attack continues, more anomalies added         critical         18
  02:20  attack continues                               critical         24
  02:25  last anomaly seen                              critical         26
   ...   (no new anomalies)
  02:55  AUTO-RESOLVED — no activity for 30 min         resolved         26
```

### Cases (optional, manual)

Analysts create cases when they want to:
- Link multiple alerts into one investigation (e.g., IP alert + endpoint alert = same incident)
- Track resolution for audit/compliance
- Assign to a team member

Cases are created manually from the alert UI. No auto-creation — alerts are the primary work surface.

### Workflow Integration

```
Detection Workflow (every 5 min)
  │
  │  Anomaly Detection (✅ exists)
  │  ─────────────────────────────
  ├─ 1. DetectRateAnomalies
  ├─ 2. DetectErrorBursts
  ├─ 3. DetectSuspiciousPatterns
  ├─ 4. DetectMethodMismatch
  ├─ 5. DetectNewEndpoints
  ├─ 6. DetectUserAgentAnomalies
  ├─ 7. DetectClientIPAnomalies
  ├─ 8. DetectASNAnomalies
  ├─ 9. DetectAuthAnomalies
  │
  │  Alert Correlation (planned — Phase 1)
  │  ─────────────────────────────────────
  ├─ 10. CorrelateAlerts          create new / update existing alerts
  │
  │  Housekeeping (✅ exists)
  │  ────────────────────────
  └─ 11. CleanupStale            auto-resolve stale alerts, purge old data
```

## 📊 Detection Rules Catalog

All rules are stored in `config.httpmonitor_rules` with a unique `rule_key` for code lookups and an integer `rule_id` for user-facing references. Severity follows CVE convention: **critical > high > medium > low > info**.

### Rate & Volume

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `traffic_spike_warning` | `traffic_spike` | medium | Z-score > 2 | 24h rolling avg |
| ✅ | `traffic_spike_high` | `traffic_spike` | high | Z-score > 3 | 24h rolling avg |
| ✅ | `traffic_drop` | `traffic_drop` | high | Z < -2 AND actual < 10% of avg | 24h rolling avg |
| ✅ | `off_hours_spike` | `off_hours_spike` | medium | Traffic > 3x baseline during 00:00–05:00 local | Time-of-day windowing |

### Error Patterns

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `5xx_burst` | `5xx_burst` | high | 5xx rate > 5% AND count > 5 | Per window |
| ✅ | `error_burst` | `error_burst` | medium | Error rate > 5% AND count > 10 | Per window |

### Suspicious Patterns

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `scanner_detected` | `scanner_detected` | medium | Scanner UA found (sqlmap, nikto, etc.) | Config rules |
| ✅ | `single_ip_flood` | `single_ip_flood` | medium | 1 IP sending > 50 req in 5 min | Per window |
| ✅ | `method_mismatch_warning` | `method_mismatch` | medium | Wrong HTTP method, > 10 requests | Endpoint schema |
| ✅ | `method_mismatch_high` | `method_mismatch` | high | Wrong HTTP method, > 50 requests | Endpoint schema |

### Endpoint Discovery

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `new_endpoint` | `new_endpoint` | info | Unmapped URI, > 100 req/hour | Endpoint inventory |
| ✅ | `endpoint_enumeration` | `endpoint_enumeration` | high | > 30 distinct URIs returning 404 per IP in 5 min | Per source IP |

### User-Agent Analysis

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `new_user_agent_warning` | `new_user_agent` | low | New UA family, > 5% share | 7-day UA baseline |
| ✅ | `new_user_agent_high` | `new_user_agent` | high | New UA family, > 20% share | 7-day UA baseline |
| ✅ | `ua_share_shift` | `ua_share_shift` | medium | UA share changed > ±30 pct pts | 7-day UA baseline |
| ✅ | `automated_client` | `automated_client` | medium | Library UA on protected endpoint | Config rules |
| ✅ | `ua_spoofing` | `ua_spoofing` | medium | Browser UA + high request rate + no static asset requests | Cross-field correlation |

### Client IP & Geo

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `new_source_ip_warning` | `new_source_ip` | low | New IP, > 5% share | 7-day IP baseline |
| ✅ | `new_source_ip_high` | `new_source_ip` | high | New IP, > 20% share | 7-day IP baseline |
| ✅ | `ip_concentration` | `ip_concentration` | medium | Single IP > 50% of traffic | Per window |
| ✅ | `geo_shift_new_country` | `geo_shift` | medium | New country > 10% share | 7-day country baseline |
| ✅ | `geo_shift_new_country_high` | `geo_shift` | high | New country > 20% share | 7-day country baseline |
| ✅ | `geo_shift_existing` | `geo_shift` | medium | Existing country share shift > ±20 pct pts | 7-day country baseline |
| ✅ | `external_on_internal` | `external_on_internal` | high | External IP on internal-only endpoint | 7-day internal ratio |
| ✅ | `sanctioned_country` | `sanctioned_country` | critical | Traffic from OFAC/sanctioned country | Country blocklist config |
| | `tor_exit_node` | `tor_exit_node` | high | Traffic from known Tor exit node IPs | Tor exit list feed |
| | `vpn_proxy_detected` | `vpn_proxy_detected` | medium | Traffic from known VPN/proxy provider | VPN/proxy IP list |
| | `impossible_travel` | `impossible_travel` | critical | Same session/user from 2+ countries within 30 min | Session/user tracking |
| ✅ | `ip_rotation` | `ip_rotation` | medium | > 10 distinct IPs from same /24 hitting same URI in 5 min | Subnet aggregation |

### ASN & Network

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `new_asn_warning` | `new_asn` | low | New ASN, > 3% share | 7-day ASN baseline |
| ✅ | `new_asn_high` | `new_asn` | high | New ASN, > 10% share | 7-day ASN baseline |
| ✅ | `hosting_provider` | `hosting_provider` | medium | Hosting share > 30%, baseline < 10% | 7-day + config rules |
| ✅ | `asn_concentration` | `asn_concentration` | medium | Single ASN > 60%, baseline < 30% | 7-day ASN baseline |

### Authentication & Credential Abuse

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `auth_failure_burst` | `auth_failure_burst` | critical | > 20 401/403 on login/token endpoints per IP in 5 min | Auth endpoint keywords |
| ✅ | `credential_stuffing` | `credential_stuffing` | critical | > 50 distinct login URIs with 401 from one IP in 5 min | Auth endpoint keywords |
| ✅ | `otp_brute_force` | `otp_brute_force` | critical | > 10 requests to OTP/MFA endpoints per IP in 5 min | Auth endpoint keywords |
| ✅ | `privilege_escalation_probe` | `privilege_escalation_probe` | high | Non-internal IP accessing admin/management paths | Auth endpoint keywords |
| ✅ | `password_reset_flood` | `password_reset_flood` | high | > 10 requests to password reset endpoints per IP in 5 min | Auth endpoint keywords |
| ✅ | `registration_abuse` | `registration_abuse` | medium | > 10 requests to registration endpoints per IP in 5 min | Auth endpoint keywords |
| ✅ | `rate_limit_triggered` | `rate_limit_triggered` | medium | IP receiving > 5 429 responses in 5 min | Per window |
| ✅ | `auth_success_after_burst` | `auth_success_after_burst` | critical | IP has > 10 auth failures then success — possible brute force win | Auth endpoint keywords |

### Data Exfiltration

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `response_size_anomaly` | `response_size_anomaly` | high | `total_body_bytes_sent` Z-score > 3 vs 24h baseline | Field exists in silver |
| ✅ | `bulk_data_extraction` | `bulk_data_extraction` | high | Single client downloading > 10 MB in 5 min (unique_client_count=1) | Per window |
| ✅ | `pagination_scraping` | `pagination_scraping` | medium | > 50 distinct query variants of same base URI per IP in 5 min | URI param analysis |

### Injection & Attack Patterns

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| ✅ | `path_traversal` | `path_traversal` | critical | URI contains `../`, `%2e%2e`, encoded traversal, sensitive file paths | CRS 930 patterns |
| ✅ | `sql_injection_probe` | `sql_injection_probe` | critical | URI contains SQL injection patterns (UNION, stacked queries, blind) | CRS 942 patterns |
| ✅ | `command_injection_probe` | `command_injection_probe` | critical | URI contains shell operators + commands, encoded variants | CRS 932 patterns |
| ✅ | `xss_probe` | `xss_probe` | high | URI contains XSS patterns (script tags, event handlers, JS URIs) | CRS 941 patterns |
| ✅ | `ssrf_probe` | `ssrf_probe` | critical | URI contains SSRF patterns (metadata IPs, internal hostnames) | CRS 934 patterns |

### Evasion & Fingerprinting

| Status | Rule Key | Type | Severity | Trigger | Baseline |
|:------:|----------|------|:--------:|---------|----------|
| | `tls_fingerprint_mismatch` | `tls_fingerprint_mismatch` | info | UA claims mobile app but TLS fingerprint mismatches | TLS fingerprint data |

## 📊 Coverage Summary

| Category | Done | Planned | Total |
|----------|:----:|:-------:|:-----:|
| Rate / Volume | 4 | — | 4 |
| Error Patterns | 2 | — | 2 |
| Suspicious Patterns | 4 | — | 4 |
| Endpoint Discovery | 2 | — | 2 |
| User-Agent Analysis | 5 | — | 5 |
| Client IP / Geo | 9 | 3 | 12 |
| ASN / Network | 4 | — | 4 |
| Auth / Credential Abuse | 8 | — | 8 |
| Data Exfiltration | 3 | — | 3 |
| Injection / Attack Patterns | 5 | — | 5 |
| Evasion / Fingerprint | — | 1 | 1 |
| **Total** | **46** | **4** | **50** |

## ⚙️ Config-Driven Rules

Stored in `config` schema. Detection rules are loaded at workflow start; match rules are cached 5 minutes, hot-reloadable.

| Table | Purpose | Key | Example Values |
|-------|---------|-----|----------------|
| `httpmonitor_rules` | Detection rule catalog (50 rules) | `rule_key` (unique) | `traffic_spike_high`, `geo_shift_existing` |
| `hosting_indicators` | Cloud/hosting provider detection | `(indicator_type, value)` | amazon.com, cloudflare, "vps" |
| `scanner_patterns` | Security scanner tools | `keyword` | sqlmap, nikto, nmap |
| `library_uas` | Automated HTTP clients | `family` | curl, wget, python-requests |
| `uri_attack_patterns` | URI injection/traversal/SSRF patterns | `(pattern_type, pattern)` | `/../`, `UNION SELECT`, `169.254.169.254` |
| `auth_endpoint_patterns` | Auth endpoint URI keywords | `(pattern_type, pattern)` | `login`, `otp`, `reset-password`, `admin` |

The `httpmonitor_rules` table is the full rule inventory — live and planned rules, with configurable thresholds, severity, active/inactive toggle, and status (live/planned/deprecated). Detection activities look up thresholds by `rule_key` and skip inactive rules. Seed data is defined in `pkg/seed/config/` and applied by `cmd/migrate` on every run (additive — `ON CONFLICT DO NOTHING`).

## 🚀 Roadmap

Detection-first approach: complete all anomaly detection rules, then add alert correlation on top.

### Phase 1 — Quick Win Rules (done)

Four rules added with no new infrastructure: `response_size_anomaly`, `off_hours_spike`, `endpoint_enumeration`, `sanctioned_country` (with new `config.sanctioned_countries` table + OFAC seed).

### Phase 2 — Injection & URI Pattern Detection (done)

Config table `config.uri_attack_patterns` with 87 patterns derived from OWASP CRS. Five boolean flags on silver (`is_lfi_detected`, `is_sqli_detected`, `is_rce_detected`, `is_xss_detected`, `is_ssrf_detected`) set during normalization. Five detection rules: `path_traversal`, `sql_injection_probe`, `command_injection_probe`, `xss_probe`, `ssrf_probe`.

### Phase 3 — Auth & Credential Detection (done)

Config table `config.auth_endpoint_patterns` with 31 keyword patterns across 6 categories (login, otp, password_reset, register, admin, token). Detection uses direct JOIN at query time — no silver schema changes needed. Eight detection rules: `auth_failure_burst`, `credential_stuffing`, `otp_brute_force`, `privilege_escalation_probe`, `password_reset_flood`, `registration_abuse`, `rate_limit_triggered`, `auth_success_after_burst`.

### Phase 4 — Behavioral & Data Exfiltration (done)

Four cross-field correlation rules with no schema changes. `ua_spoofing`: browser UA + high rate + no static assets (silver JOIN). `ip_rotation`: many distinct IPs from same /24 subnet per URI (silver aggregation). `bulk_data_extraction`: single client downloading > 10 MB (unique_client_count=1). `pagination_scraping`: many distinct query variants of same base URI per IP.

### Phase 5 — External Feeds & Advanced

External data sources and infrastructure-dependent detection.

| Item | Type | Effort | Dependency |
|------|------|:------:|------------|
| IP reputation feed service | Infra | Medium | Tor exit list (public), VPN/proxy lists |
| `tor_exit_node` | Rule | Small | IP reputation feed |
| `vpn_proxy_detected` | Rule | Small | IP reputation feed |
| `impossible_travel` | Rule | Large | Session/user tracking (new silver table) |
| `tls_fingerprint_mismatch` | Rule | Large | TLS JA3/JA4 data in logs |

### Alert Correlation (after detection is comprehensive)

Group anomalies into actionable alerts. Simple once detection is solid.

| Item | Type | Effort | Dependency |
|------|------|:------:|------------|
| Alert schema (`gold.httpmonitor_alerts`) | Infra | Small | Gold schema addition |
| `alert_id` + `client_ip` on anomaly schema | Infra | Small | Schema migration |
| `CorrelateAlerts` activity | Infra | Medium | Alert schema |
| Auto-resolve stale alerts in `CleanupStale` | Infra | Small | Alert schema |

## 📂 File Locations

| Component | Path |
|-----------|------|
| Detection activities | `pkg/detect/httpmonitor/activities.go` |
| Detection workflow | `pkg/detect/httpmonitor/workflows.go` |
| Registration | `pkg/detect/httpmonitor/register.go` |
| Detection rules loader | `pkg/detect/httpmonitor/rules.go` |
| Gold anomaly schema | `pkg/schema/gold/httpmonitor/anomaly.go` |
| Match rules service | `pkg/base/matchrule/matchrule.go` |
| Rule config schemas | `pkg/schema/config/rule/` |
| Seed data (all config) | `pkg/seed/config/` |
| Silver traffic schemas | `pkg/schema/silver/httptraffic/` |
| Bronze accesslog schemas | `pkg/schema/bronze/accesslog/` |
| Normalization logic | `pkg/normalize/httptraffic/` |
| GeoIP enrichment | `pkg/base/geoip/` |

## 📋 Data Retention

Retention period is configurable via `accesslog.retention_days` in app config (default: 30 days).

| Table | Retention | Managed By |
|-------|-----------|------------|
| `gold.httpmonitor_anomalies` | `retention_days` | `CleanupStale` activity |
| `gold.httpmonitor_alerts` | 90 days | `CleanupStale` activity |
| `silver.httptraffic_*` | `retention_days` | `CleanupStale` activity |
| `bronze.accesslog_*` | TBD | Not yet automated |
| `config.* (rules)` | Permanent | Manual / admin UI |
