package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedSoftwareMatchRules upserts all default software matching rules.
// Uses ON CONFLICT DO UPDATE so that new versions can update system-managed
// fields (extra_value) while preserving user's is_active settings.
func SeedSoftwareMatchRules(ctx context.Context, db *sql.DB) error {
	type matchRule struct {
		productSlug string
		ruleType    string
		osType      string
		value       string
		extraValue  string
	}

	var rules []matchRule

	// Extra prefixes — all OS.
	for slug, prefixes := range defaultExtraPrefixesAll {
		for _, p := range prefixes {
			rules = append(rules, matchRule{slug, "extra_prefix", "", p, ""})
		}
	}

	// Extra prefixes — OS-specific.
	for osType, slugMap := range defaultExtraPrefixesByOS {
		for slug, prefixes := range slugMap {
			for _, p := range prefixes {
				rules = append(rules, matchRule{slug, "extra_prefix", osType, p, ""})
			}
		}
	}

	// Product excludes.
	for slug, excludes := range defaultProductExcludes {
		for _, ex := range excludes {
			rules = append(rules, matchRule{slug, "exclude", "", ex, ""})
		}
	}

	// Name-cycle maps.
	for slug, ncMap := range defaultProductNameCycleMaps {
		for name, cycle := range ncMap {
			rules = append(rules, matchRule{slug, "name_cycle_map", "", name, cycle})
		}
	}

	if len(rules) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.software_match_rules
		(product_slug, rule_type, os_type, value, extra_value, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(rules)*8)
	for i, r := range rules {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 8
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
		var extraValue *string
		if r.extraValue != "" {
			extraValue = &r.extraValue
		}
		args = append(args, r.productSlug, r.ruleType, r.osType, r.value, extraValue, true, now, now)
	}

	b.WriteString(` ON CONFLICT (product_slug, rule_type, os_type, value) DO UPDATE SET
		extra_value = EXCLUDED.extra_value,
		updated_at = EXCLUDED.updated_at`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert software match rules (%d rules): %w", len(rules), err)
	}
	return nil
}

// --- Default data ---

var defaultProductExcludes = map[string][]string{
	"ansible":       {"core", "test", "collection", "freeipa", "pcp"},
	"gstreamer":     {"packagekit", "nice", "clutter", "pipewire", "libcamera", "fdkaac"},
	"mongodb":       {"mongosh", "database", "compass"},
	"mysql":         {"connector", "router", "shell"},
	"mariadb":       {"connector"},
	"kubernetes":    {"cni", "csi"},
	"oracle-jdk":    {"openjdk", "common", "wrappers"},
	"visual-studio": {"code"},
	"mssqlserver":   {"management studio"},
	"unity":         {"hub"},
}

var defaultProductNameCycleMaps = map[string]map[string]string{
	"mssqlserver": {
		"2008 r2": "10.50",
		"2008":    "10.0",
		"2012":    "11.0",
		"2014":    "12.0",
		"2016":    "13.0",
		"2017":    "14.0",
		"2019":    "15.0",
		"2022":    "16.0",
		"2025":    "17.0",
	},
}

var defaultExtraPrefixesAll = map[string][]string{
	"chrome":  {"google chrome"},
	"firefox": {"firefox"},
}

var defaultExtraPrefixesByOS = map[string]map[string][]string{
	"linux": {
		"apache-http-server":       {"apache2", "httpd"},
		"docker-engine":            {"docker-ce", "docker"},
		"openssl":                  {"libssl"},
		"containerd":               {"containerd.io"},
		"kubernetes":               {"kubectl"},
		"red-hat-build-of-openjdk": {"jdk"},
		"azul-zulu":                {"zulu"},
		"firefox":                  {"mozilla firefox", "mozilla maintenance service"},
		"oracle-jdk":               {"java"},
		"nodejs":                   {"node.js"},
		"postgresql":               {"postgresql"},
		"openvpn":                  {"openvpn"},
		"unity":                    {"unity"},
	},
	"windows": {
		"chrome":            {"google chrome"},
		"firefox":           {"mozilla firefox", "mozilla maintenance service"},
		"office":            {"microsoft office", "microsoft 365 apps", "microsoft word", "microsoft excel", "microsoft powerpoint", "microsoft outlook", "microsoft onenote"},
		"visual-studio":     {"visual studio", "microsoft visual studio"},
		"dotnet":            {"microsoft .net runtime", "microsoft windows desktop runtime", "microsoft asp.net core"},
		"dotnetfx":          {"microsoft .net framework"},
		"oracle-jdk":        {"java"},
		"notepad-plus-plus": {"notepad++"},
		"nodejs":            {"node.js"},
		"mssqlserver":       {"microsoft sql server", "sql server browser for sql server"},
	},
	"macos": {
		"firefox":       {"firefox"},
		"office":        {"microsoft word", "microsoft excel", "microsoft powerpoint", "microsoft outlook", "microsoft onenote"},
		"visual-studio": {"visual studio"},
		"openvpn":       {"openvpn"},
		"unity":         {"unity"},
	},
}
