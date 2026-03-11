package admin

import (
	"net/http"
	"sync"
)

// RouteRegistration describes an API route.
type RouteRegistration struct {
	Method  string
	Path    string
	Handler http.HandlerFunc

	// Nav provides sidebar navigation metadata.
	// If nil, the route is not shown in the sidebar (e.g., stats endpoints).
	Nav *NavMeta
}

// NavMeta describes how a route appears in the sidebar navigation.
type NavMeta struct {
	// Label is the display text (e.g., "Compute Instances").
	Label string

	// Group is the breadcrumb path (e.g., ["Bronze", "GCP"]).
	// The first element becomes a top-level sidebar section.
	Group []string
}

var (
	routesMu sync.Mutex
	routes   []RouteRegistration
)

// RegisterRoute adds a route to the registry. Called from Register() functions.
func RegisterRoute(r RouteRegistration) {
	routesMu.Lock()
	defer routesMu.Unlock()
	routes = append(routes, r)
}

// Routes returns a copy of all registered routes.
func Routes() []RouteRegistration {
	routesMu.Lock()
	defer routesMu.Unlock()
	return append([]RouteRegistration{}, routes...)
}
