package listhandler

import (
	"context"
	"fmt"

	"danny.vn/hotpot/pkg/admin"
)

// queryFilterOptions queries DISTINCT values with counts for each configured
// column. Empty strings are collected under the "(empty)" sentinel so the
// frontend can display a user-friendly label.
func queryFilterOptions(ctx context.Context, cfg *FilterOptionsConfig) map[string][]admin.FilterOption {
	if cfg == nil || cfg.DB == nil {
		return nil
	}

	opts := make(map[string][]admin.FilterOption, len(cfg.Columns))
	for _, col := range cfg.Columns {
		// Allow callers to override the column expression (e.g. to extract
		// the short name from a GCP URL path).
		expr := fmt.Sprintf(`"%s"`, col)
		if cfg.ColumnExprs != nil {
			if e, ok := cfg.ColumnExprs[col]; ok {
				expr = e
			}
		}

		from := fmt.Sprintf(`"%s"."%s"`, cfg.Schema, cfg.Table)
		if cfg.From != "" {
			from = fmt.Sprintf(`(%s) _fo`, cfg.From)
		}
		q := fmt.Sprintf(
			`SELECT COALESCE(%s, ''), COUNT(*) FROM %s GROUP BY 1 ORDER BY 2 DESC, 1`,
			expr, from,
		)
		rows, err := cfg.DB.QueryContext(ctx, q)
		if err != nil {
			continue
		}

		var vals []admin.FilterOption
		var emptyCount int
		for rows.Next() {
			var v string
			var cnt int
			if rows.Scan(&v, &cnt) == nil {
				if v == "" {
					emptyCount = cnt
				} else {
					vals = append(vals, admin.FilterOption{Value: v, Count: cnt})
				}
			}
		}
		rows.Close()

		if emptyCount > 0 {
			vals = append(vals, admin.FilterOption{Value: emptyFilterValue, Count: emptyCount})
		}
		if len(vals) > 0 {
			opts[col] = vals
		}
	}
	return opts
}
