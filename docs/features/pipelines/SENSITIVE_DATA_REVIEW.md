# Sensitive Data Review

Detect and mask sensitive data exposure across the bronze → silver → gold pipeline.

## 🎯 Approach

Config-driven. Users define which field names are sensitive. We ship predefined defaults — users extend for their environment. Match on **field names**, never extract or log actual values.

```
  config.sensitive_fields          Pipeline layers
  ═══════════════════════          ════════════════

  ┌──────────────────────┐
  │  Predefined defaults  │
  │  + user-defined rules │
  └──────────┬───────────┘
             │
     ┌───────┼────────────────────────────────────────┐
     │       │                                         │
     │       ▼            BRONZE (ingest)              │
     │  Raw logs ──► Parse URI + JSON body             │
     │               Match field names                 │
     │               Mask values (*** )                │
     │               Store masked data                 │
     │                       │                         │
     ├───────────────────────┼─────────────────────────┤
     │                       │                         │
     │                       ▼    SILVER (normalize)   │
     │  Tag endpoints:                                 │
     │    sensitive_fields: ["email","pan"]             │
     │    has_sensitive_data: true                      │
     │                       │                         │
     ├───────────────────────┼─────────────────────────┤
     │                       │                         │
     │                       ▼    GOLD (detect)        │
     │  Detection rules:                               │
     │    sensitive_field_in_uri                        │
     │    sensitive_field_in_body                       │
     │               │                                 │
     │               ▼                                 │
     │    Anomalies ──► Alerts                         │
     │    "POST /api/payment exposes card_number"      │
     │                                                 │
     └─────────────────────────────────────────────────┘

  Standalone: Preprod Audit
  ═════════════════════════
  One-time scan of existing data ──► Report
```

## ⚙️ Config — `config.sensitive_fields`

Stored in `config` schema, cached 5 min, hot-reloadable via match rule service. Used by all layers.

| Field | Type | Purpose |
|-------|------|---------|
| `name` | string | Field name to match |
| `match_type` | enum | `exact`, `prefix`, `contains` |
| `category` | enum | `pii`, `credential`, `financial` |
| `severity` | enum | `critical`, `high`, `medium` |
| `source` | enum | `predefined`, `user` |
| `description` | string | What this field contains |

### Predefined Defaults

Ship with the system. Users cannot delete but can disable.

**PII** (severity: high):

| Name | Match | Description |
|------|:-----:|-------------|
| `email` | exact | Email address |
| `mail` | exact | Email address (alt) |
| `phone` | exact | Phone number |
| `mobile` | exact | Mobile number |
| `ssn` | exact | Social security number |
| `id_number` | exact | National ID |
| `citizen_id` | exact | Citizen ID |
| `passport` | exact | Passport number |
| `dob` | exact | Date of birth |
| `date_of_birth` | exact | Date of birth (alt) |

**Credential** (severity: critical):

| Name | Match | Description |
|------|:-----:|-------------|
| `password` | exact | Password |
| `passwd` | exact | Password (alt) |
| `secret` | exact | Secret value |
| `token` | contains | Auth / API token |
| `api_key` | exact | API key |
| `apikey` | exact | API key (alt) |
| `access_key` | exact | Access key |
| `private_key` | exact | Private key |
| `authorization` | exact | Auth header value |
| `otp` | exact | One-time password |
| `pin` | exact | PIN code |

**Financial** (severity: high):

| Name | Match | Description |
|------|:-----:|-------------|
| `pan` | exact | Card number (PAN) |
| `card_number` | exact | Card number |
| `card` | prefix | Card-related fields |
| `account_number` | exact | Bank account |
| `routing_number` | exact | Routing number |
| `cvv` | exact | Card CVV |
| `cvc` | exact | Card CVC |

### User-Defined Rules

Users add rules via admin UI or config API. Examples:

| Name | Match | Category | Why |
|------|:-----:|----------|-----|
| `nric` | exact | pii | Singapore national ID |
| `nik` | exact | pii | Indonesian citizen ID |
| `cpf` | exact | pii | Brazilian tax ID |
| `partner_secret` | exact | credential | Internal partner auth |
| `ref_code` | exact | financial | Internal reference code |

## 🥉 Bronze — Ingest & Mask

Sensitive data masking at ingestion, before raw data persists in bronze tables.

### What Gets Scanned

| Location | Format | Example |
|----------|--------|---------|
| URI query params | `?email=xxx&token=xxx` | GET requests leaking data in URL |
| JSON request body | `{"password": "xxx"}` | POST/PUT payloads |
| JSON response body | `{"card_number": "xxx"}` | API responses containing sensitive fields |

### Masking

Parse URI query params and JSON bodies during ingestion. Match field names against `sensitive_fields` config. Replace values with `***`. Preserve field names for downstream analytics.

```
  Raw log                               After masking
  ═══════                               ════════════

  URI:
  /api/user?email=john@ex.com       →  /api/user?email=***
  /api/login?password=secret        →  /api/login?password=***

  JSON body:
  {"email": "john@ex.com",          →  {"email": "***",
   "card_number": "4111..."}             "card_number": "***"}
```

JSON scanning uses **leaf key** matching — `customer.email` matches rule `email`. Works regardless of nesting:

```
  {                                  Matched keys
    "customer": {                    ════════════
      "email": "john@ex.com",       → email         ← pii
      "phone": "0912345678"         → phone         ← pii
    },
    "payment": {
      "card_number": "4111...",      → card_number   ← financial
      "cvv": "123"                   → cvv           ← financial
    }
  }
```

### Bronze Changes

| Change | Where | Purpose |
|--------|-------|---------|
| URI masking | `pkg/ingest/accesslog/` | Redact query param values before storing |
| JSON body masking | `pkg/ingest/accesslog/` | Redact JSON field values before storing |
| `is_masked` flag | `accesslog_http_counts` | Track which records were masked |

## 🥈 Silver — Normalize & Tag

Silver normalization tags endpoints with sensitive field metadata. Downstream layers (gold detection, dashboards) use these tags without re-scanning.

### New Fields on `httptraffic_traffic_5m`

| Field | Type | Purpose |
|-------|------|---------|
| `has_sensitive_data` | bool | Endpoint URI or body contains sensitive field names |
| `sensitive_fields` | string[] | List of matched field names (e.g., `["email", "pan"]`) |
| `sensitive_categories` | string[] | Distinct categories (e.g., `["pii", "financial"]`) |
| `max_sensitive_severity` | enum | Highest severity among matched fields |

### How It Works

During normalization, re-check field names from bronze URI + body data against config. Tag the silver record. This runs every 5-min cycle as part of existing normalization.

```
  Bronze record                       Silver record
  ═════════════                       ═════════════
  uri: /api/user?email=***            has_sensitive_data: true
  body_fields: ["email","name"]   →   sensitive_fields: ["email"]
                                      sensitive_categories: ["pii"]
                                      max_sensitive_severity: high
```

## 🥇 Gold — Detect & Alert

Detection rules in the httpmonitor pipeline. Alert when endpoints expose sensitive data.

### Detection Rules

| Type | Severity | Trigger |
|------|:--------:|---------|
| `sensitive_field_in_uri` | From config | Query param name matches `sensitive_fields` |
| `sensitive_field_in_body` | From config | JSON body key matches `sensitive_fields` |

One anomaly per (endpoint, field, location, window). Severity from the matched config entry.

Runs in `DetectSuspiciousPatterns` activity, reads `has_sensitive_data` / `sensitive_fields` from silver — no re-parsing needed.

### Alert Behavior

Grouped by endpoint scope. One alert per endpoint exposing sensitive data. Alert lists all sensitive fields found and their categories. Stays open as long as the endpoint keeps receiving sensitive fields.

```
  Alert: "POST /api/payment exposes sensitive data"
  ════════════════════════════════════════════════
  scope: endpoint
  severity: critical (credential fields found)
  sensitive_fields: [card_number, cvv, email]
  categories: [financial, pii]
  status: open (ongoing exposure)
```

## 📋 Preprod Audit (standalone)

One-time scan of existing preprod data. Uses the same `sensitive_fields` config. Produces a report, not alerts.

### How It Works

1. Load `sensitive_fields` (predefined + user-defined)
2. **URI scan**: query bronze `accesslog_http_counts` — extract query param names, match against config
3. **JSON scan**: sample request/response bodies from bronze — parse JSON, walk keys, match against config
4. Group by endpoint + field name + location
5. Output report

### Report Output

| Endpoint | Method | Location | Sensitive Field | Category | Severity |
|----------|--------|----------|-----------------|----------|:--------:|
| `/api/user` | GET | uri | `email` | pii | high |
| `/api/login` | POST | request_body | `password` | credential | critical |
| `/api/customer` | GET | response_body | `phone` | pii | high |
| `/api/payment` | POST | request_body | `card_number` | financial | high |
| `/api/payment` | POST | request_body | `cvv` | financial | high |
| `/api/verify` | POST | request_body | `otp` | credential | critical |

### Deliverables

| Output | Format | Purpose |
|--------|--------|---------|
| Findings report | CSV/JSON | Endpoints with sensitive fields, by location |
| Remediation list | Table | Backend fixes (encrypt, hash, remove from response) |
| Risk summary | Table | Count by category, severity, and location |

## 📊 Summary

| Layer | What | Effort | Dependency |
|-------|------|:------:|------------|
| Config | `sensitive_fields` table + predefined defaults | Small | New config schema |
| Bronze | Mask URI + JSON body values at ingestion | Medium | Config table |
| Silver | Tag endpoints with `has_sensitive_data` + field list | Small | Bronze masking |
| Gold | `sensitive_field_in_uri` + `sensitive_field_in_body` rules | Small | Silver tags |
| Standalone | Preprod audit scan → report | Medium | Config table |

## 📂 Related Files

| Component | Path |
|-----------|------|
| HTTP Monitor doc | [HTTPMONITOR.md](./HTTPMONITOR.md) |
| Bronze accesslog schemas | `pkg/schema/bronze/accesslog/` |
| Normalization logic | `pkg/normalize/httptraffic/` |
| Match rules service | `pkg/base/matchrule/matchrule.go` |
| Rule config schemas | `pkg/schema/config/rule/` |
