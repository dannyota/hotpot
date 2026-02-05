package config

// Config holds all application configuration values.
type Config struct {
	GCP      GCPConfig      `json:"gcp"`
	Database DatabaseConfig `json:"database"`
	Temporal TemporalConfig `json:"temporal"`
}

// GCPConfig holds GCP-specific configuration.
type GCPConfig struct {
	// CredentialsJSON is the raw JSON bytes of the service account.
	// Loaded from Vault secret. Preferred over CredentialsFile.
	CredentialsJSON []byte `json:"-"`

	// CredentialsFile is the path to credentials file (fallback).
	// Used only when CredentialsJSON is empty.
	CredentialsFile string `json:"credentials_file,omitempty"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// TemporalConfig holds Temporal connection configuration.
type TemporalConfig struct {
	HostPort  string `json:"host_port"`
	Namespace string `json:"namespace"`
}
