# Admin Table Pages Guide

How to add a dedicated table page to the admin UI.

Two levels of table pages exist:

| Level | When to use | Frontend | Backend |
|-------|------------|----------|---------|
| **Generic** | Simple read-only tables, no custom rendering | `GenericTablePage.vue` (automatic) | `lh.RegisterSQL()` |
| **Dedicated** | Custom columns, stats, search, export, drawers | New `XxxPage.vue` | `lh.Handler()` with `lh.Config` |

Generic pages require zero frontend code — just register the SQL table in Go and it appears in the nav.

This guide covers both levels: **generic pages first** (fastest path), then **dedicated pages**.

## Reference Implementations

| Level | Backend | Frontend |
|-------|---------|----------|
| Generic | `pkg/admin/bronze/meec/register.go` | `GenericTablePage.vue` (automatic) |
| Dedicated | `pkg/admin/bronze/greennode/compute/register.go` | `ServersPage.vue` |

## Adding a New Provider (Generic / RegisterSQL)

Fastest path — zero frontend code. Creates browsable table pages with pagination, sorting, filters, and CSV export automatically.

### 1. Create register.go

Location: `pkg/admin/bronze/{provider}/register.go`

```go
package vault

import (
    "database/sql"

    "danny.vn/hotpot/pkg/admin"
    lh "danny.vn/hotpot/pkg/admin/listhandler"
)

func Register(db *sql.DB) {
    lh.RegisterSQL(db, sqlTables)
}

var sqlTables = []lh.SQLTable{
    {
        API:    "/api/v1/bronze/vault/pki/certificates",
        Schema: "bronze",
        Table:  "vault_pki_certificates",
        Nav:    admin.NavMeta{Label: "Certificates", Group: []string{"Bronze", "Vault", "PKI"}},
        Columns: []string{
            "resource_id", "vault_name", "mount_path", "common_name",
            "serial_number", "key_type", "not_before", "not_after",
            "is_revoked", "collected_at", "first_collected_at",
        },
        Filters: []lh.SQLFilterDef{
            {Column: "common_name", Kind: lh.Search},
            {Column: "vault_name", Kind: lh.Multi},
            {Column: "mount_path", Kind: lh.Multi},
            {Column: "key_type", Kind: lh.Multi},
            {Column: "is_revoked", Kind: lh.Multi},
        },
        DefaultSort:         "collected_at",
        DefaultDesc:         true,
        FilterOptionColumns: []string{"vault_name", "mount_path", "key_type", "is_revoked"},
    },
}
```

### 2. Wire into parent

Add the call in `pkg/admin/bronze/register.go`:

```go
import "danny.vn/hotpot/pkg/admin/bronze/vault"

func Register(driver dialect.Driver, db *sql.DB) {
    gcp.Register(driver, db)
    greennode.Register(driver, db)
    s1.Register(driver, db)
    meec.Register(db)
    vault.Register(db)       // ← new
}
```

### 3. Done

No frontend code needed. `GenericTablePage.vue` auto-renders from the nav registration. The page appears in the sidebar navigation with search, filters, pagination, and CSV export.

### SQLTable Fields

| Field | Required | Purpose |
|-------|:--------:|---------|
| `API` | ✅ | API path (`/api/v1/bronze/...`) |
| `Schema` | ✅ | PG schema (`bronze`, `silver`, `gold`) |
| `Table` | ✅ | PG table name |
| `Nav` | ✅ | Sidebar label and group hierarchy |
| `Columns` | | Column whitelist (omit = all columns) |
| `Filters` | | Search and multi-select filter definitions |
| `DefaultSort` | | Sort column (default: first column) |
| `DefaultDesc` | | Sort descending (default: false) |
| `FilterOptionColumns` | | Columns for dropdown option counts |

### SQLFilterDef Kinds

| Kind | Usage |
|------|-------|
| `lh.Search` | Substring search (ILIKE) on column |
| `lh.Multi` | Multi-select dropdown (IN) on column |

### Helper Pattern

For multiple tables under one provider, use a helper:

```go
func bronzeMEEC(api, table, label string) lh.SQLTable {
    return lh.SQLTable{
        API:    "/api/v1/bronze/meec/" + api,
        Schema: "bronze",
        Table:  table,
        Nav:    admin.NavMeta{Label: label, Group: []string{"Bronze", "MEEC"}},
    }
}
```

### Checklist (Generic)

- [ ] Backend: `register.go` with `lh.RegisterSQL(db, tables)`
- [ ] Backend: wire into `pkg/admin/bronze/register.go`
- [ ] Build: `go build ./pkg/admin/...`

## Provider Specs

### Vault PKI Certificates

Table: `bronze.vault_pki_certificates`

| Column | Type | Filter | Purpose |
|--------|------|:------:|---------|
| `resource_id` | string | | `{vault_name}/{mount_path}/{serial}` |
| `vault_name` | string | Multi | Vault instance name |
| `mount_path` | string | Multi | PKI mount path |
| `common_name` | string | Search | Certificate CN |
| `serial_number` | string | | Certificate serial |
| `key_type` | string | Multi | RSA, ECDSA, etc. |
| `key_bits` | int | | Key size |
| `not_before` | timestamp | | Valid from |
| `not_after` | timestamp | | Expiry date |
| `is_revoked` | bool | Multi | Revocation status |
| `collected_at` | timestamp | | Last seen |
| `first_collected_at` | timestamp | | First seen |

Nav: Bronze > Vault > PKI > Certificates

### API Catalog Endpoints

Table: `bronze.apicatalog_endpoints_raw`

| Column | Type | Filter | Purpose |
|--------|------|:------:|---------|
| `resource_id` | string | | UUID |
| `name` | string | Search | Route name |
| `service_name` | string | | Service label |
| `upstream` | string | Multi | Upstream code (e.g. "dbs") |
| `uri` | string | Search | API path |
| `method` | string | Multi | HTTP method(s) |
| `route_status` | string | Multi | Active/Inactive |
| `plugin_auth` | string | | Auth plugin name |
| `source_file` | string | | Import CSV filename |
| `collected_at` | timestamp | | Last seen |
| `first_collected_at` | timestamp | | First seen |

Nav: Bronze > API Catalog > Endpoints

## Dedicated Pages

For custom columns, stats, search, export, drawers — create a dedicated `XxxPage.vue` with `lh.Handler()`.

Reference: `ServersPage.vue` + `pkg/admin/bronze/greennode/compute/register.go`.

### 1. Create register.go

Location: `pkg/admin/{layer}/{provider}/{service}/register.go`

```go
package myservice

import (
    "database/sql"
    "entgo.io/ent/dialect"
    entsql "entgo.io/ent/dialect/sql"

    "danny.vn/hotpot/pkg/admin"
    lh "danny.vn/hotpot/pkg/admin/listhandler"
    entmyservice "danny.vn/hotpot/pkg/storage/ent/myservice"
    p "danny.vn/hotpot/pkg/storage/ent/myservice/bronzemyentity"
    "danny.vn/hotpot/pkg/storage/ent/myservice/predicate"
)

func Register(driver dialect.Driver, db *sql.DB) {
    entClient := entmyservice.NewClient(
        entmyservice.Driver(driver),
        entmyservice.AlternateSchema(entmyservice.DefaultSchemaConfig()),
    )

    admin.RegisterRoute(admin.RouteRegistration{
        Method: "GET",
        Path:   "/api/v1/bronze/provider/service/entities",
        Nav:    &admin.NavMeta{Label: "Entities", Group: []string{"Bronze", "Provider", "Service"}},
        Handler: lh.Handler(lh.Config{
            EntityName:    "entities",
            AllowedFields: map[string]bool{ /* ... */ },
            NewQuery:      func() lh.QueryAdapter { /* ... */ },
            Filters:       []lh.FilterDef{ /* ... */ },
            SortFields:    map[string]lh.SortFunc{ /* ... */ },
            DefaultOrder:  p.ByCollectedAt(entsql.OrderDesc()),
            FilterOptions: &lh.FilterOptionsConfig{ /* ... */ },
        }),
    })
}
```

### 2. Config fields

#### AllowedFields

Whitelist of fields accepted in `sort` and `filter[*]` query params. Include every sortable column and every filterable field.

```go
AllowedFields: map[string]bool{
    "name": true, "status": true, "collected_at": true,
},
```

#### NewQuery — QueryAdapter

Type-erased wrapper around an ent query builder. Four methods:

```go
NewQuery: func() lh.QueryAdapter {
    q := entClient.MyEntity.Query()
    return lh.QueryAdapter{
        Where:      func(ps ...lh.Predicate) {
            q.Where(lh.ConvertSlice[predicate.MyEntity](ps)...)
        },
        CloneCount: func(ctx context.Context) (int, error) {
            return q.Clone().Count(ctx)
        },
        Order:      func(os ...lh.Predicate) {
            q.Order(lh.ConvertSlice[p.OrderOption](os)...)
        },
        Fetch:      func(ctx context.Context, off, lim int) (any, error) {
            return q.Offset(off).Limit(lim).All(ctx)
        },
    }
},
```

#### Filters

Three filter kinds:

| Kind | Usage | Required fields |
|------|-------|----------------|
| `Search` | Substring search (ILIKE) | `Pred` |
| `Exact` | Exact match (EQ) | `Pred` |
| `Multi` | Multi-select dropdown (IN + empty sentinel) | `InFn`, `EqFn` |

```go
Filters: []lh.FilterDef{
    // Search: text input, substring match
    {Field: "name", Kind: lh.Search, Pred: lh.Pred(p.NameContainsFold)},

    // Exact: programmatic exact match
    {Field: "project_id", Kind: lh.Exact, Pred: lh.Pred(p.ProjectIDEQ)},

    // Multi: dropdown with checkboxes, supports "(empty)" sentinel
    {Field: "status", Kind: lh.Multi,
        InFn: lh.PredIn(p.StatusIn),
        EqFn: lh.Pred(p.StatusEQ)},
},
```

**Custom search predicates** (OR across multiple fields):

```go
{Field: "q", Kind: lh.Search, Pred: nameOrIPSearchPred},

func nameOrIPSearchPred(v string) lh.Predicate {
    return func(s *entsql.Selector) {
        s.Where(entsql.Or(
            entsql.P(func(b *entsql.Builder) {
                b.WriteString("LOWER(name) LIKE LOWER(")
                b.Arg("%" + v + "%")
                b.WriteByte(')')
            }),
            entsql.P(func(b *entsql.Builder) {
                b.WriteString("CAST(interfaces_json AS TEXT) ILIKE ")
                b.Arg("%" + v + "%")
            }),
        ))
    }
}
```

#### SortFields

Map field names to ent order constructors:

```go
SortFields: map[string]lh.SortFunc{
    "name":         lh.Sort(p.ByName),
    "status":       lh.Sort(p.ByStatus),
    "collected_at": lh.Sort(p.ByCollectedAt),
},
```

#### FilterOptions

Drives the multi-select dropdown counts. Queries `SELECT col, COUNT(*) GROUP BY col` for each column:

```go
FilterOptions: &lh.FilterOptionsConfig{
    DB:      db,
    Schema:  "bronze",
    Table:   "my_table",
    Columns: []string{"status", "region"},
},
```

### 3. Stats endpoint (optional)

Page-specific stats live in the same `register.go`. Each page owns its stats — no shared stats handler.

```go
admin.RegisterRoute(admin.RouteRegistration{
    Method:  "GET",
    Path:    "/api/v1/bronze/provider/service/entities/stats",
    Handler: myStatsHandler(db),
})
```

### 4. Wire into parent

Add the call in the parent `register.go`:

```go
// pkg/admin/bronze/provider/register.go
func Register(driver dialect.Driver, db *sql.DB) {
    myservice.Register(driver, db)
}
```

## Frontend

### 1. Create page component

Location: `admin/ui/src/pages/{layer}/{provider}/{service}/XxxPage.vue`

### 2. Script setup structure

```vue
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useTableQuery } from '@/composables/useTableQuery'
import { exportCsv } from '@/composables/useExport'
import type { Column } from '@/components/app/HDataTable.vue'
import HDataTable from '@/components/app/HDataTable.vue'
import HMultiSelect from '@/components/app/HMultiSelect.vue'
import HDetailDrawer from '@/components/app/HDetailDrawer.vue'
import HJsonViewer from '@/components/app/HJsonViewer.vue'
import HJsonPeek from '@/components/app/HJsonPeek.vue'
import { Search, X, Download } from 'lucide-vue-next'
import { columnLabel } from '@/composables/columns'
import { useTimezone } from '@/composables/useTimezone'
import { TooltipRoot, TooltipTrigger, TooltipPortal, TooltipContent } from 'reka-ui'

const { formatDateTime, relativeTime } = useTimezone()
const route = useRoute()
const ENDPOINT = '/api/v1/layer/provider/service/entities'

const {
  data, meta, filterOptions, loading, error, sort, filters,
  reload, setSort, setFilter, clearFilters, setPage, setPageSize,
} = useTableQuery<Record<string, any>>({
  endpoint: ENDPOINT,
  defaultSort: '-collected_at',
})
```

### 3. Columns

```ts
const columns = computed<Column<Record<string, any>>[]>(() => [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'collected_at', label: columnLabel('collected_at'), sortable: true },
])
```

Use `columnLabel()` for standard fields (`collected_at` → "Last Seen", `first_collected_at` → "First Seen").

### 4. Multi-select filters

Single record pattern — avoids separate refs per filter:

```ts
const multiFilters = ref<Record<string, string[]>>({})

function onMultiFilterChange(field: string, values: string[]) {
  multiFilters.value = { ...multiFilters.value, [field]: values }
  setFilter(field, values.join(','))
}
```

Template — loop over filter field names:

```vue
<HMultiSelect
  v-for="field in ['status', 'region']"
  :key="field"
  :label="field.charAt(0).toUpperCase() + field.slice(1)"
  :options="filterOptions[field] ?? []"
  :model-value="multiFilters[field] ?? []"
  @update:model-value="onMultiFilterChange(field, $event)"
/>
```

### 5. Debounced search

```ts
const searchText = ref('')
let searchTimeout: ReturnType<typeof setTimeout> | null = null

function onSearchInput() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    setFilter('q', searchText.value)  // or 'name' for single-field search
  }, 300)
}
```

### 6. Detail drawer (optional)

```ts
const drawerRow = ref<Record<string, any> | null>(null)

function drawerFields(row: Record<string, any>) {
  return [
    { label: 'Name', value: row.name },
    { label: 'Status', value: row.status },
    { label: 'Created', value: row.created_at ? formatDateTime(row.created_at) : null },
  ]
}
```

```vue
<HDataTable @row-click="(row) => drawerRow = row" ... />

<HDetailDrawer
  v-if="drawerRow"
  :title="drawerRow.name"
  :fields="drawerFields(drawerRow)"
  @close="drawerRow = null"
/>
```

### 7. CSV export (optional)

```ts
const exporting = ref(false)

async function onExport() {
  exporting.value = true
  try {
    await exportCsv(ENDPOINT, filters.value, sort.value, meta.value.total, 'entities.csv')
  } finally {
    exporting.value = false
  }
}
```

### 8. Stats strip (optional)

Fetch from a dedicated `/stats` endpoint. Format inline with the export button:

```ts
interface StatGroup { count: number; breakdown: Record<string, number> }
const stats = ref<Record<string, StatGroup>>({})

async function fetchStats() {
  try {
    const res = await window.fetch(`${ENDPOINT}/stats`)
    if (res.ok) stats.value = await res.json()
  } catch { /* ignore */ }
}

onMounted(() => { reload(); fetchStats() })
```

### 9. Template layout

```
┌─ Stats strip ──────────────────────────────── [Export CSV] ─┐
├─ [Search input] [Filter] [Filter] [Filter]     [Clear]     ─┤
├─ HDataTable (with column slots)                             ─┤
├─ HDetailDrawer (slide-in from right)                        ─┤
└─ HJsonViewer (modal)                                        ─┘
```

Standard spacing: `p-6 space-y-2 max-w-full`

### 10. Date columns

Three patterns:

```vue
<!-- Absolute date -->
<template #created_at="{ value }">
  <span class="text-zinc-500 dark:text-zinc-400 text-xs tabular-nums">
    {{ formatDateTime(value) }}
  </span>
</template>

<!-- Relative time with tooltip -->
<template #collected_at="{ value }">
  <TooltipRoot :delay-duration="200">
    <TooltipTrigger as-child>
      <span class="text-zinc-500 dark:text-zinc-400 text-xs tabular-nums cursor-default">
        {{ relativeTime(value) }}
      </span>
    </TooltipTrigger>
    <TooltipPortal>
      <TooltipContent side="top" :side-offset="4"
        class="z-[100] rounded-md border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 shadow-lg px-2.5 py-1.5 text-xs text-zinc-700 dark:text-zinc-300 animate-in fade-in-0 zoom-in-95">
        {{ formatDateTime(value) }}
      </TooltipContent>
    </TooltipPortal>
  </TooltipRoot>
</template>
```

### 11. Status badges

```ts
function statusColor(status: string): string {
  switch (status?.toUpperCase()) {
    case 'ACTIVE': return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'ERROR': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    default: return 'bg-zinc-100 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400'
  }
}
```

### 12. Add route

In `admin/ui/src/router/index.ts`:

```ts
{
  path: '/layer/provider/service/entities',
  name: 'provider-service-entities',
  component: () => import('@/pages/layer/provider/service/EntitiesPage.vue'),
  meta: { breadcrumb: ['Layer', 'Provider', 'Service', 'Entities'] },
},
```

## Reusable Components

| Component | Props | Purpose |
|-----------|-------|---------|
| `HDataTable` | columns, data, meta, sort, loading, error | Paginated sortable table |
| `HMultiSelect` | label, options (`FilterOption[]`), modelValue | Multi-select dropdown with counts |
| `HDetailDrawer` | title, fields | Slide-in row detail panel |
| `HJsonPeek` | label, data, title, extra | Inline JSON preview with tooltip |
| `HJsonViewer` | data, title | Full-screen JSON modal |
| `HAlert` | type, message, dismissible | Inline error/warning/info banner |

## Composables

| Composable | Purpose |
|------------|---------|
| `useTableQuery(opts)` | Wires pagination + sort + filter state to `useListApi` |
| `useListApi(endpoint)` | Fetches paginated list, returns data/meta/filterOptions |
| `useApi(endpoint)` | Fetches single JSON object |
| `buildUrl(base, params)` | Builds URL with query params (shared by useListApi + useExport) |
| `exportCsv(...)` | Fetches all filtered rows and downloads as CSV |
| `useTimezone()` | `formatDateTime()` and `relativeTime()` with timezone preference |
| `columnLabel(key)` | Maps field names to display labels (`collected_at` → "Last Seen") |
| `useNotifications()` | Persistent notification store (auto-captures API errors) |

## API Contract

All list endpoints return the same envelope:

```json
{
  "data": [...],
  "meta": { "page": 1, "size": 20, "total": 100, "total_pages": 5 },
  "filter_options": {
    "status": [
      { "value": "ACTIVE", "count": 95 },
      { "value": "STOPPED", "count": 5 }
    ]
  }
}
```

Query params: `?page=1&size=20&sort=-name&filter[status]=ACTIVE,STOPPED&filter[q]=search`

- Sort: field name (ascending) or `-field` (descending)
- Filters: `filter[field]=value` for single, `filter[field]=a,b,c` for multi
- Size: 1–10000 (default 20, max 10000 for CSV export)

## Checklist

- [ ] Backend: `register.go` with `lh.Config`
- [ ] Backend: wire into parent `register.go`
- [ ] Backend: stats endpoint (if needed)
- [ ] Frontend: page component with `useTableQuery`
- [ ] Frontend: columns definition
- [ ] Frontend: search input + multi-select filters
- [ ] Frontend: column slot templates (status badges, dates, JSON)
- [ ] Frontend: detail drawer (if needed)
- [ ] Frontend: CSV export (if needed)
- [ ] Frontend: route in `router/index.ts`
- [ ] Build: `go build ./cmd/admin/...` + `npx vue-tsc --noEmit`
