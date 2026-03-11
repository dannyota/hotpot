package listhandler

import (
	"context"
	"net/http"
	"slices"

	entsql "entgo.io/ent/dialect/sql"

	"danny.vn/hotpot/pkg/admin"
	"danny.vn/hotpot/pkg/admin/query"
)

// ---------------------------------------------------------------------------
// Ent handler types
// ---------------------------------------------------------------------------

// FilterDef declares a filterable field and how to build its ent predicate.
type FilterDef struct {
	Field string
	Kind  FilterKind

	// Pred builds a predicate from a single value (Search / Exact).
	Pred func(string) Predicate
	// InFn builds an IN predicate (Multi).
	InFn func(...string) Predicate
	// EqFn builds an EQ predicate (Multi — also used for empty-string match).
	EqFn func(string) Predicate
}

// QueryAdapter provides type-erased access to an ent query builder.
type QueryAdapter struct {
	Where      func(ps ...Predicate)
	CloneCount func(ctx context.Context) (int, error)
	Order      func(os ...Predicate)
	Fetch      func(ctx context.Context, offset, limit int) (any, error)
}

// Config declares everything needed for an ent-backed list endpoint.
type Config struct {
	EntityName    string              // for error messages (e.g. "servers")
	AllowedFields map[string]bool    // query-param validation
	NewQuery      func() QueryAdapter // creates a fresh query
	Filters       []FilterDef
	SortFields    map[string]SortFunc // field → ent order constructor
	DefaultOrder  Predicate           // applied when no sort param given
	FilterOptions *FilterOptionsConfig
}

// ---------------------------------------------------------------------------
// Ent handler
// ---------------------------------------------------------------------------

// Handler returns an http.HandlerFunc that implements the standard ent list
// pattern: parse → filter → count → sort → paginate → respond.
func Handler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := query.Parse(r, cfg.AllowedFields)
		if err != nil {
			admin.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		ctx := r.Context()
		q := cfg.NewQuery()

		// Apply filters.
		for _, f := range params.Filters {
			for _, fd := range cfg.Filters {
				if f.Field != fd.Field {
					continue
				}
				switch fd.Kind {
				case Search, Exact:
					q.Where(fd.Pred(f.Value))
				case Multi:
					q.Where(entMultiFilter(f.Values, fd.InFn, fd.EqFn))
				}
				break
			}
		}

		// Count total before pagination.
		total, err := q.CloneCount(ctx)
		if err != nil {
			admin.WriteServerError(w, "failed to load "+cfg.EntityName, err)
			return
		}

		// Apply sort.
		if params.Sort.Field != "" {
			if orderFn, ok := cfg.SortFields[params.Sort.Field]; ok {
				if params.Sort.Desc {
					q.Order(orderFn(entsql.OrderDesc()))
				} else {
					q.Order(orderFn())
				}
			}
		} else if cfg.DefaultOrder != nil {
			q.Order(cfg.DefaultOrder)
		}

		// Fetch page.
		data, err := q.Fetch(ctx, params.Offset(), params.Size)
		if err != nil {
			admin.WriteServerError(w, "failed to load "+cfg.EntityName, err)
			return
		}

		// Filter options.
		var filterOpts map[string][]admin.FilterOption
		if cfg.FilterOptions != nil {
			filterOpts = queryFilterOptions(ctx, cfg.FilterOptions)
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
// Ent predicate helpers
// ---------------------------------------------------------------------------

// entMultiFilter builds an ent predicate that handles the "(empty)" sentinel.
func entMultiFilter(
	values []string,
	inFn func(...string) Predicate,
	eqFn func(string) Predicate,
) Predicate {
	hasEmpty := slices.Contains(values, emptyFilterValue)
	var real []string
	for _, v := range values {
		if v != emptyFilterValue {
			real = append(real, v)
		}
	}

	switch {
	case hasEmpty && len(real) > 0:
		return entsql.OrPredicates(inFn(real...), eqFn(""))
	case hasEmpty:
		return eqFn("")
	case len(real) > 1:
		return inFn(real...)
	default:
		return eqFn(real[0])
	}
}

// Pred wraps a typed ent predicate function for use in FilterDef.
func Pred[P ~func(*entsql.Selector)](fn func(string) P) func(string) Predicate {
	return func(v string) Predicate { return fn(v) }
}

// PredIn wraps a typed ent IN predicate function for use in FilterDef.
func PredIn[P ~func(*entsql.Selector)](fn func(...string) P) func(...string) Predicate {
	return func(vs ...string) Predicate { return fn(vs...) }
}

// BoolPred wraps a typed boolean ent predicate, parsing "true"/"false" strings.
func BoolPred[P ~func(*entsql.Selector)](fn func(bool) P) func(string) Predicate {
	return func(v string) Predicate { return fn(v == "true") }
}

// Suffix wraps a HasSuffix predicate, prepending "/" so short names match full
// URL paths (e.g. "e2-medium" matches ".../machineTypes/e2-medium").
func Suffix[P ~func(*entsql.Selector)](fn func(string) P) func(string) Predicate {
	return func(v string) Predicate { return fn("/" + v) }
}

// SuffixIn builds an OR of HasSuffix predicates for multi-value suffix matching.
func SuffixIn[P ~func(*entsql.Selector)](fn func(string) P) func(...string) Predicate {
	return func(vs ...string) Predicate {
		preds := make([]Predicate, len(vs))
		for i, v := range vs {
			preds[i] = Predicate(fn("/" + v))
		}
		return entsql.OrPredicates(preds...)
	}
}

// Sort wraps a typed ent ByXxx order function for use in SortFields.
func Sort[O ~func(*entsql.Selector)](fn func(...entsql.OrderTermOption) O) SortFunc {
	return func(opts ...entsql.OrderTermOption) Predicate { return fn(opts...) }
}

// ConvertSlice converts a slice of Predicate to a typed predicate / order slice.
func ConvertSlice[T ~func(*entsql.Selector)](ps []Predicate) []T {
	out := make([]T, len(ps))
	for i, p := range ps {
		out[i] = T(p)
	}
	return out
}
