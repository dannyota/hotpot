package main

import "testing"

func TestParseOSName(t *testing.T) {
	// Simulate known cycles from endoflife.date.
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

	// Simulate Windows build → cycle map (from endoflife.date latest field).
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
		// S1 Linux: RHEL — prefers minor EUS cycle when known
		{"Red Hat Enterprise release 9.4 (Plow)", "linux", "rhel", "9.4"},
		{"Red Hat Enterprise release 9.3 (Plow)", "linux", "rhel", "9"},   // 9.3 not in EUS → falls back to major
		{"Red Hat Enterprise Linux Server release 7.9 (Maipo)", "linux", "rhel", "7.9"},
		{"RHEL 8.10", "linux", "rhel", "8"},   // 8.10 not in EUS → falls back to major

		// S1 Linux: Ubuntu
		{"Ubuntu 22.04.5 LTS", "linux", "ubuntu", "22.04"},
		{"Ubuntu 20.04.6 LTS", "linux", "ubuntu", "20.04"},
		{"Ubuntu 24.04.1 LTS", "linux", "ubuntu", "24.04"},

		// S1 Linux: CentOS
		{"CentOS Linux release 7.9.2009 (Core)", "linux", "centos", "7"},

		// S1 Linux: Debian
		{"Debian GNU/Linux 12 (bookworm)", "linux", "debian", "12"},

		// S1 Linux: Rocky
		{"Rocky Linux release 9.4 (Blue Onyx)", "linux", "rocky-linux", "9"},

		// S1 Linux: AlmaLinux
		{"AlmaLinux release 9.4 (Seafoam Ocelot)", "linux", "almalinux", "9"},

		// S1 Linux: Oracle
		{"Oracle Linux Server release 8.10", "linux", "oracle-linux", "8"},

		// S1 Linux: Amazon
		{"Amazon Linux 2023.6.20250203", "linux", "amazon-linux", "2023"},
		{"Amazon Linux 2", "linux", "amazon-linux", "2"},

		// S1 Linux: SLES
		{"SUSE Linux Enterprise Server 15 SP5", "linux", "sles", "15.5"},
		{"SUSE Linux Enterprise Server 15 SP4", "linux", "sles", "15.4"},
		{"SUSE Linux Enterprise Server 12 SP5", "linux", "sles", "12.5"},

		// S1 Windows: Desktop — matches build number to cycle
		{"Windows 11 Pro (Build 26200)", "windows", "windows", "11-25h2-w"},
		{"Windows 10 Enterprise (Build 19045)", "windows", "windows", "10-22h2"},
		{"Windows 11 Pro (Build 22631)", "windows", "windows", "11-23h2-w"},
		{"Windows 11 Pro (Build 26100)", "windows", "windows", "11-24h2-w"},
		// No build number → falls back to major version
		{"Windows 11 Pro", "windows", "windows", "11"},

		// S1 Windows: Server
		{"Windows Server 2019 Standard (Build 17763)", "windows", "windows-server", "2019"},
		{"Windows Server 2022 Datacenter (Build 20348)", "windows", "windows-server", "2022"},

		// S1 macOS
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
