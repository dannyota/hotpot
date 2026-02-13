package config

import "testing"

func TestGCPEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{"nil config", nil, false},
		{"disabled", &Config{}, false},
		{"enabled", &Config{GCP: GCPConfig{Enabled: true}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{config: tt.config}
			if got := s.GCPEnabled(); got != tt.want {
				t.Errorf("GCPEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS1Enabled(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{"nil config", nil, false},
		{"disabled", &Config{}, false},
		{"enabled", &Config{S1: S1Config{Enabled: true}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{config: tt.config}
			if got := s.S1Enabled(); got != tt.want {
				t.Errorf("S1Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDOEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{"nil config", nil, false},
		{"disabled", &Config{}, false},
		{"enabled", &Config{DO: DOConfig{Enabled: true}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{config: tt.config}
			if got := s.DOEnabled(); got != tt.want {
				t.Errorf("DOEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnabledProviders(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   []string
	}{
		{"nil config", nil, nil},
		{"none enabled", &Config{}, nil},
		{"gcp only", &Config{GCP: GCPConfig{Enabled: true}}, []string{"gcp"}},
		{"all enabled", &Config{
			GCP: GCPConfig{Enabled: true},
			S1:  S1Config{Enabled: true},
			DO:  DOConfig{Enabled: true},
		}, []string{"gcp", "s1", "do"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{config: tt.config}
			got := s.EnabledProviders()
			if len(got) != len(tt.want) {
				t.Fatalf("EnabledProviders() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("EnabledProviders()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
