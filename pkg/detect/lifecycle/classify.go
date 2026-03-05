package lifecycle

import (
	"sort"
	"strings"
	"time"
)

// productMapping holds the matching configuration for one endoflife.date product.
type productMapping struct {
	slug          string
	name          string
	eolCategory   string
	prefixes      []string // PURL/repology-derived — subject to exactOnly
	extraPrefixes []string // manually added — always prefix-matched
	excludes      []string
	exactOnly     bool
	nameCycleMap  map[string]string // extract cycle from app name instead of version
}

// eolCycleInfo holds the EOL dates for one product cycle.
type eolCycleInfo struct {
	Product string
	Cycle   string
	EOL     *time.Time
	EOAS    *time.Time
	EOES    *time.Time
	Latest  string
}

// matchAppToProduct returns the first matching product for the given app name,
// or nil if no match is found. Mappings must be sorted by slug (first match wins).
func matchAppToProduct(name string, mappings []productMapping) *productMapping {
	if m := matchName(name, mappings); m != nil {
		return m
	}
	if norm := normalizeAppName(name); norm != name {
		return matchName(norm, mappings)
	}
	return nil
}

func matchName(name string, mappings []productMapping) *productMapping {
	for i := range mappings {
		for _, prefix := range mappings[i].prefixes {
			if mappings[i].exactOnly {
				if name == prefix {
					return &mappings[i]
				}
				continue
			}
			if matchWithExcludes(name, prefix, mappings[i].excludes) {
				return &mappings[i]
			}
		}
		for _, prefix := range mappings[i].extraPrefixes {
			if matchWithExcludes(name, prefix, mappings[i].excludes) {
				return &mappings[i]
			}
		}
	}
	return nil
}

func matchWithExcludes(name, prefix string, excludes []string) bool {
	if !matchPrefix(name, prefix) {
		return false
	}
	if len(excludes) > 0 && len(name) > len(prefix)+1 {
		suffix := name[len(prefix)+1:]
		for _, ex := range excludes {
			for _, seg := range splitSegments(suffix) {
				if seg == ex {
					return false
				}
			}
		}
	}
	return true
}

func splitSegments(s string) []string {
	f := func(c rune) bool { return c == '-' || c == ' ' }
	return strings.FieldsFunc(s, f)
}

// matchPrefix checks if name starts with prefix followed by a separator (- or space) or is exact.
func matchPrefix(name, prefix string) bool {
	if name == prefix {
		return true
	}
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	sep := name[len(prefix)]
	return sep == '-' || sep == ' '
}

// normalizeAppName strips platform suffixes and embedded version digits.
//
//	"mozilla firefox (x64 en-us)" → "mozilla firefox"
//	"postgresql17-ee-libs"        → "postgresql-ee-libs"
//	"openssl3-libs"               → "openssl-libs"
//	"nginx"                       → "nginx" (unchanged)
func normalizeAppName(name string) string {
	for strings.HasSuffix(name, ")") {
		idx := strings.LastIndex(name, " (")
		if idx < 0 {
			break
		}
		name = strings.TrimSpace(name[:idx])
	}
	for i := 1; i < len(name); i++ {
		if name[i] >= '0' && name[i] <= '9' && name[i-1] >= 'a' && name[i-1] <= 'z' {
			j := i
			for j < len(name) && ((name[j] >= '0' && name[j] <= '9') ||
				(name[j] == '.' && j+1 < len(name) && name[j+1] >= '0' && name[j+1] <= '9')) {
				j++
			}
			if j == len(name) || name[j] == '-' || name[j] == ' ' {
				name = name[:i] + name[j:]
				break
			}
		}
	}
	return strings.TrimRight(name, "- .")
}

// extractCycleFromMapping tries nameCycleMap first (for products like SQL Server
// where the cycle is in the app name), then falls back to version parsing.
func extractCycleFromMapping(appName, version string, pm *productMapping, knownCycles map[string]bool) string {
	if pm.nameCycleMap != nil {
		for keyword, cycle := range pm.nameCycleMap {
			if strings.Contains(appName, keyword) {
				bestKeyword, bestCycle := keyword, cycle
				for k2, c2 := range pm.nameCycleMap {
					if len(k2) > len(bestKeyword) && strings.Contains(appName, k2) {
						bestKeyword, bestCycle = k2, c2
					}
				}
				return bestCycle
			}
		}
	}
	return extractCycleWithFallback(version, knownCycles)
}

func extractCycleWithFallback(version string, knownCycles map[string]bool) string {
	c1 := extractCycle(version, 1)
	if knownCycles[c1] {
		return c1
	}
	c2 := extractCycle(version, 2)
	if knownCycles[c2] {
		return c2
	}
	return c2
}

func extractCycle(version string, depth int) string {
	v := version
	if idx := strings.IndexByte(v, ':'); idx >= 0 {
		v = v[idx+1:]
	}
	if idx := strings.IndexByte(v, '-'); idx >= 0 {
		v = v[:idx]
	}
	if idx := strings.IndexByte(v, '+'); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) < depth {
		return v
	}
	return strings.Join(parts[:depth], ".")
}

func parsePURLPackageName(purl string) string {
	parts := strings.SplitN(purl, "/", 3)
	if len(parts) < 3 {
		return ""
	}
	name := parts[2]
	if idx := strings.IndexByte(name, '?'); idx >= 0 {
		name = name[:idx]
	}
	if strings.HasPrefix(name, "/") || strings.Contains(name, ".") {
		return ""
	}
	return name
}

// isOSCore checks if a name is OS core by exact match, version-stripped prefix,
// or known OS core patterns. The osCoreSuffixes parameter replaces the old
// package-level osCoreSuffixesAll var.
func isOSCore(name string, osCoreNames map[string]bool, eolSlugs map[string]*productMapping, osCoreExact map[string]bool, osCorePrefixes []string, osCoreSuffixes []string) bool {
	if osCoreNames[name] || osCoreExact[name] {
		return true
	}
	if base := stripVersionSuffix(name); base != name && osCoreNames[base] {
		return true
	}
	if guardedByEOL(name, eolSlugs) {
		return false
	}
	for _, p := range osCorePrefixes {
		if strings.HasPrefix(name, p) || name == p {
			return true
		}
	}
	for _, s := range osCoreSuffixes {
		if strings.HasSuffix(name, s) || strings.Contains(name, s+"-") {
			return true
		}
	}
	return false
}

// guardedByEOL returns true if the name could match an EOL product,
// meaning pattern-based OS core filters (prefix/suffix) should not apply.
func guardedByEOL(name string, eolSlugs map[string]*productMapping) bool {
	if m, ok := eolSlugs[name]; ok && !m.exactOnly {
		return true
	}
	if base := strings.SplitN(name, "-", 2)[0]; base != name {
		if m, ok := eolSlugs[base]; ok && !m.exactOnly {
			return true
		}
	}
	return false
}

// stripVersionSuffix removes version numbers after a separator from package names.
// "linux-headers-6.14.0-37-generic" → "linux-headers"
// "libpython3.10-minimal" → "libpython"
func stripVersionSuffix(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] >= '0' && name[i] <= '9' {
			if i > 0 && (name[i-1] == '-' || name[i-1] == '.') {
				return strings.TrimRight(name[:i-1], "-.")
			}
		}
	}
	return name
}

// determineEOLStatus computes the EOL status based on dates.
// Order: eoes_expired > eol_expired > eoas_expired > active > unknown.
func determineEOLStatus(eolDate, eoasDate, eoesDate *time.Time, now time.Time) string {
	if eoesDate != nil && !eoesDate.After(now) {
		return "eoes_expired"
	}
	if eolDate != nil && !eolDate.After(now) {
		return "eol_expired"
	}
	if eoasDate != nil && !eoasDate.After(now) {
		return "eoas_expired"
	}
	if eolDate != nil || eoasDate != nil || eoesDate != nil {
		return "active"
	}
	return "unknown"
}

func sortMappings(m []productMapping) {
	sort.Slice(m, func(i, j int) bool {
		return m[i].slug < m[j].slug
	})
}
