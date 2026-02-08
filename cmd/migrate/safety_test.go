package main

import (
	"testing"
)

func TestSafetyCheck(t *testing.T) {
	tests := []struct {
		name      string
		targetDSN string
		devDSN    string
		wantPanic bool
	}{
		{
			name:      "different databases - safe",
			targetDSN: "host=localhost port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			devDSN:    "host=localhost port=5432 user=postgres password=secret dbname=hotpot_dev sslmode=disable",
			wantPanic: false,
		},
		{
			name:      "same database - UNSAFE",
			targetDSN: "host=localhost port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			devDSN:    "host=localhost port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			wantPanic: true,
		},
		{
			name:      "different hosts - safe",
			targetDSN: "host=prod.example.com port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			devDSN:    "host=localhost port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			wantPanic: false,
		},
		{
			name:      "different ports - safe",
			targetDSN: "host=localhost port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			devDSN:    "host=localhost port=5433 user=postgres password=secret dbname=hotpot sslmode=disable",
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetURL, err := dsnToURL(tt.targetDSN)
			if err != nil {
				t.Fatalf("Failed to convert target DSN: %v", err)
			}

			devURL, err := dsnToURL(tt.devDSN)
			if err != nil {
				t.Fatalf("Failed to convert dev DSN: %v", err)
			}

			// Check if URLs are the same (this is what the safety check does)
			same := (targetURL == devURL)

			if tt.wantPanic && !same {
				t.Errorf("Expected URLs to be the same (unsafe), but they differ")
			}
			if !tt.wantPanic && same {
				t.Errorf("Expected URLs to be different (safe), but they are the same")
			}
		})
	}
}
