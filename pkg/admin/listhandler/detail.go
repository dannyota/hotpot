package listhandler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"danny.vn/hotpot/pkg/admin"
	"danny.vn/hotpot/pkg/admin/query"
)

// ---------------------------------------------------------------------------
// Detail handler types
// ---------------------------------------------------------------------------

// SQLDetail defines a detail endpoint backed by raw SQL.
// It mirrors the SQLTable / RegisterSQL pattern for list endpoints.
type SQLDetail struct {
	API      string // base list API, e.g. "/api/v1/bronze/gcp/compute/instances"
	Schema   string // "bronze"
	Table    string // "gcp_compute_instances"
	IDColumn string // default "resource_id"

	// Edges are child tables included inline in the detail response.
	Edges []SQLDetailEdge

	// Related are cross-resource endpoints served as separate paginated lists.
	Related []SQLRelated
}

// SQLDetailEdge defines a child table whose rows are included inline
// in the detail response under the "edges" key.
type SQLDetailEdge struct {
	Key      string // JSON key in response, e.g. "disks"
	Schema   string // defaults to parent Schema
	Table    string // e.g. "gcp_compute_instance_disks"
	FKColumn string // FK referencing parent ID, e.g. "bronze_gcp_compute_instance_disks"
	OrderBy  string // optional, e.g. "index"
}

// SQLRelated defines a cross-resource related endpoint served as a separate
// paginated list at GET {API}/{id}/{Key}.
type SQLRelated struct {
	Key         string         // URL segment, e.g. "disks" → GET {API}/{id}/disks
	Schema      string         // PG schema
	Table       string         // PG table (used for filter options)
	Columns     []string       // whitelisted columns for sort/filter
	Filters     []SQLFilterDef // filterable columns
	DefaultSort string
	DefaultDesc bool
	// From is a custom SELECT with $1 = parent ID placeholder.
	// Used as a subquery: SELECT row_to_json(t) FROM (From) t WHERE ... ORDER BY ... LIMIT ... OFFSET ...
	From string
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

// RegisterSQLDetail registers detail and related endpoints for a resource.
func RegisterSQLDetail(db *sql.DB, cfg SQLDetail) {
	if cfg.IDColumn == "" {
		cfg.IDColumn = "resource_id"
	}

	// Validate edge definitions.
	for _, e := range cfg.Edges {
		if !validColumn.MatchString(e.FKColumn) {
			panic(fmt.Sprintf("listhandler: invalid FKColumn %q in edge %q", e.FKColumn, e.Key))
		}
		if e.OrderBy != "" && !validColumn.MatchString(e.OrderBy) {
			panic(fmt.Sprintf("listhandler: invalid OrderBy %q in edge %q", e.OrderBy, e.Key))
		}
	}

	// Validate related definitions.
	for _, rel := range cfg.Related {
		for _, col := range rel.Columns {
			if !validColumn.MatchString(col) {
				panic(fmt.Sprintf("listhandler: invalid column %q in related %q", col, rel.Key))
			}
		}
		if rel.DefaultSort != "" && !validColumn.MatchString(rel.DefaultSort) {
			panic(fmt.Sprintf("listhandler: invalid DefaultSort %q in related %q", rel.DefaultSort, rel.Key))
		}
		for _, f := range rel.Filters {
			if !validColumn.MatchString(f.Column) {
				panic(fmt.Sprintf("listhandler: invalid filter column %q in related %q", f.Column, rel.Key))
			}
		}
	}

	// GET {API}/{id} — detail with edges
	admin.RegisterRoute(admin.RouteRegistration{
		Method:  "GET",
		Path:    cfg.API + "/{id}",
		Handler: sqlDetailHandler(db, cfg),
	})

	// GET {API}/{id}/{key} — related paginated lists
	for _, rel := range cfg.Related {
		rel := rel
		admin.RegisterRoute(admin.RouteRegistration{
			Method:  "GET",
			Path:    cfg.API + "/{id}/" + rel.Key,
			Handler: sqlRelatedHandler(db, rel),
		})
	}
}

// ---------------------------------------------------------------------------
// Detail handler
// ---------------------------------------------------------------------------

func sqlDetailHandler(db *sql.DB, cfg SQLDetail) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			admin.WriteError(w, http.StatusBadRequest, "missing id")
			return
		}

		ctx := r.Context()

		// Query main row.
		var raw json.RawMessage
		q := fmt.Sprintf(`SELECT row_to_json(t) FROM "%s"."%s" t WHERE "%s" = $1`,
			cfg.Schema, cfg.Table, cfg.IDColumn)
		err := db.QueryRowContext(ctx, q, id).Scan(&raw)
		if err == sql.ErrNoRows {
			admin.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			admin.WriteServerError(w, "failed to load detail", err)
			return
		}

		// Decode main row so we can add edges.
		var result map[string]any
		if err := json.Unmarshal(raw, &result); err != nil {
			admin.WriteServerError(w, "failed to decode detail", err)
			return
		}

		// Query each edge.
		edges := make(map[string]any, len(cfg.Edges))
		for _, e := range cfg.Edges {
			schema := e.Schema
			if schema == "" {
				schema = cfg.Schema
			}
			orderBy := ""
			if e.OrderBy != "" {
				orderBy = fmt.Sprintf(` ORDER BY "%s"`, e.OrderBy)
			}

			eq := fmt.Sprintf(`SELECT row_to_json(t) FROM "%s"."%s" t WHERE "%s" = $1%s`,
				schema, e.Table, e.FKColumn, orderBy)
			rows, err := db.QueryContext(ctx, eq, id)
			if err != nil {
				admin.WriteServerError(w, "failed to load edges", err)
				return
			}

			edgeData, err := scanJSONRows(rows)
			rows.Close()
			if err != nil {
				admin.WriteServerError(w, "failed to load edges", err)
				return
			}
			edges[e.Key] = edgeData
		}

		result["edges"] = edges
		admin.WriteJSON(w, http.StatusOK, admin.DetailResponse{Data: result})
	}
}

// ---------------------------------------------------------------------------
// Related handler (paginated list scoped to parent)
// ---------------------------------------------------------------------------

func sqlRelatedHandler(db *sql.DB, rel SQLRelated) http.HandlerFunc {
	allowedFields := buildAllowedFields(rel.Columns)
	filterMap := buildFilterMap(rel.Filters)

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			admin.WriteError(w, http.StatusBadRequest, "missing id")
			return
		}

		ctx := r.Context()

		params := parseParams(r, allowedFields, w)
		if params == nil {
			return
		}

		// Build WHERE/ORDER starting from $2 since $1 is parent ID.
		wb := sqlBuildWhereIdx(params.Filters, filterMap, 2)
		orderSQL := sqlBuildOrder(params.Sort, rel.DefaultSort, rel.DefaultDesc, allowedFields)

		// The From clause is a complete SELECT with $1 for parent ID.
		// Wrap as subquery for counting and data retrieval.
		dataFrom := fmt.Sprintf(`(%s) t`, rel.From)
		countFrom := dataFrom

		// Count.
		var total int
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s%s`, countFrom, wb.SQL)
		countArgs := append([]any{id}, wb.Args...)
		if err := db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
			admin.WriteServerError(w, "failed to count related", err)
			return
		}

		size, offset := params.Size, params.Offset()

		// Data query.
		dataQuery := fmt.Sprintf(`SELECT row_to_json(t) FROM %s%s%s LIMIT $%d OFFSET $%d`,
			dataFrom, wb.SQL, orderSQL, wb.NextArg, wb.NextArg+1)
		dataArgs := append([]any{id}, wb.Args...)
		dataArgs = append(dataArgs, size, offset)

		rows, err := db.QueryContext(ctx, dataQuery, dataArgs...)
		if err != nil {
			admin.WriteServerError(w, "failed to load related", err)
			return
		}
		defer rows.Close()

		data, err := scanJSONRows(rows)
		if err != nil {
			admin.WriteServerError(w, "failed to load related", err)
			return
		}

		totalPages := (total + params.Size - 1) / params.Size
		admin.WriteJSON(w, http.StatusOK, admin.ListResponse{
			Data: data,
			Meta: admin.PaginationMeta{Page: params.Page, Size: params.Size, Total: total, TotalPages: totalPages},
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// scanJSONRows reads all row_to_json rows from a result set.
func scanJSONRows(rows *sql.Rows) ([]json.RawMessage, error) {
	var data []json.RawMessage
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		data = append(data, raw)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if data == nil {
		data = []json.RawMessage{}
	}
	return data, nil
}

// sqlBuildWhereIdx is like sqlBuildWhere but starts placeholder indices at startIdx.
func sqlBuildWhereIdx(filters []query.FilterParam, filterMap map[string]SQLFilterDef, startIdx int) whereResult {
	var clauses []string
	var args []any
	idx := startIdx

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
