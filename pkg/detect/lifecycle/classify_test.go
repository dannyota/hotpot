package lifecycle

import (
	"testing"
	"time"
)

func TestMatchAppToProduct(t *testing.T) {
	mappings := []productMapping{
		{slug: "docker-engine", prefixes: []string{"docker-ce"}, extraPrefixes: []string{"docker"}, excludes: nil},
		{slug: "firefox", prefixes: []string{"firefox"}, extraPrefixes: []string{"mozilla firefox"}},
		{slug: "mysql", prefixes: []string{"mysql"}, excludes: []string{"connector", "router", "shell"}},
		{slug: "python", prefixes: []string{"python"}, exactOnly: true},
	}
	sortMappings(mappings)

	tests := []struct {
		name    string
		appName string
		want    string // expected slug, or "" for nil
	}{
		{"exact prefix", "docker-ce-cli", "docker-engine"},
		{"extra prefix", "docker-compose", "docker-engine"},
		{"exact match on prefixes", "firefox", "firefox"},
		{"extra prefix match", "mozilla firefox", "firefox"},
		{"exclude blocks", "mysql-connector-odbc", ""},
		{"exclude allows base", "mysql-server", "mysql"},
		{"exact only match", "python", "python"},
		{"exact only rejects prefix", "python-dateutil", ""},
		{"no match", "unknown-app", ""},
		{"normalized match", "mozilla firefox (x64 en-us)", "firefox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchAppToProduct(tt.appName, mappings)
			if tt.want == "" {
				if got != nil {
					t.Errorf("matchAppToProduct(%q) = %q, want nil", tt.appName, got.slug)
				}
			} else {
				if got == nil {
					t.Errorf("matchAppToProduct(%q) = nil, want %q", tt.appName, tt.want)
				} else if got.slug != tt.want {
					t.Errorf("matchAppToProduct(%q) = %q, want %q", tt.appName, got.slug, tt.want)
				}
			}
		})
	}
}

func TestNormalizeAppName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mozilla firefox (x64 en-us)", "mozilla firefox"},
		{"postgresql17-ee-libs", "postgresql-ee-libs"},
		{"openssl3-libs", "openssl-libs"},
		{"nginx", "nginx"},
		{"google chrome (64-bit)", "google chrome"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeAppName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeAppName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripVersionSuffix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"linux-headers-6.14.0-37-generic", "linux-headers"},
		{"libpython3.10-minimal", "libpython3"},
		{"nginx", "nginx"},
		{"postgresql17-ee", "postgresql17-ee"}, // digit embedded in word, no separator before it
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripVersionSuffix(tt.input)
			if got != tt.want {
				t.Errorf("stripVersionSuffix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractCycle(t *testing.T) {
	tests := []struct {
		version string
		depth   int
		want    string
	}{
		{"1:3.4.5-1ubuntu1", 2, "3.4"},
		{"3.4.5-1ubuntu1", 1, "3"},
		{"3.4.5+dfsg-1", 2, "3.4"},
		{"10.0", 2, "10.0"},
		{"10", 2, "10"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := extractCycle(tt.version, tt.depth)
			if got != tt.want {
				t.Errorf("extractCycle(%q, %d) = %q, want %q", tt.version, tt.depth, got, tt.want)
			}
		})
	}
}

func TestExtractCycleWithFallback(t *testing.T) {
	knownCycles := map[string]bool{
		"3":    true,
		"10.0": true,
	}

	tests := []struct {
		version string
		want    string
	}{
		{"3.4.5-1ubuntu1", "3"},       // depth 1 matches
		{"10.0.1-2", "10.0"},          // depth 2 matches
		{"99.99.1-1", "99.99"},        // neither matches, returns depth 2
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := extractCycleWithFallback(tt.version, knownCycles)
			if got != tt.want {
				t.Errorf("extractCycleWithFallback(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestExtractCycleFromMapping(t *testing.T) {
	pm := &productMapping{
		slug: "mssqlserver",
		nameCycleMap: map[string]string{
			"2008 r2": "10.50",
			"2008":    "10.0",
			"2019":    "15.0",
		},
	}
	knownCycles := map[string]bool{"15.0": true, "10.0": true, "10.50": true}

	tests := []struct {
		appName string
		version string
		want    string
	}{
		{"microsoft sql server 2019", "15.0.4003.23", "15.0"},
		{"microsoft sql server 2008 r2", "10.50.1234", "10.50"},
		{"microsoft sql server 2008", "10.0.1234", "10.0"},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			got := extractCycleFromMapping(tt.appName, tt.version, pm, knownCycles)
			if got != tt.want {
				t.Errorf("extractCycleFromMapping(%q, %q) = %q, want %q", tt.appName, tt.version, got, tt.want)
			}
		})
	}
}

func TestParsePURLPackageName(t *testing.T) {
	tests := []struct {
		purl string
		want string
	}{
		{"pkg:deb/ubuntu/openssl", "openssl"},
		{"pkg:rpm/redhat/httpd", "httpd"},
		{"pkg:deb/ubuntu/libssl1.1?arch=amd64", ""},          // has dot after query strip → rejected
		{"pkg:deb", ""},                                        // too few parts
	}

	for _, tt := range tests {
		t.Run(tt.purl, func(t *testing.T) {
			got := parsePURLPackageName(tt.purl)
			if got != tt.want {
				t.Errorf("parsePURLPackageName(%q) = %q, want %q", tt.purl, got, tt.want)
			}
		})
	}
}

func TestIsOSCore(t *testing.T) {
	osCoreNames := map[string]bool{
		"libc6":         true,
		"linux-headers": true,
	}
	eolSlugs := map[string]*productMapping{
		"openssl": {slug: "openssl"},
		"python":  {slug: "python", exactOnly: true},
	}
	osCoreExact := map[string]bool{
		"gmail":   true,
		"weather": true,
	}
	osCorePrefixes := []string{
		"linux-",
		"lib",
		"google-cloud-cli",
	}
	osCoreSuffixes := []string{
		"-keyring",
		"-repo",
	}

	tests := []struct {
		name string
		want bool
	}{
		{"libc6", true},                           // exact repo name match
		{"linux-headers-6.14.0-37-generic", true}, // version-stripped match
		{"gmail", true},                           // exact rule match
		{"weather", true},                         // exact rule match
		{"linux-image", true},                     // prefix match
		{"libfoo", true},                          // prefix match
		{"google-cloud-cli", true},                // prefix match (exact)
		{"brave-keyring", true},                   // suffix match
		{"pgdg-redhat-repo", true},                // suffix match
		{"keyring-extra-data", false},             // suffix not at end, but contains "-keyring-"
		{"openssl-libs", false},                   // guarded by EOL (not exactOnly)
		{"python", false},                         // guarded by EOL (exactOnly) — but wait, exactOnly means NOT guarded
		{"python-dateutil", false},                // exactOnly → not guarded, but no prefix/suffix match either
		{"unknown-app", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOSCore(tt.name, osCoreNames, eolSlugs, osCoreExact, osCorePrefixes, osCoreSuffixes)
			if got != tt.want {
				t.Errorf("isOSCore(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGuardedByEOL(t *testing.T) {
	eolSlugs := map[string]*productMapping{
		"openssl":     {slug: "openssl"},
		"python":      {slug: "python", exactOnly: true},
		"docker":      {slug: "docker"},
	}

	tests := []struct {
		name string
		want bool
	}{
		{"openssl", true},          // exact match, not exactOnly
		{"openssl-libs", true},     // base "openssl" matches
		{"python", false},          // exactOnly → not guarded
		{"python-dateutil", false}, // base "python" is exactOnly → not guarded
		{"docker", true},           // exact match
		{"docker-ce", true},        // base "docker" matches
		{"nginx", false},           // no match
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guardedByEOL(tt.name, eolSlugs)
			if got != tt.want {
				t.Errorf("guardedByEOL(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestDetermineEOLStatus(t *testing.T) {
	now := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	past := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		eolDate  *time.Time
		eoasDate *time.Time
		eoesDate *time.Time
		want     string
	}{
		{"all nil", nil, nil, nil, "unknown"},
		{"eol expired", &past, nil, nil, "eol_expired"},
		{"eoas expired but eol future", &future, &past, nil, "eoas_expired"},
		{"eoes expired", &past, &past, &past, "eoes_expired"},
		{"eoes expired eol future", &future, &past, &past, "eoes_expired"},
		{"all future", &future, &future, &future, "active"},
		{"only eol future", &future, nil, nil, "active"},
		{"only eoes future", nil, nil, &future, "active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineEOLStatus(tt.eolDate, tt.eoasDate, tt.eoesDate, now)
			if got != tt.want {
				t.Errorf("determineEOLStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   bool
	}{
		{"docker-ce", "docker", true},     // prefix + dash separator
		{"docker ce", "docker", true},     // prefix + space separator
		{"docker", "docker", true},        // exact match
		{"dockerfoo", "docker", false},    // no separator
		{"doc", "docker", false},          // name shorter than prefix
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.prefix, func(t *testing.T) {
			got := matchPrefix(tt.name, tt.prefix)
			if got != tt.want {
				t.Errorf("matchPrefix(%q, %q) = %v, want %v", tt.name, tt.prefix, got, tt.want)
			}
		})
	}
}
