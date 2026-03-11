package admin

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/base/config"
)

// RegisterAll is set by cmd/ entry points to register all admin routes.
// Called from newAPIMux with the ent driver and raw *sql.DB.
var RegisterAll func(driver dialect.Driver, db *sql.DB)

// uiConfigResponse is the JSON payload for GET /api/v1/admin/ui-config.
type uiConfigResponse struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Title       string    `json:"title"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color,omitempty"`
	Nav         []navItem `json:"nav"`
}

type navItem struct {
	Label    string    `json:"label"`
	Icon     string    `json:"icon,omitempty"`
	Path     string    `json:"path,omitempty"`
	API      string    `json:"api,omitempty"`
	Children []navItem `json:"children,omitempty"`
}

// groupIcons maps top-level group names to lucide icon names.
var groupIcons = map[string]string{
	"Bronze": "database",
	"Silver": "layers",
	"Gold":   "shield-alert",
}

// navNode is a recursive tree node used to build the sidebar nav.
type navNode struct {
	label    string
	children map[string]*navNode
	order    []string    // child insertion order
	leaves   []navItem   // leaf items at this level
}

func newNavNode(label string) *navNode {
	return &navNode{label: label, children: map[string]*navNode{}}
}

func (n *navNode) getOrCreate(label string) *navNode {
	if child, ok := n.children[label]; ok {
		return child
	}
	child := newNavNode(label)
	n.children[label] = child
	n.order = append(n.order, label)
	return child
}

func (n *navNode) toNavItem() navItem {
	item := navItem{Label: n.label, Icon: groupIcons[n.label]}
	for _, key := range n.order {
		item.Children = append(item.Children, n.children[key].toNavItem())
	}
	item.Children = append(item.Children, n.leaves...)
	return item
}

// isDisabled checks whether a route path matches any disable entry.
// Entries ending with "/" are treated as prefixes; others are exact matches.
func isDisabled(path string, disable []string) bool {
	for _, d := range disable {
		if strings.HasSuffix(d, "/") {
			if strings.HasPrefix(path, d) {
				return true
			}
		} else if path == d {
			return true
		}
	}
	return false
}

// buildNav constructs the sidebar nav tree from registered routes,
// excluding any disabled API paths.
func buildNav(disable []string) []navItem {
	// Dashboard is always first.
	nav := []navItem{{Label: "Dashboard", Icon: "layout-dashboard", Path: "/dashboard"}}

	root := newNavNode("")

	for _, r := range Routes() {
		if r.Nav == nil || isDisabled(r.Path, disable) || len(r.Nav.Group) == 0 {
			continue
		}

		// Walk the group path to find the parent node.
		current := root
		for _, g := range r.Nav.Group {
			current = current.getOrCreate(g)
		}

		// Derive frontend path: /api/v1/bronze/gcp/compute/instances → /bronze/gcp/compute/instances
		frontendPath := strings.TrimPrefix(r.Path, "/api/v1")
		current.leaves = append(current.leaves, navItem{
			Label: r.Nav.Label, Path: frontendPath, API: r.Path,
		})
	}

	// Emit top-level groups in a fixed order (Bronze → Silver → Gold),
	// then any remaining groups in registration order.
	groupOrder := []string{"Bronze", "Silver", "Gold"}
	seen := map[string]bool{}
	for _, key := range groupOrder {
		if _, ok := root.children[key]; ok {
			nav = append(nav, root.children[key].toNavItem())
			seen[key] = true
		}
	}
	for _, key := range root.order {
		if !seen[key] {
			nav = append(nav, root.children[key].toNavItem())
		}
	}

	return nav
}

// extractDB extracts *sql.DB from an ent dialect driver.
func extractDB(driver dialect.Driver) *sql.DB {
	type dbAccessor interface{ DB() *sql.DB }
	if d, ok := driver.(dbAccessor); ok {
		return d.DB()
	}
	return nil
}

// newAPIMux creates an HTTP mux with all registered API routes
// plus the built-in ui-config endpoint.
func newAPIMux(configService *config.Service, driver dialect.Driver) *http.ServeMux {
	// Register all routes via the callback set by cmd/ entry points.
	if RegisterAll != nil {
		RegisterAll(driver, extractDB(driver))
	}

	mux := http.NewServeMux()

	// Built-in: serve auto-generated UI config.
	ui := configService.AdminUIConfig()
	disable := ui.Disable

	mux.HandleFunc("GET /api/v1/admin/ui-config", func(w http.ResponseWriter, r *http.Request) {
		resp := uiConfigResponse{
			Name:        ui.Name,
			Description: ui.Description,
			Title:       ui.Title,
			Icon:        ui.Icon,
			Color:       ui.Color,
			Nav:         buildNav(disable),
		}
		WriteJSON(w, http.StatusOK, resp)
	})

	for _, r := range Routes() {
		if isDisabled(r.Path, disable) {
			continue
		}
		mux.HandleFunc(r.Method+" "+r.Path, r.Handler)
	}
	return mux
}

// serve starts an HTTP server with graceful shutdown on ctx cancellation.
func serve(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// RunAPI starts the admin API server without serving the frontend.
// Use this for development with Vite HMR handling the UI.
func RunAPI(ctx context.Context, configService *config.Service, driver dialect.Driver) error {
	mux := newAPIMux(configService, driver)
	addr := configService.AdminAddr()
	slog.Info("admin API server started", "addr", addr)
	return serve(ctx, addr, mux)
}

// Run starts the admin HTTP server with both API routes and embedded Vue SPA.
func Run(ctx context.Context, configService *config.Service, driver dialect.Driver, distFS embed.FS) error {
	mux := newAPIMux(configService, driver)

	// Serve Vue SPA from embedded filesystem.
	uiDist, err := fs.Sub(distFS, "ui/dist")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(uiDist))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: serve index.html for non-file paths.
		path := r.URL.Path
		if path != "/" {
			if _, err := fs.Stat(uiDist, path[1:]); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := configService.AdminAddr()
	slog.Info("admin server started", "addr", addr)
	return serve(ctx, addr, mux)
}
