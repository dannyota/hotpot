package machine

import (
	"sort"
	"testing"

	"danny.vn/hotpot/pkg/normalize/inventory/mergeutil"
)

func TestNamespacedKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]string
		expected []string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string][]string{},
			expected: nil,
		},
		{
			name:     "single namespace single value",
			input:    map[string][]string{"mac": {"AA:BB:CC:DD:EE:FF"}},
			expected: []string{"mac:AA:BB:CC:DD:EE:FF"},
		},
		{
			name:     "single namespace multiple values",
			input:    map[string][]string{"mac": {"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"}},
			expected: []string{"mac:AA:BB:CC:DD:EE:FF", "mac:11:22:33:44:55:66"},
		},
		{
			name: "multiple namespaces",
			input: map[string][]string{
				"mac":      {"AA:BB:CC:DD:EE:FF"},
				"hostname": {"server-01"},
			},
			expected: []string{"hostname:server-01", "mac:AA:BB:CC:DD:EE:FF"},
		},
		{
			name:     "empty values are skipped",
			input:    map[string][]string{"mac": {"", "AA:BB:CC:DD:EE:FF", ""}},
			expected: []string{"mac:AA:BB:CC:DD:EE:FF"},
		},
		{
			name:     "all empty values",
			input:    map[string][]string{"mac": {"", ""}},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeutil.NamespacedKeys(tc.input)
			sort.Strings(got)
			sort.Strings(tc.expected)

			if len(got) != len(tc.expected) {
				t.Fatalf("got %d keys %v, want %d keys %v", len(got), got, len(tc.expected), tc.expected)
			}
			for i := range got {
				if got[i] != tc.expected[i] {
					t.Errorf("key[%d] = %q, want %q", i, got[i], tc.expected[i])
				}
			}
		})
	}
}

func TestSetIfEmpty(t *testing.T) {
	tests := []struct {
		name     string
		dst      string
		val      string
		expected string
	}{
		{
			name:     "empty target gets filled",
			dst:      "",
			val:      "hello",
			expected: "hello",
		},
		{
			name:     "non-empty target is preserved",
			dst:      "existing",
			val:      "new",
			expected: "existing",
		},
		{
			name:     "empty target with empty value stays empty",
			dst:      "",
			val:      "",
			expected: "",
		},
		{
			name:     "non-empty target with empty value is preserved",
			dst:      "existing",
			val:      "",
			expected: "existing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dst := tc.dst
			mergeutil.SetIfEmpty(&dst, tc.val)
			if dst != tc.expected {
				t.Errorf("got %q, want %q", dst, tc.expected)
			}
		})
	}
}

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colon-separated uppercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "colon-separated lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "dash-separated",
			input:    "AA-BB-CC-DD-EE-FF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "dot-separated Cisco style",
			input:    "AABB.CCDD.EEFF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "no separators",
			input:    "AABBCCDDEEFF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "mixed case no separators",
			input:    "aaBBccDDeeFF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "too short",
			input:    "AA:BB:CC",
			expected: "",
		},
		{
			name:     "too long",
			input:    "AA:BB:CC:DD:EE:FF:00",
			expected: "",
		},
		{
			name:     "all zeros",
			input:    "00:00:00:00:00:00",
			expected: "",
		},
		{
			name:     "all zeros no separators",
			input:    "000000000000",
			expected: "",
		},
		{
			name:     "leading and trailing whitespace",
			input:    "  AA:BB:CC:DD:EE:FF  ",
			expected: "AA:BB:CC:DD:EE:FF",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeMAC(tc.input)
			if got != tc.expected {
				t.Errorf("NormalizeMAC(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestInferEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		s1Site   string
		expected string
	}{
		{
			name:     "s1Site takes precedence",
			hostname: "prd-server-01",
			s1Site:   "SiteName",
			expected: "SiteName",
		},
		{
			name:     "prd prefix",
			hostname: "prd-server-01",
			s1Site:   "",
			expected: "PRODUCTION",
		},
		{
			name:     "prod prefix",
			hostname: "prod-server-01",
			s1Site:   "",
			expected: "PRODUCTION",
		},
		{
			name:     "PRD prefix uppercase",
			hostname: "PRD-server-01",
			s1Site:   "",
			expected: "PRODUCTION",
		},
		{
			name:     "uat prefix",
			hostname: "uat-server-01",
			s1Site:   "",
			expected: "UAT",
		},
		{
			name:     "drv prefix",
			hostname: "drv-server-01",
			s1Site:   "",
			expected: "UAT",
		},
		{
			name:     "dev prefix",
			hostname: "dev-server-01",
			s1Site:   "",
			expected: "UAT",
		},
		{
			name:     "unknown prefix",
			hostname: "web-server-01",
			s1Site:   "",
			expected: "",
		},
		{
			name:     "empty hostname",
			hostname: "",
			s1Site:   "",
			expected: "",
		},
		{
			name:     "prefix without dash is not matched",
			hostname: "production-server",
			s1Site:   "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InferEnvironment(tc.hostname, tc.s1Site)
			if got != tc.expected {
				t.Errorf("InferEnvironment(%q, %q) = %q, want %q", tc.hostname, tc.s1Site, got, tc.expected)
			}
		})
	}
}
