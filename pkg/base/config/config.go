package config

// Config holds all application configuration values.
type Config struct {
	GCP      GCPConfig      `yaml:"gcp"`
	Database DatabaseConfig `yaml:"database"`
	Temporal TemporalConfig `yaml:"temporal"`
}

// GCPConfig holds GCP-specific configuration.
type GCPConfig struct {
	// CredentialsJSON is the raw JSON bytes of the service account.
	// Loaded from Vault secret. Preferred over CredentialsFile.
	CredentialsJSON []byte `yaml:"-"`

	// CredentialsFile is the path to credentials file (fallback).
	// Used only when CredentialsJSON is empty.
	CredentialsFile string `yaml:"credentials_file,omitempty"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// TemporalConfig holds Temporal connection configuration.
type TemporalConfig struct {
	HostPort  string `yaml:"host_port"`
	Namespace string `yaml:"namespace"`
}
