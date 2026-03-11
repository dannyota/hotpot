// Package listhandler provides config-driven HTTP handlers for paginated list
// endpoints. Two flavours are supported: ent-backed (Handler) and raw-SQL-backed
// (RegisterSQL).
package listhandler

import (
	"database/sql"

	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
)

// ---------------------------------------------------------------------------
// Shared types
// ---------------------------------------------------------------------------

// Predicate is a type alias for ent / raw-SQL selector functions.
type Predicate = func(*entsql.Selector)

// SortFunc matches ent ByXxx order constructors.
type SortFunc = func(opts ...entsql.OrderTermOption) func(*entsql.Selector)

// FilterKind describes how a filter field is matched.
type FilterKind int

const (
	Search FilterKind = iota // ContainsFold / ILIKE — substring search
	Exact                    // EQ / = — exact match
	Multi                    // IN with (empty) sentinel support
)

// emptyFilterValue is the sentinel used by the frontend for NULL / empty values.
const emptyFilterValue = "(empty)"

// ---------------------------------------------------------------------------
// Filter options (shared by both ent and SQL handlers)
// ---------------------------------------------------------------------------

// FilterOptionsConfig describes how to query DISTINCT filter-option values via
// raw SQL. Both the ent handler and the SQL handler use it.
type FilterOptionsConfig struct {
	DB          *sql.DB
	Schema      string            // PG schema  (e.g. "bronze")
	Table       string            // PG table   (e.g. "greennode_compute_servers")
	Columns     []string          // columns to query DISTINCT values for
	ColumnExprs map[string]string // optional SQL expression overrides (e.g. extract short name from URL path)
	From        string            // optional: override FROM clause (e.g. subquery with JOIN)
}

// NavMeta re-exports admin.NavMeta for convenience.
type NavMeta = admin.NavMeta
