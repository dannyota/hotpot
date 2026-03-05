package lifecycle

import "testing"

func TestParseOSName(t *testing.T) {
	knownCycles := map[string]map[string]bool{
		"rhel":           {"7": true, "7.9": true, "8": true, "8.8": true, "9": true, "9.0": true, "9.2": true, "9.4": true, "9.6": true},
		"ubuntu":         {"18.04": true, "20.04": true, "22.04": true, "24.04": true},
		"centos":         {"7": true, "8": true},
		"debian":         {"10": true, "11": true, "12": true},
		"rocky-linux":    {"8": true, "9": true},
		"almalinux":      {"8": true, "9": true},
		"oracle-linux":   {"7": true, "8": true, "9": true},
		"amazon-linux":   {"2": true, "2023": true},
		"sles":           {"12.5": true, "15.4": true, "15.5": true},
		"windows":        {"10-22h2": true, "11-24h2-w": true, "11-25h2-w": true, "11-23h2-w": true},
		"windows-server": {"2016": true, "2019": true, "2022": true},
		"macos":          {"13": true, "14": true, "15": true},
		"opensuse":       {"15.5": true, "15.6": true},
	}

	winBuildToCycle := map[string]string{
		"19045": "10-22h2",
		"22631": "11-23h2-w",
		"26100": "11-24h2-w",
		"26200": "11-25h2-w",
	}

	tests := []struct {
		osName   string
		osType   string
		wantSlug string
		wantCyc  string
	}{
		// RHEL — prefers minor EUS cycle when known
		{"Red Hat Enterprise release 9.4 (Plow)", "linux", "rhel", "9.4"},
		{"Red Hat Enterprise release 9.3 (Plow)", "linux", "rhel", "9"},
		{"Red Hat Enterprise Linux Server release 7.9 (Maipo)", "linux", "rhel", "7.9"},
		{"RHEL 8.10", "linux", "rhel", "8"},

		// Ubuntu
		{"Ubuntu 22.04.5 LTS", "linux", "ubuntu", "22.04"},
		{"Ubuntu 20.04.6 LTS", "linux", "ubuntu", "20.04"},
		{"Ubuntu 24.04.1 LTS", "linux", "ubuntu", "24.04"},

		// CentOS
		{"CentOS Linux release 7.9.2009 (Core)", "linux", "centos", "7"},

		// Debian
		{"Debian GNU/Linux 12 (bookworm)", "linux", "debian", "12"},

		// Rocky
		{"Rocky Linux release 9.4 (Blue Onyx)", "linux", "rocky-linux", "9"},

		// AlmaLinux
		{"AlmaLinux release 9.4 (Seafoam Ocelot)", "linux", "almalinux", "9"},

		// Oracle
		{"Oracle Linux Server release 8.10", "linux", "oracle-linux", "8"},

		// Amazon
		{"Amazon Linux 2023.6.20250203", "linux", "amazon-linux", "2023"},
		{"Amazon Linux 2", "linux", "amazon-linux", "2"},

		// SLES
		{"SUSE Linux Enterprise Server 15 SP5", "linux", "sles", "15.5"},
		{"SUSE Linux Enterprise Server 15 SP4", "linux", "sles", "15.4"},
		{"SUSE Linux Enterprise Server 12 SP5", "linux", "sles", "12.5"},

		// Windows Desktop — matches build number to cycle
		{"Windows 11 Pro (Build 26200)", "windows", "windows", "11-25h2-w"},
		{"Windows 10 Enterprise (Build 19045)", "windows", "windows", "10-22h2"},
		{"Windows 11 Pro (Build 22631)", "windows", "windows", "11-23h2-w"},
		{"Windows 11 Pro (Build 26100)", "windows", "windows", "11-24h2-w"},
		// No build number → falls back to major version
		{"Windows 11 Pro", "windows", "windows", "11"},

		// Windows Server
		{"Windows Server 2019 Standard (Build 17763)", "windows", "windows-server", "2019"},
		{"Windows Server 2022 Datacenter (Build 20348)", "windows", "windows-server", "2022"},

		// macOS
		{"macOS 14.7.4", "macos", "macos", "14"},
		{"macOS 15.3.1", "macos", "macos", "15"},

		// GreenNode image-based
		{"centos-7-x86_64-genericcloud-2003", "linux", "centos", "7"},
		{"ubuntu-22.04-server-cloudimg-amd64", "linux", "ubuntu", "22.04"},
		{"rhel-9.2-x86_64-kvm", "linux", "rhel", "9.2"},
		{"Redhat-Enterprise-Linux-9.2-x86_64", "linux", "rhel", "9.2"},

		// macOS revision-only (S1 os_revision fallback)
		{"15.7.4 (24G517)", "macos", "macos", "15"},
		{"26.2 (25C56)", "macos", "macos", "26"},
		{"12.7.6 (21H1320)", "macos", "macos", "12"},

		// openSUSE
		{"openSUSE Leap 15.5", "linux", "opensuse", "15.5"},

		// Empty / unknown
		{"", "linux", "", ""},
		{"Some Custom OS", "linux", "", ""},
	}

	for _, tt := range tests {
		slug, cycle, _ := parseOSName(tt.osName, tt.osType, knownCycles, winBuildToCycle)
		if slug != tt.wantSlug || cycle != tt.wantCyc {
			t.Errorf("parseOSName(%q, %q) = (%q, %q), want (%q, %q)",
				tt.osName, tt.osType, slug, cycle, tt.wantSlug, tt.wantCyc)
		}
	}
}

func TestBuildWindowsBuildMap(t *testing.T) {
	cycles := []eolCycleInfo{
		{Product: "windows", Cycle: "11-25h2-w", Latest: "10.0.26200"},
		{Product: "windows", Cycle: "11-25h2-e", Latest: "10.0.26200"},
		{Product: "windows", Cycle: "10-22h2", Latest: "10.0.19045"},
		{Product: "windows", Cycle: "11-24h2-w", Latest: "10.0.26100"},
		{Product: "windows", Cycle: "11-24h2-iot-lts", Latest: "10.0.26100"},
	}

	m := buildWindowsBuildMap(cycles)

	// Workstation preferred over enterprise
	if got := m["26200"]; got != "11-25h2-w" {
		t.Errorf("build 26200: got %q, want %q", got, "11-25h2-w")
	}
	if got := m["19045"]; got != "10-22h2" {
		t.Errorf("build 19045: got %q, want %q", got, "10-22h2")
	}
	// Workstation preferred over IoT
	if got := m["26100"]; got != "11-24h2-w" {
		t.Errorf("build 26100: got %q, want %q", got, "11-24h2-w")
	}
}
