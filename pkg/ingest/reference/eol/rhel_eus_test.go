package eol

import (
	"testing"
)

func TestParseRHELEUS(t *testing.T) {
	// Simulated content from the Red Hat errata page.
	content := `
Red Hat Enterprise Linux Life Cycle

In Red Hat Enterprise Linux 8, EUS is available for:
8.4 (ended May 31, 2023)
8.6 (ended May 31, 2024)
8.8 (ends May 31, 2025)

In Red Hat Enterprise Linux 9, EUS is available for:
9.0 (ended May 31, 2024)
9.2 (ends May 31, 2025)
9.4 (ends April 30, 2026)
9.6 (ends May 31, 2027)
EUS is planned for RHEL 9.8

In Red Hat Enterprise Linux 9, Enhanced EUS is available for RHEL:
9.0 (ends May 31, 2026)
9.2 (ends May 31, 2027)
9.4 (ends April 30, 2028)
9.6 (ends May 31, 2029)
Enhanced EUS is planned for RHEL 9.8

In Red Hat Enterprise Linux 10, Enhanced EUS is available for RHEL:
10.0 (ends May 31, 2029)

SAP Solutions update services for RHEL 9 (E4S) is available for:
9.0 (ends May 31, 2026)
9.2 (ends May 31, 2027)
`

	data, err := ParseRHELEUS(content)
	if err != nil {
		t.Fatalf("ParseRHELEUS: %v", err)
	}

	// Standard EUS cycles: 8.4, 8.6, 8.8, 9.0, 9.2, 9.4, 9.6
	if got := len(data.EUSCycles); got != 7 {
		t.Errorf("EUSCycles count = %d, want 7", got)
		for _, c := range data.EUSCycles {
			t.Logf("  EUS: %s -> %v", c.Cycle, c.EndDate)
		}
	}

	// Enhanced EUS cycles: 9.0, 9.2, 9.4, 9.6, 10.0
	if got := len(data.EnhancedEUSCycles); got != 5 {
		t.Errorf("EnhancedEUSCycles count = %d, want 5", got)
		for _, c := range data.EnhancedEUSCycles {
			t.Logf("  Enhanced: %s -> %v", c.Cycle, c.EndDate)
		}
	}

	// Check specific dates.
	eusMap := make(map[string]string)
	for _, c := range data.EUSCycles {
		eusMap[c.Cycle] = c.EndDate.Format("2006-01-02")
	}

	wantEUS := map[string]string{
		"8.4": "2023-05-31",
		"8.6": "2024-05-31",
		"8.8": "2025-05-31",
		"9.0": "2024-05-31",
		"9.2": "2025-05-31",
		"9.4": "2026-04-30",
		"9.6": "2027-05-31",
	}
	for cycle, want := range wantEUS {
		if got, ok := eusMap[cycle]; !ok {
			t.Errorf("EUS cycle %s not found", cycle)
		} else if got != want {
			t.Errorf("EUS cycle %s date = %s, want %s", cycle, got, want)
		}
	}

	eEUSMap := make(map[string]string)
	for _, c := range data.EnhancedEUSCycles {
		eEUSMap[c.Cycle] = c.EndDate.Format("2006-01-02")
	}

	wantEnhanced := map[string]string{
		"9.0":  "2026-05-31",
		"9.2":  "2027-05-31",
		"9.4":  "2028-04-30",
		"9.6":  "2029-05-31",
		"10.0": "2029-05-31",
	}
	for cycle, want := range wantEnhanced {
		if got, ok := eEUSMap[cycle]; !ok {
			t.Errorf("Enhanced EUS cycle %s not found", cycle)
		} else if got != want {
			t.Errorf("Enhanced EUS cycle %s date = %s, want %s", cycle, got, want)
		}
	}

	// SAP/E4S lines should NOT appear in standard EUS.
	for _, c := range data.EUSCycles {
		if c.Cycle == "9.0" {
			if c.EndDate.Format("2006-01-02") == "2026-05-31" {
				t.Errorf("SAP/E4S date leaked into standard EUS for cycle %s", c.Cycle)
			}
		}
	}
}
