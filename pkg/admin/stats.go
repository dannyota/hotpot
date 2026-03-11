package admin

import (
	"fmt"
	"net/http"
	"strings"
)

// StatsFilter describes how a filter param maps to a SQL column.
type StatsFilter struct {
	Column string // SQL column name
	Suffix bool   // if true, use LIKE '%/<value>' instead of = (for GCP URL paths)
	Expr   string // optional SQL expression to use instead of Column for comparisons
}

// StatsWhere builds a parameterized WHERE clause from filter[x] query parameters.
// Returns the WHERE clause (including " WHERE " prefix, or empty string) and args.
func StatsWhere(r *http.Request, allowed map[string]StatsFilter) (string, []any) {
	var conditions []string
	var args []any
	argIdx := 1

	for param, values := range r.URL.Query() {
		if !strings.HasPrefix(param, "filter[") || !strings.HasSuffix(param, "]") {
			continue
		}
		field := param[7 : len(param)-1]
		value := values[0]
		if value == "" {
			continue
		}

		sf, ok := allowed[field]
		if !ok {
			continue
		}

		vals := strings.Split(value, ",")

		// Use Expr if provided, otherwise quote the column name.
		col := fmt.Sprintf(`"%s"`, sf.Column)
		if sf.Expr != "" {
			col = sf.Expr
		}

		if sf.Suffix {
			// OR of LIKE '%/<val>' for each value
			var parts []string
			for _, v := range vals {
				parts = append(parts, fmt.Sprintf(`%s LIKE $%d`, col, argIdx))
				args = append(args, "%/"+v)
				argIdx++
			}
			if len(parts) == 1 {
				conditions = append(conditions, parts[0])
			} else {
				conditions = append(conditions, "("+strings.Join(parts, " OR ")+")")
			}
		} else if len(vals) == 1 {
			conditions = append(conditions, fmt.Sprintf(`%s = $%d`, col, argIdx))
			args = append(args, vals[0])
			argIdx++
		} else {
			placeholders := make([]string, len(vals))
			for i, v := range vals {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, v)
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf(`%s IN (%s)`, col, strings.Join(placeholders, ",")))
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
