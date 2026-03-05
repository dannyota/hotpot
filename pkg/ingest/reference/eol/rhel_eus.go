package eol

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const rhelErrataURL = "https://access.redhat.com/support/policy/updates/errata"

// RHELEUSData holds parsed EUS and Enhanced EUS cycles from the Red Hat errata page.
type RHELEUSData struct {
	EUSCycles         []RHELEUSCycle
	EnhancedEUSCycles []RHELEUSCycle
}

// RHELEUSCycle holds a single parsed EUS cycle.
type RHELEUSCycle struct {
	Cycle   string     // e.g. "9.4"
	EndDate *time.Time // EUS or Enhanced EUS end date
}

// FetchRHELEUS fetches the Red Hat errata page and parses EUS dates.
func FetchRHELEUS(httpClient *http.Client) (*RHELEUSData, error) {
	resp, err := httpClient.Get(rhelErrataURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", rhelErrataURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", rhelErrataURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return ParseRHELEUS(string(body))
}

// ParseRHELEUS parses EUS and Enhanced EUS data from the Red Hat errata page content.
func ParseRHELEUS(content string) (*RHELEUSData, error) {
	data := &RHELEUSData{}

	// Split content into sections. Each section starts with a line that indicates
	// the type of support (EUS, Enhanced EUS, SAP/E4S).
	// Parse version+date pairs that follow each section header.
	sections := splitEUSSections(content)

	for _, sec := range sections {
		lower := strings.ToLower(sec.header)

		// Skip SAP/E4S sections.
		if strings.Contains(lower, "sap") || strings.Contains(lower, "e4s") {
			continue
		}

		isEnhanced := strings.Contains(lower, "enhanced")
		matches := cycleRe.FindAllStringSubmatch(sec.body, -1)
		for _, m := range matches {
			cycle := m[1]
			dateStr := m[2]

			t, err := parseErrataDate(dateStr)
			if err != nil {
				continue
			}

			entry := RHELEUSCycle{Cycle: cycle, EndDate: t}
			if isEnhanced {
				data.EnhancedEUSCycles = append(data.EnhancedEUSCycles, entry)
			} else {
				data.EUSCycles = append(data.EUSCycles, entry)
			}
		}
	}

	return data, nil
}

type eusSection struct {
	header string
	body   string
}

// eusHeaderRe matches section headers like "EUS is available for" or "Enhanced EUS is available".
var eusHeaderRe = regexp.MustCompile(`(?i)(?:enhanced\s+)?(?:eus|extended update support|e4s|sap).*(?:available|offered)`)

// cycleRe matches patterns like "9.4 (ends April 30, 2026)" or "8.4 (ended May 31, 2023)".
var cycleRe = regexp.MustCompile(`(\d+\.\d+)\s*\((?:ends?|ended)\s+(\w+\s+\d{1,2},\s*\d{4})\)`)

func splitEUSSections(content string) []eusSection {
	lines := strings.Split(content, "\n")

	var sections []eusSection
	var current *eusSection

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if eusHeaderRe.MatchString(trimmed) {
			// Start a new section.
			if current != nil {
				sections = append(sections, *current)
			}
			current = &eusSection{header: trimmed}
		} else if current != nil {
			current.body += trimmed + "\n"
		}
	}
	if current != nil {
		sections = append(sections, *current)
	}

	return sections
}

// parseErrataDate parses dates like "May 31, 2025" or "April 30, 2026".
func parseErrataDate(s string) (*time.Time, error) {
	t, err := time.Parse("January 2, 2006", s)
	if err != nil {
		t, err = time.Parse("January  2, 2006", s)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}
