# Software Classification Flow

How `debug-software` classifies installed apps (from S1 agents) into buckets for analysis.

Input: 377K app records → 8,702 unique names.

## Output Buckets

| Bucket | Description | Example |
|--------|-------------|---------|
| **Matched** | Mapped to a known product with lifecycle data | nginx, redis, postgresql |
| **OS Core** | Lifecycle tied to OS release, not tracked independently | libxml2, libreoffice-calc, pacemaker |
| **Unmatched** | Not classified — needs review or new data sources | sentinelagent, veeam, docker-compose-plugin |

## Classification Steps

### Step 1: Match app name → known product

Try each app name against product mappings (currently 223 from endoflife.date).

**Prefix match** (most products): name starts with slug + separator (`-` or space).
```
"nginx-common"    → slug "nginx"       (prefix "nginx" + "-")
"redis-server"    → slug "redis"       (prefix "redis" + "-")
"postgresql-16"   → slug "postgresql"  (prefix "postgresql" + "-")
```

**Exact match** (`exactOnly=true`, for broad names): name must equal slug exactly.
```
"libreoffice"      → slug "libreoffice"  ✓ (exact)
"libreoffice-calc" → slug "libreoffice"  ✗ (not exact, sub-package)
```

**Extra prefixes** (manual overrides for display name differences):
```
"node.js"  → slug "nodejs"  (Windows display name ≠ slug)
```

**Normalize fallback**: strip parenthesized suffixes and embedded version digits, then retry.
```
"mozilla firefox (x64 en-us)" → "mozilla firefox" → slug "firefox"
"postgresql17-ee-libs"        → "postgresql-ee-libs" → no match
```

**Excludes**: some products exclude sub-packages (e.g. redis excludes `-doc`).

Result: → **Matched** bucket.

### Step 2: Already matched? → skip

Names in `matchedNames` set are not checked further.

### Step 3: OS Core — exact match from package repos

Names found in Ubuntu/RPM package repos are OS core.

Sources (loaded from DB):
- `reference_ubuntu_packages` — 78K entries (noble+jammy, main+universe)
- `reference_rpm_packages` — 61K entries (10 repos: baseos, appstream, ha, crb, epel9, rhel7-os, updates, sclo, extras, epel7)

```
"libreoffice-calc"  → found in Ubuntu repos    → OS core
"libxml2"           → found in Ubuntu + RPM     → OS core
"pacemaker"         → found in RPM (rhel9-ha)   → OS core
"gnome-keyring"     → found in Ubuntu repos     → OS core
"sentinelagent"     → not in any repo           → continue
```

### Step 4: OS Core — exact name match (hardcoded)

Special cases like Google PWA shortcuts that appear as installed apps:
```
"gmail", "docs", "youtube", "slides", "sheets", "outlook (pwa)"
```

### Step 5: OS Core — version-stripped match

Strip version numbers from name, then check repos again.
```
"linux-headers-6.14.0-37-generic" → "linux-headers" → in repos → OS core
"libpython3.10-minimal"           → "libpython"     → in repos → OS core
```

### Step 6: OS Core — prefix patterns

Hardcoded prefixes for packages not in repos but clearly OS/infra:

| Category | Prefixes |
|----------|----------|
| Linux kernel/RHEL | `linux-`, `redhat-`, `gpg-pubkey`, `oem-` |
| Shared libraries | `lib`, `mesa-lib` |
| Google Cloud infra | `google-cloud-cli`, `google-cloud-sdk`, `google-cloud-ops-`, `google-cloud-sap-`, `google-rhui-client`, `gce-disk-expand`, `gcsfuse` |
| Windows VC++ | `microsoft visual c++` |
| Windows drivers | `intel(`, `intel®`, `oneapi `, `realtek `, `displaylink `, `thunderbolt` |
| Windows OEM | `lenovo `, `dell `, `hp ` |
| Windows OS | `windows pc health`, `windows 11 installation`, `windows 10 update`, `windows subsystem for`, `windows sdk`, `windows software dev`, `windows driver package`, `microsoft update health`, `microsoft search in bing`, `update for windows` |

### Step 7: OS Core — suffix patterns

Repo infrastructure packages not in repos:
```
"-keyring"        → "brave-keyring", "synaptics-repository-keyring"
"-repo"           → "pgdg-redhat-repo"
"-release-notes"  → OS release notes
"-release_notes"  → RHEL release notes (underscore variant)
```

Also matches mid-name: `-keyring-` catches `synaptics-repository-keyring-1.0`.

### Product Guard (applies to steps 6 + 7)

Before checking any prefix/suffix patterns, `guardedByEOL()` checks if the name could match a known product in step 1. If so, all pattern-based OS core filters are skipped — the name should stay visible for review, not silently hidden.

The guard only protects non-exactOnly products. For exactOnly products (e.g. `libreoffice`), only the exact slug name is protected — sub-packages like `libreoffice-calc` are not.

```
"libvirt-daemon" → split on "-" → "libvirt"
  → found in product slugs, not exactOnly
  → guard protects → skip all patterns → NOT os core

"libreoffice-calc" → split → "libreoffice"
  → found, BUT exactOnly=true
  → guard does NOT protect → patterns apply → OS core
  (correct: sub-packages of exactOnly products are OS core)

"brave-keyring" → split → "brave"
  → not in product slugs (Brave Browser not in endoflife.date)
  → guard does NOT protect → suffix "-keyring" applies → OS core
  (correct: just a repo signing key, not a browser)
```

### Unmatched

Everything that passed all steps without being classified:

```
sentinelagent          610 machines  (vendor agent)
insights-client        519 machines  (Red Hat agent)
nessusagent            394 machines  (vendor agent)
docker-compose-plugin   80 machines  (standalone versioning)
veeam                   41 machines  (backup software)
```

## Current Numbers

| Bucket | Linux | Windows |
|--------|------:|--------:|
| Total unique names | 8,702 | 845 |
| Matched | 324 | 114 |
| OS core | 8,202 | 169 |
| Unmatched | 176 | 562 |
