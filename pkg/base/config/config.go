package config

// Config holds all application configuration values.
type Config struct {
	GCP      GCPConfig      `yaml:"gcp"`
	S1       S1Config       `yaml:"s1"`
	DO       DOConfig       `yaml:"do"`
	Database DatabaseConfig `yaml:"database"`
	Temporal TemporalConfig `yaml:"temporal"`
	Redis    RedisConfig    `yaml:"redis"`
}

// GCPConfig holds GCP-specific configuration.
type GCPConfig struct {
	// CredentialsJSON is the raw JSON bytes of the service account.
	// Loaded from Vault or YAML config. Falls back to ADC if empty.
	CredentialsJSON []byte `yaml:"credentials_json,omitempty"`

	// RateLimitPerMinute is the max API requests per minute across all GCP clients.
	// Default: 600 (see Service.GCPRateLimitPerMinute()).
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`
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
	HostPort  string `yaml:"host_port"`
	Namespace string `yaml:"namespace,omitempty"`
}

// S1Config holds SentinelOne configuration.
type S1Config struct {
	BaseURL            string `yaml:"base_url"`
	APIToken           string `yaml:"api_token"`
	RateLimitPerMinute int    `yaml:"rate_limit_per_minute,omitempty"`
	BatchSize          int    `yaml:"batch_size,omitempty"`
}

// DOConfig holds DigitalOcean configuration.
type DOConfig struct {
	APIToken           string `yaml:"api_token"`
	RateLimitPerMinute int    `yaml:"rate_limit_per_minute,omitempty"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db,omitempty"`
}
