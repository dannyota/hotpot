package httptraffic

import (
	"testing"
)

func TestPathMatcherExact(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/protected/app/api/users"},
		{ID: "2", URIPattern: "/public/health"},
	}
	pm := NewPathMatcher(endpoints)

	if ep := pm.Match("/protected/app/api/users"); ep == nil || ep.ID != "1" {
		t.Errorf("expected endpoint 1, got %v", ep)
	}
	if ep := pm.Match("/public/health"); ep == nil || ep.ID != "2" {
		t.Errorf("expected endpoint 2, got %v", ep)
	}
}

func TestPathMatcherWildcard(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/protected/app/api/users/*"},
		{ID: "2", URIPattern: "/protected/app/api/*/profile"},
	}
	pm := NewPathMatcher(endpoints)

	if ep := pm.Match("/protected/app/api/users/123"); ep == nil || ep.ID != "1" {
		t.Errorf("expected endpoint 1, got %v", ep)
	}
	if ep := pm.Match("/protected/app/api/orders/profile"); ep == nil || ep.ID != "2" {
		t.Errorf("expected endpoint 2, got %v", ep)
	}
}

func TestPathMatcherCatchAll(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/protected/app/api/*"},
	}
	pm := NewPathMatcher(endpoints)

	if ep := pm.Match("/protected/app/api/users"); ep == nil || ep.ID != "1" {
		t.Errorf("expected endpoint 1 for /protected/app/api/users, got %v", ep)
	}
	if ep := pm.Match("/protected/app/api/users/123/profile"); ep == nil || ep.ID != "1" {
		t.Errorf("expected endpoint 1 for deep path, got %v", ep)
	}
}

func TestPathMatcherUnmatched(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/protected/app/api/users"},
	}
	pm := NewPathMatcher(endpoints)

	if ep := pm.Match("/unknown/path"); ep != nil {
		t.Errorf("expected nil for unmatched path, got %v", ep)
	}
	if ep := pm.Match("/protected/app/api/orders"); ep != nil {
		t.Errorf("expected nil for unmatched path, got %v", ep)
	}
}

func TestPathMatcherEdgeCases(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/"},
	}
	pm := NewPathMatcher(endpoints)

	// Empty endpoints.
	pm2 := NewPathMatcher(nil)
	if ep := pm2.Match("/any/path"); ep != nil {
		t.Errorf("expected nil for empty matcher, got %v", ep)
	}

	// Root path matches root endpoint.
	if ep := pm.Match("/"); ep == nil || ep.ID != "1" {
		t.Errorf("expected endpoint 1 for root path, got %v", ep)
	}
}

func TestPathMatcherPriorityExactOverWildcard(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/api/users/*"},
		{ID: "2", URIPattern: "/api/users/me"},
	}
	pm := NewPathMatcher(endpoints)

	// Exact match should take priority over wildcard.
	if ep := pm.Match("/api/users/me"); ep == nil || ep.ID != "2" {
		t.Errorf("expected endpoint 2 (exact), got %v", ep)
	}
	if ep := pm.Match("/api/users/123"); ep == nil || ep.ID != "1" {
		t.Errorf("expected endpoint 1 (wildcard), got %v", ep)
	}
}

func TestPathMatcherMultiLevelWildcards(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/api/*/items/*/detail"},
	}
	pm := NewPathMatcher(endpoints)

	tests := []struct {
		name   string
		uri    string
		wantID string
	}{
		{"both wildcards filled", "/api/foo/items/bar/detail", "1"},
		{"numeric wildcards", "/api/123/items/456/detail", "1"},
		{"no match wrong suffix", "/api/foo/items/bar/edit", ""},
		{"no match missing segment", "/api/foo/items/detail", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ep := pm.Match(tc.uri)
			if tc.wantID == "" {
				if ep != nil {
					t.Errorf("Match(%q) = %v, want nil", tc.uri, ep)
				}
			} else if ep == nil || ep.ID != tc.wantID {
				t.Errorf("Match(%q) = %v, want endpoint %s", tc.uri, ep, tc.wantID)
			}
		})
	}
}

func TestPathMatcherCatchAllWithExact(t *testing.T) {
	// Register catch-all first, then exact. Exact should still win.
	endpoints := []MatchEndpoint{
		{ID: "catch", URIPattern: "/api/*"},
		{ID: "exact", URIPattern: "/api/users"},
	}
	pm := NewPathMatcher(endpoints)

	tests := []struct {
		name   string
		uri    string
		wantID string
	}{
		{"exact match wins over catch-all", "/api/users", "exact"},
		{"single segment matches catch-all", "/api/orders", "catch"},
		{"deep path matches catch-all", "/api/anything/deep/path", "catch"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ep := pm.Match(tc.uri)
			if ep == nil || ep.ID != tc.wantID {
				t.Errorf("Match(%q) = %v, want endpoint %s", tc.uri, ep, tc.wantID)
			}
		})
	}
}

func TestPathMatcherTrailingSlash(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/api/users"},
	}
	pm := NewPathMatcher(endpoints)

	// splitPath trims trailing slashes, so these should be equivalent.
	tests := []struct {
		name   string
		uri    string
		wantID string
	}{
		{"without trailing slash", "/api/users", "1"},
		{"with trailing slash", "/api/users/", "1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ep := pm.Match(tc.uri)
			if ep == nil || ep.ID != tc.wantID {
				t.Errorf("Match(%q) = %v, want endpoint %s", tc.uri, ep, tc.wantID)
			}
		})
	}
}

func TestPathMatcherQueryString(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/api/users"},
	}
	pm := NewPathMatcher(endpoints)

	// splitPath does NOT strip query strings. The query string becomes part of
	// the last segment, so "/api/users?page=1" will NOT match "/api/users".
	// Callers are expected to strip query strings before calling Match.
	ep := pm.Match("/api/users?page=1")
	if ep != nil {
		t.Errorf("Match with query string should not match (caller must strip query strings), got %v", ep)
	}

	// Verify the clean version still works.
	ep = pm.Match("/api/users")
	if ep == nil || ep.ID != "1" {
		t.Errorf("Match without query string should match, got %v", ep)
	}
}

func TestPathMatcherEmptySegment(t *testing.T) {
	endpoints := []MatchEndpoint{
		{ID: "1", URIPattern: "/api/users"},
	}
	pm := NewPathMatcher(endpoints)

	// Double slash creates an empty string segment: splitPath("/api//users")
	// produces ["api", "", "users"]. The empty segment "" does not match "users",
	// so the URI does not match the pattern "/api/users".
	ep := pm.Match("/api//users")
	if ep != nil {
		t.Errorf("Match with double slash should not match /api/users (empty segment mismatch), got %v", ep)
	}
}

func TestPathMatcherDuplicateRegistration(t *testing.T) {
	// Register two endpoints with the same pattern. The first one wins because
	// NewPathMatcher only sets node.endpoint when it is nil.
	endpoints := []MatchEndpoint{
		{ID: "first", URIPattern: "/api/users"},
		{ID: "second", URIPattern: "/api/users"},
	}
	pm := NewPathMatcher(endpoints)

	ep := pm.Match("/api/users")
	if ep == nil || ep.ID != "first" {
		t.Errorf("Match(%q) = %v, want endpoint 'first' (first registration wins)", "/api/users", ep)
	}
}
