package main

import (
	"strings"
	"testing"
)

func TestSafetyCheck(t *testing.T) {
	tests := []struct {
		name     string
		dbName   string
		wantSafe bool
	}{
		{
			name:     "ends with _dev - safe",
			dbName:   "hotpot_dev",
			wantSafe: true,
		},
		{
			name:     "production name - UNSAFE",
			dbName:   "hotpot",
			wantSafe: false,
		},
		{
			name:     "dev in middle - UNSAFE",
			dbName:   "hotpot_dev_old",
			wantSafe: false,
		},
		{
			name:     "just _dev - safe",
			dbName:   "_dev",
			wantSafe: true,
		},
		{
			name:     "empty - UNSAFE",
			dbName:   "",
			wantSafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe := strings.HasSuffix(tt.dbName, "_dev")

			if tt.wantSafe && !safe {
				t.Errorf("Expected %q to be safe (ends with _dev), but it was not", tt.dbName)
			}
			if !tt.wantSafe && safe {
				t.Errorf("Expected %q to be unsafe, but it passed the check", tt.dbName)
			}
		})
	}
}
