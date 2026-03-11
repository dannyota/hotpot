package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name: "valid config",
			config: Config{
				Temporal: TemporalConfig{
					HostPort:  "localhost:7233",
					Namespace: "default",
				},
				Database: DatabaseConfig{
					Host:   "localhost",
					Port:   5432,
					User:   "user",
					DBName: "hotpot",
				},
			},
			wantErr: "",
		},
		{
			name: "missing temporal host_port",
			config: Config{
				Database: DatabaseConfig{
					Host:   "localhost",
					Port:   5432,
					User:   "user",
					DBName: "hotpot",
				},
			},
			wantErr: "temporal.host_port is required",
		},
		{
			name: "missing database host",
			config: Config{
				Temporal: TemporalConfig{
					HostPort: "localhost:7233",
				},
				Database: DatabaseConfig{
					Port:   5432,
					User:   "user",
					DBName: "hotpot",
				},
			},
			wantErr: "database.host is required",
		},
		{
			name: "missing database port",
			config: Config{
				Temporal: TemporalConfig{
					HostPort: "localhost:7233",
				},
				Database: DatabaseConfig{
					Host:   "localhost",
					User:   "user",
					DBName: "hotpot",
				},
			},
			wantErr: "database.port is required",
		},
		{
			name: "missing database user",
			config: Config{
				Temporal: TemporalConfig{
					HostPort: "localhost:7233",
				},
				Database: DatabaseConfig{
					Host:   "localhost",
					Port:   5432,
					DBName: "hotpot",
				},
			},
			wantErr: "database.user is required",
		},
		{
			name: "missing database dbname",
			config: Config{
				Temporal: TemporalConfig{
					HostPort: "localhost:7233",
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					User: "user",
				},
			},
			wantErr: "database.dbname is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}
