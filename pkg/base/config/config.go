package config

// Config holds all application configuration values.
type Config struct {
	GCP      GCPConfig      `yaml:"gcp"`
	AWS      AWSConfig      `yaml:"aws"`
	S1       S1Config       `yaml:"s1"`
	DO        DOConfig        `yaml:"do"`
	GreenNode GreenNodeConfig `yaml:"greennode"`
	Database DatabaseConfig `yaml:"database"`
	Temporal TemporalConfig `yaml:"temporal"`
	Redis    RedisConfig    `yaml:"redis"`
}

// AWSConfig holds AWS-specific configuration.
type AWSConfig struct {
	// Enabled controls whether AWS ingestion runs and tables are created.
	Enabled bool `yaml:"enabled"`

	// AccessKeyID is the AWS access key ID for static credentials.
	// Falls back to default credential chain if empty.
	AccessKeyID string `yaml:"access_key_id,omitempty"`

	// SecretAccessKey is the AWS secret access key for static credentials.
	// Falls back to default credential chain if empty.
	SecretAccessKey string `yaml:"secret_access_key,omitempty"`

	// Regions is an optional filter for which AWS regions to scan.
	// If empty, all enabled regions are discovered via DescribeRegions.
	Regions []string `yaml:"regions,omitempty"`

	// RateLimitPerMinute is the max API requests per minute across all AWS clients.
	// Default: 600 (see Service.AWSRateLimitPerMinute()).
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`
}

// GCPConfig holds GCP-specific configuration.
type GCPConfig struct {
	// Enabled controls whether GCP ingestion runs and tables are created.
	Enabled bool `yaml:"enabled"`

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
	Enabled            bool   `yaml:"enabled"`
	BaseURL            string `yaml:"base_url"`
	APIToken           string `yaml:"api_token"`
	RateLimitPerMinute int    `yaml:"rate_limit_per_minute,omitempty"`
	BatchSize          int    `yaml:"batch_size,omitempty"`
}

// DOConfig holds DigitalOcean configuration.
type DOConfig struct {
	Enabled            bool   `yaml:"enabled"`
	APIToken           string `yaml:"api_token"`
	RateLimitPerMinute int    `yaml:"rate_limit_per_minute,omitempty"`
}

// GreenNodeConfig holds GreenNode (formerly VNG Cloud) configuration.
type GreenNodeConfig struct {
	// Enabled controls whether GreenNode ingestion runs and tables are created.
	Enabled bool `yaml:"enabled"`

	// Regions is the list of GreenNode regions to ingest (e.g., ["hcm-3", "han-1"]).
	Regions []string `yaml:"regions"`

	// ClientID is the service account client ID for OAuth2 authentication.
	ClientID string `yaml:"client_id,omitempty"`

	// ClientSecret is the service account client secret.
	ClientSecret string `yaml:"client_secret,omitempty"`

	// ProjectID is the GreenNode project to ingest resources from.
	ProjectID string `yaml:"project_id"`

	// RootEmail is the IAM root account email (for IAM user auth).
	RootEmail string `yaml:"root_email,omitempty"`

	// Username is the IAM username (for IAM user auth).
	Username string `yaml:"username,omitempty"`

	// Password is the IAM password (for IAM user auth).
	Password string `yaml:"password,omitempty"`

	// TOTPSecret is the base32-encoded TOTP secret for 2FA (optional, for IAM user auth).
	TOTPSecret string `yaml:"totp_secret,omitempty"`

	// RateLimitPerMinute is the max API requests per minute across all GreenNode clients.
	// Default: 300 (see Service.GreenNodeRateLimitPerMinute()).
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db,omitempty"`
}
