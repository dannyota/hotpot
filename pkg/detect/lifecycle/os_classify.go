package lifecycle

import (
	"regexp"
	"strings"
)

// osPattern defines how to match an os_name string to an endoflife.date product slug.
type osPattern struct {
	slug       string
	match      func(lower string) bool
	extractVer func(lower string) string
	cycleDepth int // how many version segments to keep (1 = major, 2 = major.minor); 0 = use extractVer as-is
}

var versionRe = regexp.MustCompile(`(\d+(?:\.\d+)*)`)

var osPatterns = []osPattern{
	// Windows Server must come before Windows desktop.
	{
		slug: "windows-server",
		match: func(s string) bool {
			return strings.Contains(s, "windows server") || strings.Contains(s, "windows-server")
		},
		extractVer: func(s string) string {
			if idx := strings.Index(s, "windows server"); idx >= 0 {
				return extractFirstVersion(s[idx:])
			}
			if idx := strings.Index(s, "windows-server"); idx >= 0 {
				return extractFirstVersion(s[idx:])
			}
			return ""
		},
		cycleDepth: 1,
	},
	{
		slug: "windows",
		match: func(s string) bool {
			return strings.Contains(s, "windows") &&
				!strings.Contains(s, "windows server") &&
				!strings.Contains(s, "windows-server")
		},
		extractVer: func(s string) string { return extractAfter(s, "windows") },
		cycleDepth: 1,
	},
	{
		slug:       "macos",
		match:      func(s string) bool { return strings.Contains(s, "macos") || strings.Contains(s, "mac os") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	// macOS revision-only: S1 sometimes returns just "15.7.4 (24G517)" without "macOS" prefix.
	{
		slug:       "macos",
		match:      func(s string) bool { return isMacOSRevision(s) },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "ubuntu",
		match:      func(s string) bool { return strings.Contains(s, "ubuntu") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 2,
	},
	{
		slug: "rhel",
		match: func(s string) bool {
			return strings.Contains(s, "red hat enterprise") ||
				strings.Contains(s, "redhat-enterprise") ||
				strings.Contains(s, "red-hat-enterprise") ||
				strings.HasPrefix(s, "rhel")
		},
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "centos",
		match:      func(s string) bool { return strings.Contains(s, "centos") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "rocky-linux",
		match:      func(s string) bool { return strings.Contains(s, "rocky") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "almalinux",
		match:      func(s string) bool { return strings.Contains(s, "alma") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "oracle-linux",
		match:      func(s string) bool { return strings.Contains(s, "oracle linux") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug: "amazon-linux",
		match: func(s string) bool {
			return strings.Contains(s, "amazon linux") || strings.Contains(s, "amzn")
		},
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	{
		slug:       "debian",
		match:      func(s string) bool { return strings.Contains(s, "debian") },
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 1,
	},
	// SLES: "SUSE Linux Enterprise Server 15 SP5" → cycle "15.5"
	{
		slug:  "sles",
		match: func(s string) bool { return strings.Contains(s, "suse") && strings.Contains(s, "enterprise") },
		extractVer: func(s string) string {
			ver := extractFirstVersion(s)
			if idx := strings.Index(s, " sp"); idx >= 0 {
				sp := extractFirstVersion(s[idx:])
				if sp != "" {
					return ver + "." + sp
				}
			}
			return ver
		},
		cycleDepth: 0,
	},
	{
		slug: "opensuse",
		match: func(s string) bool {
			return strings.Contains(s, "opensuse") || strings.Contains(s, "open suse")
		},
		extractVer: func(s string) string { return extractFirstVersion(s) },
		cycleDepth: 2,
	},
}

// windowsBuildRe extracts the build number from "Windows 11 Pro (Build 26200)".
var windowsBuildRe = regexp.MustCompile(`(?i)\(build\s+(\d+)\)`)

// macOSRevisionRe matches S1 os_revision format: "15.7.4 (24G517)".
var macOSRevisionRe = regexp.MustCompile(`^\d+\.\d+(?:\.\d+)?\s*\(`)

// parseOSName matches an os_name string to an endoflife.date product slug and cycle.
func parseOSName(osName, osType string, knownCycles map[string]map[string]bool, winBuildToCycle map[string]string) (slug, cycle, fullVer string) {
	if osName == "" {
		return "", "", ""
	}
	lower := strings.ToLower(osName)

	for _, p := range osPatterns {
		if !p.match(lower) {
			continue
		}

		// Windows desktop: use build number for precise cycle matching.
		if p.slug == "windows" {
			if m := windowsBuildRe.FindStringSubmatch(osName); m != nil {
				build := m[1]
				if c, ok := winBuildToCycle[build]; ok {
					return "windows", c, build
				}
			}
			ver := p.extractVer(lower)
			return p.slug, extractCycle(ver, 1), ver
		}

		ver := p.extractVer(lower)
		if ver == "" {
			return p.slug, "", ""
		}
		if p.cycleDepth > 0 {
			cycle = extractCycle(ver, p.cycleDepth)
		} else {
			cycle = ver
		}
		// Try matching against known cycles — prefer more specific (minor) over major.
		if p.cycleDepth >= 1 {
			if known, ok := knownCycles[p.slug]; ok {
				c2 := extractCycle(ver, 2)
				if known[c2] {
					return p.slug, c2, ver
				}
				c1 := extractCycle(ver, 1)
				if known[c1] {
					return p.slug, c1, ver
				}
			}
		}
		return p.slug, cycle, ver
	}
	return "", "", ""
}

// buildWindowsBuildMap creates a map from Windows build number to the best
// endoflife.date cycle name. Prefers workstation/base cycles over enterprise.
func buildWindowsBuildMap(windowsCycles []eolCycleInfo) map[string]string {
	result := make(map[string]string)
	for _, c := range windowsCycles {
		if c.Latest == "" {
			continue
		}
		parts := strings.Split(c.Latest, ".")
		if len(parts) < 3 {
			continue
		}
		build := parts[2]

		cycleName := c.Cycle
		if strings.HasSuffix(cycleName, "-e") || strings.Contains(cycleName, "-e-") || strings.Contains(cycleName, "-iot") {
			if _, ok := result[build]; !ok {
				result[build] = cycleName
			}
			continue
		}
		result[build] = cycleName
	}
	return result
}

func isMacOSRevision(s string) bool {
	return macOSRevisionRe.MatchString(s)
}

func extractAfter(s, keyword string) string {
	idx := strings.Index(s, keyword)
	if idx < 0 {
		return ""
	}
	return extractFirstVersion(s[idx+len(keyword):])
}

func extractFirstVersion(s string) string {
	return versionRe.FindString(s)
}
