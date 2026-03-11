package httptraffic

import "strings"

// ParseUAFamily extracts a normalized family name from a full user agent string.
func ParseUAFamily(ua string) string {
	if ua == "" {
		return ""
	}

	// Browser detection in Mozilla UA strings.
	if strings.Contains(ua, "Edg/") {
		return "edge"
	}
	if strings.Contains(ua, "Chrome/") {
		return "chrome"
	}
	if strings.Contains(ua, "Firefox/") {
		return "firefox"
	}
	if strings.Contains(ua, "Safari/") {
		return "safari"
	}

	// Non-browser UAs: take first token before "/" or " (".
	family := ua
	if idx := strings.Index(family, "/"); idx > 0 {
		family = family[:idx]
	} else if idx := strings.Index(family, " ("); idx > 0 {
		family = family[:idx]
	}

	family = strings.ToLower(strings.TrimSpace(family))
	if len(family) > 30 {
		family = family[:30]
	}
	return family
}
