package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedOSCoreRules upserts all default OS core classification rules.
// Uses ON CONFLICT DO UPDATE so that new versions can update system-managed
// fields (description) while preserving user's is_active settings.
func SeedOSCoreRules(ctx context.Context, db *sql.DB) error {
	type coreRule struct {
		ruleType string
		osType   string
		value    string
	}

	var rules []coreRule

	// Prefixes — all OS.
	for _, p := range defaultOSCorePrefixesAll {
		rules = append(rules, coreRule{"prefix", "", p})
	}

	// Prefixes — OS-specific.
	for osType, prefixes := range defaultOSCorePrefixesByOS {
		for _, p := range prefixes {
			rules = append(rules, coreRule{"prefix", osType, p})
		}
	}

	// Exact — all OS.
	for name := range defaultOSCoreExactAll {
		rules = append(rules, coreRule{"exact", "", name})
	}

	// Exact — OS-specific.
	for osType, names := range defaultOSCoreExactByOS {
		for name := range names {
			rules = append(rules, coreRule{"exact", osType, name})
		}
	}

	// Suffixes — all OS.
	for _, s := range defaultOSCoreSuffixesAll {
		rules = append(rules, coreRule{"suffix", "", s})
	}

	if len(rules) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.os_core_rules
		(rule_type, os_type, value, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(rules)*6)
	for i, r := range rules {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 6
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6)
		args = append(args, r.ruleType, r.osType, r.value, true, now, now)
	}

	b.WriteString(` ON CONFLICT (rule_type, os_type, value) DO UPDATE SET
		updated_at = EXCLUDED.updated_at`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert os core rules (%d rules): %w", len(rules), err)
	}
	return nil
}

// --- Default data ---

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
		"weather":                 true,
		"maps":                    true,
		"camera":                  true,
		"calculator":              true,
		"photos":                  true,
		"clock":                   true,
		"news":                    true,
		"tips":                    true,
		"sticky notes":            true,
		"people":                  true,
		"cortana":                 true,
		"sound recorder":          true,
		"voice recorder":          true,
		"media player":            true,
		"groove music":            true,
		"movies & tv":             true,
		"paint 3d":                true,
		"print 3d":                true,
		"3d viewer":               true,
		"mixed reality portal":    true,
		"feedback hub":            true,
		"app installer":           true,
		"settings":                true,
		"get help":                true,
		"store experience host":   true,
		"snipping tool":           true,
		"alarms & clock":          true,
		"your phone":              true,
		"phone link":              true,
		"game speech window":      true,
		"game bar":                true,
		"microsoft to do":         true,
		"microsoft bing":          true,
		"microsoft news":          true,
		"operator messages":       true,
		"desktoppackagemetadata":  true,
		"trackpoint":              true,
		"onenote for windows 10":  true,
		"microsoft 365 (office)":  true,
		"supportassist":           true,
		"spotify widget":          true,
		"localservicecomponents":  true,
		"python launcher":         true,
		"outlook":                 true,
		"word":                    true,
		"excel":                   true,
		"powerpoint":              true,
		"office":                  true,
		"microsoft lists":         true,
		"microsoft teams (pwa)":   true,
		"whatsapp web":            true,
		"github":                  true,
		"notebooklm":              true,
		"postman docs":            true,
		"google password manager": true,
	},
	"macos": {
		"safari":                            true,
		"mail":                              true,
		"maps":                              true,
		"photos":                            true,
		"music":                             true,
		"tv":                                true,
		"news":                              true,
		"stocks":                            true,
		"notes":                             true,
		"reminders":                         true,
		"calendar":                          true,
		"contacts":                          true,
		"messages":                          true,
		"facetime":                          true,
		"freeform":                          true,
		"weather":                           true,
		"clock":                             true,
		"calculator":                        true,
		"passwords":                         true,
		"shortcuts":                         true,
		"preview":                           true,
		"books":                             true,
		"podcasts":                          true,
		"home":                              true,
		"findmy":                            true,
		"find my":                           true,
		"photo booth":                       true,
		"voice memos":                       true,
		"quicktime player":                  true,
		"terminal":                          true,
		"console":                           true,
		"activity monitor":                  true,
		"disk utility":                      true,
		"migration assistant":               true,
		"system information":                true,
		"bluetooth file exchange":           true,
		"font book":                         true,
		"digital color meter":               true,
		"grapher":                           true,
		"screenshot":                        true,
		"stickies":                          true,
		"chess":                             true,
		"textedit":                          true,
		"image capture":                     true,
		"automator":                         true,
		"script editor":                     true,
		"keychain access":                   true,
		"directory utility":                 true,
		"system preferences":                true,
		"system settings":                   true,
		"app store":                         true,
		"siri":                              true,
		"time machine":                      true,
		"iphone mirroring":                  true,
		"keynote":                           true,
		"numbers":                           true,
		"pages":                             true,
		"garageband":                        true,
		"imovie":                            true,
		"xcode":                             true,
		"instruments":                       true,
		"filmerge":                          true,
		"accessibility inspector":           true,
		"dictionary":                        true,
		"voiceover utility":                 true,
		"colorsync utility":                 true,
		"boot camp assistant":               true,
		"mission control":                   true,
		"airport utility":                   true,
		"screen sharing":                    true,
		"print center":                      true,
		"tips":                              true,
		"image playground":                  true,
		"launchpad":                         true,
		"magnifier":                         true,
		"games":                             true,
		"journal":                           true,
		"phone":                             true,
		"apps":                              true,
		"icon composer":                     true,
		"create ml":                         true,
		"reality composer pro":              true,
		"simulator":                         true,
		"runner":                            true,
		"xprotect":                          true,
		"mrt":                               true,
		"applemobiledevicehelper":            true,
		"applemobilesync":                    true,
		"mobiledeviceupdater":               true,
		"cocoa-applescript applet":           true,
		"droplet with settable properties":  true,
		"recursive file processing droplet": true,
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
