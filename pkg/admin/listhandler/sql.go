package listhandler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"danny.vn/hotpot/pkg/admin"
	"danny.vn/hotpot/pkg/admin/query"
)

// validColumn matches safe PG column identifiers.
var validColumn = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ---------------------------------------------------------------------------
// SQL handler types
// ---------------------------------------------------------------------------

// SQLFilterDef declares a filterable SQL column and how to match it.
type SQLFilterDef struct {
	Column string     // PG column name
	Kind   FilterKind // Search, Exact, or Multi
	Suffix bool       // use LIKE '%/<value>' instead of = (for URL path columns)
}

// SQLTable defines a generic browsable table backed by raw SQL.
type SQLTable struct {
	API    string        // route path  (e.g. "/api/v1/bronze/gcp/compute/addresses")
	Schema string        // PG schema   (e.g. "bronze")
	Table  string        // PG table    (e.g. "gcp_compute_addresses")
	Nav    admin.NavMeta // sidebar metadata

	// Columns whitelists column names for sort / filter.
	// When nil the handler runs in backwards-compatible mode (no sort/filter).
	Columns []string

	Filters     []SQLFilterDef // filterable columns
	DefaultSort string         // default ORDER BY column (empty = none)
	DefaultDesc bool           // default sort direction

	// FilterOptionColumns lists columns for the dropdown count queries.
	FilterOptionColumns []string

	// ColumnExprs overrides the SQL expression for specific filter-option
	// columns (e.g. extract short name from a GCP URL path).
	ColumnExprs map[string]string

	// From overrides the FROM clause with a custom SQL SELECT (used as a
	// subquery). Enables JOINs — e.g. enriching software rows with machine
	// hostname. When set, data/count queries use (From) instead of "schema"."table".
	From string
}

// ---------------------------------------------------------------------------
// Validation (runs at startup — panics on programming errors)
// ---------------------------------------------------------------------------

func validateSQLTable(t SQLTable) {
	colSet := make(map[string]bool, len(t.Columns))
	for _, c := range t.Columns {
		if !validColumn.MatchString(c) {
			panic(fmt.Sprintf("listhandler: invalid column name %q in Columns for table %s.%s", c, t.Schema, t.Table))
		}
		colSet[c] = true
	}
	for _, f := range t.Filters {
		if !validColumn.MatchString(f.Column) {
			panic(fmt.Sprintf("listhandler: invalid column name %q in Filters for table %s.%s", f.Column, t.Schema, t.Table))
		}
		if len(t.Columns) > 0 && !colSet[f.Column] {
			panic(fmt.Sprintf("listhandler: filter column %q not in Columns for table %s.%s", f.Column, t.Schema, t.Table))
		}
	}
	if t.DefaultSort != "" {
		if !validColumn.MatchString(t.DefaultSort) {
			panic(fmt.Sprintf("listhandler: invalid DefaultSort %q for table %s.%s", t.DefaultSort, t.Schema, t.Table))
		}
		if len(t.Columns) > 0 && !colSet[t.DefaultSort] {
			panic(fmt.Sprintf("listhandler: DefaultSort %q not in Columns for table %s.%s", t.DefaultSort, t.Schema, t.Table))
		}
	}
	for _, c := range t.FilterOptionColumns {
		if !validColumn.MatchString(c) {
			panic(fmt.Sprintf("listhandler: invalid column name %q in FilterOptionColumns for table %s.%s", c, t.Schema, t.Table))
		}
		if len(t.Columns) > 0 && !colSet[c] {
			panic(fmt.Sprintf("listhandler: FilterOptionColumns column %q not in Columns for table %s.%s", c, t.Schema, t.Table))
		}
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

// RegisterSQL registers a generic list handler for each SQL table.
func RegisterSQL(db *sql.DB, tables []SQLTable) {
	for _, t := range tables {
		validateSQLTable(t)
		nav := t.Nav
		admin.RegisterRoute(admin.RouteRegistration{
			Method:  "GET",
			Path:    t.API,
			Handler: sqlListHandler(db, t),
			Nav:     &nav,
		})
	}
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

func sqlListHandler(db *sql.DB, t SQLTable) http.HandlerFunc {
	allowedFields := buildAllowedFields(t.Columns)
	filterMap := buildFilterMap(t.Filters)

	// When From is set, wrap the custom SELECT as a subquery.
	// dataFrom includes the alias "t" for row_to_json; countFrom does not.
	var dataFrom, countFrom string
	if t.From != "" {
		dataFrom = fmt.Sprintf(`(%s) t`, t.From)
		countFrom = dataFrom
	} else {
		countFrom = fmt.Sprintf(`"%s"."%s"`, t.Schema, t.Table)
		dataFrom = countFrom + " t"
	}

	var filterOptsCfg *FilterOptionsConfig
	if len(t.FilterOptionColumns) > 0 {
		filterOptsCfg = &FilterOptionsConfig{
			DB: db, Schema: t.Schema, Table: t.Table,
			Columns: t.FilterOptionColumns, ColumnExprs: t.ColumnExprs,
			From: t.From,
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		params := parseParams(r, allowedFields, w)
		if params == nil {
			return // error already written
		}

		// Build WHERE, ORDER BY.
		wb := sqlBuildWhere(params.Filters, filterMap)
		orderSQL := sqlBuildOrder(params.Sort, t.DefaultSort, t.DefaultDesc, allowedFields)

		// Count.
		var total int
		if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s%s`, countFrom, wb.SQL), wb.Args...).Scan(&total); err != nil {
			admin.WriteServerError(w, "failed to load data", err)
			return
		}

		// CSV export adjusts size/offset.
		isCSV := r.URL.Query().Get("format") == "csv"
		size, offset := params.Size, params.Offset()
		if isCSV {
			size = min(total, 10000)
			offset = 0
		}

		// Data query.
		dataQuery := fmt.Sprintf(`SELECT row_to_json(t) FROM %s%s%s LIMIT $%d OFFSET $%d`,
			dataFrom, wb.SQL, orderSQL, wb.NextArg, wb.NextArg+1)
		dataArgs := append(wb.Args, size, offset)

		rows, err := db.QueryContext(ctx, dataQuery, dataArgs...)
		if err != nil {
			admin.WriteServerError(w, "failed to load data", err)
			return
		}
		defer rows.Close()

		if isCSV {
			writeCSV(w, rows)
			return
		}

		// JSON response.
		var data []json.RawMessage
		for rows.Next() {
			var raw json.RawMessage
			if err := rows.Scan(&raw); err != nil {
				admin.WriteServerError(w, "failed to load data", err)
				return
			}
			data = append(data, raw)
		}
		if err := rows.Err(); err != nil {
			admin.WriteServerError(w, "failed to load data", err)
			return
		}
		if data == nil {
			data = []json.RawMessage{}
		}

		var filterOpts map[string][]admin.FilterOption
		if filterOptsCfg != nil {
			filterOpts = queryFilterOptions(ctx, filterOptsCfg)
		}

		totalPages := (total + params.Size - 1) / params.Size
		admin.WriteJSON(w, http.StatusOK, admin.ListResponse{
			Data:          data,
			Meta:          admin.PaginationMeta{Page: params.Page, Size: params.Size, Total: total, TotalPages: totalPages},
			FilterOptions: filterOpts,
		})
	}
}

// ---------------------------------------------------------------------------
// SQL building helpers
// ---------------------------------------------------------------------------

// whereResult holds the built WHERE clause and its arguments.
type whereResult struct {
	SQL     string // includes " WHERE " prefix, or empty
	Args    []any
	NextArg int // next available $N placeholder index
}

// sqlBuildWhere turns parsed filter params into a parameterized WHERE clause.
func sqlBuildWhere(filters []query.FilterParam, filterMap map[string]SQLFilterDef) whereResult {
	var clauses []string
	var args []any
	idx := 1

	for _, fp := range filters {
		fd, ok := filterMap[fp.Field]
		if !ok {
			continue
		}
		col := `"` + fd.Column + `"`

		switch fd.Kind {
		case Search:
			clauses = append(clauses, fmt.Sprintf(`%s ILIKE $%d`, col, idx))
			args = append(args, "%"+fp.Value+"%")
			idx++

		case Exact:
			clauses = append(clauses, fmt.Sprintf(`%s = $%d`, col, idx))
			args = append(args, fp.Value)
			idx++

		case Multi:
			c, a, n := sqlMultiClause(col, fp.Values, fd.Suffix, idx)
			if c != "" {
				clauses = append(clauses, c)
				args = append(args, a...)
				idx = n
			}
		}
	}

	sql := ""
	if len(clauses) > 0 {
		sql = " WHERE " + strings.Join(clauses, " AND ")
	}
	return whereResult{SQL: sql, Args: args, NextArg: idx}
}

// sqlMultiClause builds a single WHERE clause for a Multi filter.
// Returns the clause string, args, and the next placeholder index.
func sqlMultiClause(col string, values []string, suffix bool, idx int) (string, []any, int) {
	hasEmpty := slices.Contains(values, emptyFilterValue)
	var real []string
	for _, v := range values {
		if v != emptyFilterValue {
			real = append(real, v)
		}
	}

	if suffix {
		return sqlSuffixClause(col, real, hasEmpty, idx)
	}
	return sqlExactMultiClause(col, real, hasEmpty, idx)
}

// sqlSuffixClause builds LIKE '%/<value>' predicates OR-ed together.
func sqlSuffixClause(col string, values []string, hasEmpty bool, idx int) (string, []any, int) {
	var parts []string
	var args []any

	if hasEmpty {
		parts = append(parts, fmt.Sprintf(`%s = $%d`, col, idx))
		args = append(args, "")
		idx++
	}
	for _, v := range values {
		parts = append(parts, fmt.Sprintf(`%s LIKE $%d`, col, idx))
		args = append(args, "%/"+v)
		idx++
	}

	switch len(parts) {
	case 0:
		return "", nil, idx
	case 1:
		return parts[0], args, idx
	default:
		return "(" + strings.Join(parts, " OR ") + ")", args, idx
	}
}

// sqlExactMultiClause builds an IN(...) or = clause, with optional empty match.
func sqlExactMultiClause(col string, values []string, hasEmpty bool, idx int) (string, []any, int) {
	var args []any

	switch {
	case hasEmpty && len(values) > 0:
		ph := make([]string, len(values))
		for i, v := range values {
			ph[i] = fmt.Sprintf("$%d", idx)
			args = append(args, v)
			idx++
		}
		emptyPh := fmt.Sprintf("$%d", idx)
		args = append(args, "")
		idx++
		return fmt.Sprintf(`(%s IN (%s) OR %s = %s)`, col, strings.Join(ph, ", "), col, emptyPh), args, idx

	case hasEmpty:
		args = append(args, "")
		return fmt.Sprintf(`%s = $%d`, col, idx), args, idx + 1

	case len(values) > 1:
		ph := make([]string, len(values))
		for i, v := range values {
			ph[i] = fmt.Sprintf("$%d", idx)
			args = append(args, v)
			idx++
		}
		return fmt.Sprintf(`%s IN (%s)`, col, strings.Join(ph, ", ")), args, idx

	case len(values) == 1:
		args = append(args, values[0])
		return fmt.Sprintf(`%s = $%d`, col, idx), args, idx + 1

	default:
		return "", nil, idx
	}
}

// sqlBuildOrder returns the ORDER BY clause.
func sqlBuildOrder(sort query.SortParam, defaultSort string, defaultDesc bool, allowedFields map[string]bool) string {
	if sort.Field != "" && allowedFields != nil {
		return fmt.Sprintf(` ORDER BY "%s" %s`, sort.Field, sqlDir(sort.Desc))
	}
	if defaultSort != "" {
		return fmt.Sprintf(` ORDER BY "%s" %s`, defaultSort, sqlDir(defaultDesc))
	}
	if allowedFields != nil {
		return " ORDER BY 1 DESC"
	}
	return ""
}

func sqlDir(desc bool) string {
	if desc {
		return "DESC"
	}
	return "ASC"
}

// ---------------------------------------------------------------------------
// Param helpers
// ---------------------------------------------------------------------------

func buildAllowedFields(columns []string) map[string]bool {
	if len(columns) == 0 {
		return nil
	}
	m := make(map[string]bool, len(columns))
	for _, c := range columns {
		m[c] = true
	}
	return m
}

func buildFilterMap(filters []SQLFilterDef) map[string]SQLFilterDef {
	m := make(map[string]SQLFilterDef, len(filters))
	for _, f := range filters {
		m[f.Column] = f
	}
	return m
}

// parseParams parses query params; returns nil and writes an error if invalid.
func parseParams(r *http.Request, allowedFields map[string]bool, w http.ResponseWriter) *query.Params {
	if allowedFields != nil {
		params, err := query.Parse(r, allowedFields)
		if err != nil {
			admin.WriteError(w, http.StatusBadRequest, err.Error())
			return nil
		}
		return &params
	}

	// Backwards-compatible mode: pagination only, no sort/filter.
	params := query.Params{Page: 1, Size: 20}
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			params.Page = n
		}
	}
	if v := q.Get("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
			params.Size = n
		}
	}
	return &params
}
