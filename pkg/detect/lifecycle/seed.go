package lifecycle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedRules upserts all default classification rules into the bronze reference tables.
// Existing user-added rules (with resource_ids not in the default set) are preserved.
func SeedRules(ctx context.Context, db *sql.DB) error {
	now := time.Now()

	if err := seedMatchRules(ctx, db, now); err != nil {
		return fmt.Errorf("seed match rules: %w", err)
	}
	if err := seedOSCoreRules(ctx, db, now); err != nil {
		return fmt.Errorf("seed os core rules: %w", err)
	}
	return nil
}

func seedMatchRules(ctx context.Context, db *sql.DB, now time.Time) error {
	type matchRule struct {
		productSlug string
		ruleType    string
		osType      string // "" = all
		value       string
		extraValue  string // for name_cycle_map
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

	var b strings.Builder
	b.WriteString(`INSERT INTO bronze.reference_software_match_rules
		(resource_id, collected_at, first_collected_at, product_slug, rule_type, os_type, value, extra_value)
		VALUES `)

	args := make([]any, 0, len(rules)*7)
	for i, r := range rules {
		if i > 0 {
			b.WriteString(", ")
		}
		resourceID := fmt.Sprintf("%s:%s:%s:%s", r.ruleType, r.productSlug, r.osType, r.value)
		base := i * 7
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7)
		var osType *string
		if r.osType != "" {
			osType = &r.osType
		}
		var extraValue *string
		if r.extraValue != "" {
			extraValue = &r.extraValue
		}
		args = append(args, resourceID, now, now, r.productSlug, r.ruleType, osType, r.value)
		_ = extraValue // handled below
	}

	// Rebuild with extra_value included.
	b.Reset()
	b.WriteString(`INSERT INTO bronze.reference_software_match_rules
		(resource_id, collected_at, first_collected_at, product_slug, rule_type, os_type, value, extra_value)
		VALUES `)
	args = args[:0]
	for i, r := range rules {
		if i > 0 {
			b.WriteString(", ")
		}
		resourceID := fmt.Sprintf("%s:%s:%s:%s", r.ruleType, r.productSlug, r.osType, r.value)
		base := i * 8
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
		var osType *string
		if r.osType != "" {
			osType = &r.osType
		}
		var extraValue *string
		if r.extraValue != "" {
			extraValue = &r.extraValue
		}
		args = append(args, resourceID, now, now, r.productSlug, r.ruleType, osType, r.value, extraValue)
	}

	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		collected_at = EXCLUDED.collected_at,
		product_slug = EXCLUDED.product_slug,
		rule_type = EXCLUDED.rule_type,
		os_type = EXCLUDED.os_type,
		value = EXCLUDED.value,
		extra_value = EXCLUDED.extra_value`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert match rules (%d rules): %w", len(rules), err)
	}
	return nil
}

func seedOSCoreRules(ctx context.Context, db *sql.DB, now time.Time) error {
	type coreRule struct {
		ruleType    string
		osType      string // "" = all
		value       string
		description string
	}

	var rules []coreRule

	// Prefixes — all OS.
	for _, p := range defaultOSCorePrefixesAll {
		rules = append(rules, coreRule{"prefix", "", p, ""})
	}

	// Prefixes — OS-specific.
	for osType, prefixes := range defaultOSCorePrefixesByOS {
		for _, p := range prefixes {
			rules = append(rules, coreRule{"prefix", osType, p, ""})
		}
	}

	// Exact — all OS.
	for name := range defaultOSCoreExactAll {
		rules = append(rules, coreRule{"exact", "", name, ""})
	}

	// Exact — OS-specific.
	for osType, names := range defaultOSCoreExactByOS {
		for name := range names {
			rules = append(rules, coreRule{"exact", osType, name, ""})
		}
	}

	// Suffixes — all OS.
	for _, s := range defaultOSCoreSuffixesAll {
		rules = append(rules, coreRule{"suffix", "", s, ""})
	}

	if len(rules) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO bronze.reference_os_core_rules
		(resource_id, collected_at, first_collected_at, rule_type, os_type, value, description)
		VALUES `)

	args := make([]any, 0, len(rules)*7)
	for i, r := range rules {
		if i > 0 {
			b.WriteString(", ")
		}
		resourceID := fmt.Sprintf("%s:%s:%s", r.ruleType, r.osType, r.value)
		base := i * 7
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7)
		var osType *string
		if r.osType != "" {
			osType = &r.osType
		}
		var description *string
		if r.description != "" {
			description = &r.description
		}
		args = append(args, resourceID, now, now, r.ruleType, osType, r.value, description)
	}

	b.WriteString(` ON CONFLICT (resource_id) DO UPDATE SET
		collected_at = EXCLUDED.collected_at,
		rule_type = EXCLUDED.rule_type,
		os_type = EXCLUDED.os_type,
		value = EXCLUDED.value,
		description = EXCLUDED.description`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert os core rules (%d rules): %w", len(rules), err)
	}
	return nil
}

// --- Default data (ported from cmd/debug-software/main.go) ---

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

var defaultOSCorePrefixesAll = []string{
	"google-cloud-cli",
	"google-cloud-sdk",
	"google-cloud-ops-",
	"google-cloud-sap-",
	"google-rhui-client",
	"gce-disk-expand",
	"gcsfuse",
}

var defaultOSCorePrefixesByOS = map[string][]string{
	"linux": {
		"linux-",
		"redhat-",
		"gpg-pubkey",
		"oem-",
		"lib",
		"mesa-lib",
	},
	"windows": {
		"microsoft visual c++",
		"intel(",
		"intel®",
		"oneapi ",
		"realtek ",
		"displaylink ",
		"thunderbolt",
		"dolby ",
		"synaptics ",
		"elan ",
		"logi ",
		"logitech ",
		"razer ",
		"steelseries ",
		"corsair ",
		"jabra ",
		"brother ",
		"fuji",
		"smart noise ",
		"smart microphone ",
		"myasus",
		"lenovo ",
		"dell ",
		"hp ",
		"thinkpad ",
		"windows pc health",
		"windows 11 ",
		"windows 10 ",
		"windows subsystem for",
		"windows sdk",
		"windows software dev",
		"windows driver package",
		"windows security",
		"windows widgets",
		"microsoft update health",
		"microsoft search in bing",
		"microsoft store",
		"microsoft edge game",
		"update for windows",
		"xbox",
		"solitaire",
		"calendar, mail",
		"global.",
		"microsoft edge",
		"microsoft onedrive",
		"microsoft teams",
		"teams machine-wide",
		"microsoft odbc driver",
		"microsoft ole db driver",
		"microsoft system clr",
		"microsoft vss writer",
		"microsoft report viewer",
		"microsoft help viewer",
		"microsoft access database",
		"microsoft asp.net",
		"microsoft silverlight",
		"microsoft web deploy",
		"microsoft mpi",
		"microsoft primary interop",
		"microsoft edge webview",
		"microsoft .net core sdk",
		"microsoft .net core runtime",
		"microsoft .net compact",
		"microsoft .net sdk",
		"microsoft azure",
		"microsoft gameinput",
		"microsoft bitlocker",
		"power automate",
		"vs_",
	},
	"macos": {},
}

var defaultOSCoreExactAll = map[string]bool{
	"gmail":         true,
	"docs":          true,
	"youtube":       true,
	"slides":        true,
	"sheets":        true,
	"outlook (pwa)": true,
}

var defaultOSCoreExactByOS = map[string]map[string]bool{
	"linux": {},
	"windows": {
		"weather":                    true,
		"maps":                       true,
		"camera":                     true,
		"calculator":                 true,
		"photos":                     true,
		"clock":                      true,
		"news":                       true,
		"tips":                       true,
		"sticky notes":               true,
		"people":                     true,
		"cortana":                    true,
		"sound recorder":             true,
		"voice recorder":             true,
		"media player":               true,
		"groove music":               true,
		"movies & tv":                true,
		"paint 3d":                   true,
		"print 3d":                   true,
		"3d viewer":                  true,
		"mixed reality portal":       true,
		"feedback hub":               true,
		"app installer":              true,
		"settings":                   true,
		"get help":                   true,
		"store experience host":      true,
		"snipping tool":              true,
		"alarms & clock":             true,
		"your phone":                 true,
		"phone link":                 true,
		"game speech window":         true,
		"game bar":                   true,
		"microsoft to do":            true,
		"microsoft bing":             true,
		"microsoft news":             true,
		"operator messages":          true,
		"desktoppackagemetadata":     true,
		"trackpoint":                 true,
		"onenote for windows 10":     true,
		"microsoft 365 (office)":     true,
		"supportassist":              true,
		"spotify widget":             true,
		"localservicecomponents":     true,
		"python launcher":            true,
		"outlook":                    true,
		"word":                       true,
		"excel":                      true,
		"powerpoint":                 true,
		"office":                     true,
		"microsoft lists":            true,
		"microsoft teams (pwa)":      true,
		"whatsapp web":               true,
		"github":                     true,
		"notebooklm":                 true,
		"postman docs":               true,
		"google password manager":    true,
	},
	"macos": {
		"safari":                                  true,
		"mail":                                    true,
		"maps":                                    true,
		"photos":                                  true,
		"music":                                   true,
		"tv":                                      true,
		"news":                                    true,
		"stocks":                                  true,
		"notes":                                   true,
		"reminders":                                true,
		"calendar":                                true,
		"contacts":                                true,
		"messages":                                true,
		"facetime":                                true,
		"freeform":                                true,
		"weather":                                 true,
		"clock":                                   true,
		"calculator":                              true,
		"passwords":                               true,
		"shortcuts":                               true,
		"preview":                                 true,
		"books":                                   true,
		"podcasts":                                true,
		"home":                                    true,
		"findmy":                                  true,
		"find my":                                 true,
		"photo booth":                             true,
		"voice memos":                             true,
		"quicktime player":                        true,
		"terminal":                                true,
		"console":                                 true,
		"activity monitor":                        true,
		"disk utility":                            true,
		"migration assistant":                     true,
		"system information":                      true,
		"bluetooth file exchange":                  true,
		"font book":                               true,
		"digital color meter":                     true,
		"grapher":                                 true,
		"screenshot":                              true,
		"stickies":                                true,
		"chess":                                   true,
		"textedit":                                true,
		"image capture":                           true,
		"automator":                               true,
		"script editor":                           true,
		"keychain access":                         true,
		"directory utility":                       true,
		"system preferences":                      true,
		"system settings":                         true,
		"app store":                               true,
		"siri":                                    true,
		"time machine":                            true,
		"iphone mirroring":                        true,
		"keynote":                                 true,
		"numbers":                                 true,
		"pages":                                   true,
		"garageband":                              true,
		"imovie":                                  true,
		"xcode":                                   true,
		"instruments":                             true,
		"filmerge":                                true,
		"accessibility inspector":                  true,
		"dictionary":                              true,
		"voiceover utility":                       true,
		"colorsync utility":                       true,
		"boot camp assistant":                     true,
		"mission control":                         true,
		"airport utility":                         true,
		"screen sharing":                          true,
		"print center":                            true,
		"tips":                                    true,
		"image playground":                        true,
		"launchpad":                               true,
		"magnifier":                               true,
		"games":                                   true,
		"journal":                                 true,
		"phone":                                   true,
		"apps":                                    true,
		"icon composer":                           true,
		"create ml":                               true,
		"reality composer pro":                    true,
		"simulator":                               true,
		"runner":                                  true,
		"xprotect":                                true,
		"mrt":                                     true,
		"applemobiledevicehelper":                  true,
		"applemobilesync":                         true,
		"mobiledeviceupdater":                     true,
		"cocoa-applescript applet":                 true,
		"droplet with settable properties":        true,
		"recursive file processing droplet":       true,
		"recursive image file processing droplet": true,
		"audio midi setup":                        true,
		"filemerge":                               true,
		"digital colour meter":                    true,
		"print centre":                            true,
		"testflight":                              true,
		"developer":                               true,
	},
}

var defaultOSCoreSuffixesAll = []string{
	"-keyring",
	"-repo",
	"-release-notes",
	"-release_notes",
}
