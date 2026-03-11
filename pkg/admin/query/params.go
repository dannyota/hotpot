package query

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Params holds parsed query parameters for list endpoints.
type Params struct {
	Page    int
	Size    int
	Sort    SortParam
	Filters []FilterParam
}

// SortParam describes a sort field and direction.
type SortParam struct {
	Field string
	Desc  bool
}

// FilterParam describes a field filter.
type FilterParam struct {
	Field  string
	Value  string   // first value (for backward compat)
	Values []string // all values (for IN queries)
}

// Offset returns the SQL offset for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.Size
}

// Parse extracts query parameters from an HTTP request.
// allowedFields is a set of field names that can be sorted/filtered.
func Parse(r *http.Request, allowedFields map[string]bool) (Params, error) {
	q := r.URL.Query()

	p := Params{
		Page: 1,
		Size: 20,
	}

	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return p, fmt.Errorf("invalid page: %s", v)
		}
		p.Page = n
	}

	if v := q.Get("size"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 10000 {
			return p, fmt.Errorf("invalid size: %s (must be 1-10000)", v)
		}
		p.Size = n
	}

	if v := q.Get("sort"); v != "" {
		field := v
		desc := false
		if strings.HasPrefix(v, "-") {
			field = v[1:]
			desc = true
		}
		if !allowedFields[field] {
			return p, fmt.Errorf("invalid sort field: %s", field)
		}
		p.Sort = SortParam{Field: field, Desc: desc}
	}

	for key, values := range q {
		if !strings.HasPrefix(key, "filter[") || !strings.HasSuffix(key, "]") {
			continue
		}
		field := key[7 : len(key)-1]
		if !allowedFields[field] {
			return p, fmt.Errorf("invalid filter field: %s", field)
		}
		for _, v := range values {
			if v != "" {
				parts := strings.Split(v, ",")
				p.Filters = append(p.Filters, FilterParam{
					Field:  field,
					Value:  parts[0],
					Values: parts,
				})
			}
		}
	}

	return p, nil
}
