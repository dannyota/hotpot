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
	// Loaded from Vault or YAML config. Falls back to ADC if empty.
	CredentialsJSON []byte `yaml:"credentials_json,omitempty"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode,omitempty"`
}

// TemporalConfig holds Temporal connection configuration.
type TemporalConfig struct {
	HostPort  string `yaml:"host_port,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}
