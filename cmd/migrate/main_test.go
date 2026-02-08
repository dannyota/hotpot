package main

import "testing"

func TestDsnToURL(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		want    string
		wantErr bool
	}{
		{
			name: "full dsn with password",
			dsn:  "host=localhost port=5432 user=postgres password=secret dbname=hotpot sslmode=disable",
			want: "postgres://postgres:secret@localhost:5432/hotpot?sslmode=disable",
		},
		{
			name: "dsn without password",
			dsn:  "host=localhost port=5432 user=postgres dbname=hotpot sslmode=require",
			want: "postgres://postgres@localhost:5432/hotpot?sslmode=require",
		},
		{
			name: "dsn without sslmode",
			dsn:  "host=db.example.com port=5433 user=admin password=pass123 dbname=prod",
			want: "postgres://admin:pass123@db.example.com:5433/prod",
		},
		{
			name:    "missing host",
			dsn:     "port=5432 user=postgres dbname=hotpot",
			wantErr: true,
		},
		{
			name:    "missing port",
			dsn:     "host=localhost user=postgres dbname=hotpot",
			wantErr: true,
		},
		{
			name:    "missing user",
			dsn:     "host=localhost port=5432 dbname=hotpot",
			wantErr: true,
		},
		{
			name:    "missing dbname",
			dsn:     "host=localhost port=5432 user=postgres",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dsnToURL(tt.dsn)
			if tt.wantErr {
				if err == nil {
					t.Errorf("dsnToURL() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("dsnToURL() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("dsnToURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
