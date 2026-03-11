package config

// Config holds all application configuration values.
type Config struct {
	// LogLevel controls the minimum log level: debug, info, warn, error.
	// Default: "info" (see Service.LogLevel()).
	LogLevel string `yaml:"log_level,omitempty"`

	GCP      GCPConfig      `yaml:"gcp"`
	AWS      AWSConfig      `yaml:"aws"`
	S1       S1Config       `yaml:"s1"`
	DO        DOConfig        `yaml:"do"`
	GreenNode GreenNodeConfig `yaml:"greennode"`
	Vault     VaultConfig     `yaml:"vault"`
	Jenkins   JenkinsConfig   `yaml:"jenkins"`
	MEEC      MEECConfig      `yaml:"meec"`
	Reference  ReferenceConfig  `yaml:"reference"`
	ApiCatalog ApiCatalogConfig `yaml:"apicatalog"`
	AccessLog  AccessLogConfig  `yaml:"accesslog"`
	Admin      AdminConfig      `yaml:"admin"`
	Database   DatabaseConfig   `yaml:"database"`
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

	// QuotaProject overrides the project used for API billing and quota checks.
	// When set, all GCP API calls use this project for quota unless auto-detection
	// finds a better match (e.g., a different project has the needed API enabled).
	// If empty, auto-detected from discovery data; falls back to credentials project.
	QuotaProject string `yaml:"quota_project,omitempty"`

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

// VaultConfig holds HashiCorp Vault configuration for multi-instance PKI ingestion.
type VaultConfig struct {
	// Enabled controls whether Vault ingestion runs and tables are created.
	Enabled bool `yaml:"enabled"`

	// RateLimitPerMinute is the max API requests per minute across all Vault instances.
	// Default: 60 (see Service.VaultRateLimitPerMinute()).
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`

	// Instances is the list of Vault servers to ingest from.
	Instances []VaultInstance `yaml:"instances"`
}

// VaultInstance holds connection details for a single Vault server.
type VaultInstance struct {
	// Name is a unique identifier for this Vault instance (e.g., "prod-vault").
	Name string `yaml:"name"`

	// Address is the full URL of the Vault server (e.g., "https://vault.example.com").
	Address string `yaml:"address"`

	// Token is the Vault authentication token.
	Token string `yaml:"token"`

	// VerifySSL controls TLS certificate verification. Default: true.
	VerifySSL *bool `yaml:"verify_ssl,omitempty"`
}

// JenkinsConfig holds Jenkins CI configuration.
type JenkinsConfig struct {
	// Enabled controls whether Jenkins ingestion runs and tables are created.
	Enabled bool `yaml:"enabled"`

	// BaseURL is the Jenkins server URL (e.g., "https://jenkins.example.com").
	BaseURL string `yaml:"base_url"`

	// Username is the Jenkins username for API authentication.
	Username string `yaml:"username"`

	// APIToken is the Jenkins API token for authentication.
	APIToken string `yaml:"api_token"`

	// VerifySSL controls TLS certificate verification. Default: true.
	VerifySSL *bool `yaml:"verify_ssl,omitempty"`

	// Timeout is the HTTP request timeout in seconds. Default: 30.
	Timeout int `yaml:"timeout,omitempty"`

	// Since filters which jobs to process. Only jobs with lastBuild.timestamp >= since
	// are included. Format: "2024-01-01".
	Since string `yaml:"since"`

	// MaxBuildsPerJob limits how many builds to pull per job per run. Default: 1000.
	MaxBuildsPerJob int `yaml:"max_builds_per_job,omitempty"`

	// ExcludeRepos is a list of repo URL patterns to exclude from build repos.
	ExcludeRepos []string `yaml:"exclude_repos,omitempty"`

	// RateLimitPerMinute is the max API requests per minute. Default: 120.
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`
}

// MEECConfig holds ManageEngine Endpoint Central configuration.
type MEECConfig struct {
	// Enabled controls whether MEEC ingestion runs and tables are created.
	Enabled bool `yaml:"enabled"`

	// BaseURL is the MEEC server URL (e.g., "https://10.91.9.133:8383").
	BaseURL string `yaml:"base_url"`

	// Username is the MEEC login username.
	Username string `yaml:"username"`

	// Password is the MEEC login password (plain text, base64-encoded before sending).
	Password string `yaml:"password"`

	// AuthType is the MEEC authentication type. Default: "local_authentication".
	AuthType string `yaml:"auth_type,omitempty"`

	// TOTPSecret is the base32-encoded TOTP secret for 2FA (optional).
	TOTPSecret string `yaml:"totp_secret,omitempty"`

	// APIVersion is the MEEC API version. Default: "1.4".
	APIVersion string `yaml:"api_version,omitempty"`

	// VerifySSL controls TLS certificate verification. Default: true.
	VerifySSL *bool `yaml:"verify_ssl,omitempty"`

	// RateLimitPerMinute is the max API requests per minute. Default: 120.
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`
}

// ReferenceConfig holds configuration for public reference data ingestion (NVD CPE, Ubuntu, RHEL).
type ReferenceConfig struct {
	// Enabled controls whether reference data ingestion runs.
	Enabled bool `yaml:"enabled"`

	// RateLimitPerMinute is the max HTTP requests per minute to public servers.
	// Default: 30 (see Service.ReferenceRateLimitPerMinute()).
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`
}

// ApiCatalogConfig holds API catalog ingestion configuration.
type ApiCatalogConfig struct {
	// Enabled controls whether API catalog ingestion runs.
	Enabled bool `yaml:"enabled"`
}

// AccessLogConfig holds access log monitoring configuration.
type AccessLogConfig struct {
	// Enabled controls whether access log monitoring runs.
	Enabled bool `yaml:"enabled"`

	// RateLimitPerMinute is the max API requests per minute for access log sources.
	// GCP Cloud Logging quota: 60 read requests/min/project.
	// Default: 60 (see Service.AccessLogRateLimitPerMinute()).
	RateLimitPerMinute int `yaml:"rate_limit_per_minute,omitempty"`

	// RetentionDays is how many days to keep anomalies and silver traffic data.
	// CleanupStale deletes data older than this.
	// Default: 90 (see Service.AccessLogRetentionDays()).
	RetentionDays int `yaml:"retention_days,omitempty"`

	// BackfillDays is how far back to ingest on first run (no cursor).
	// 0 means no backfill (starts 1 hour ago, same as default).
	// Capped to RetentionDays to avoid ingesting data that gets immediately cleaned up.
	// Default: 0 (see Service.AccessLogBackfillDays()).
	BackfillDays int `yaml:"backfill_days,omitempty"`

	// BackfillIntervalMinutes is the window size during backfill.
	// Larger = fewer API calls, coarser granularity. Only applies on first run.
	// Default: 60 (see Service.AccessLogBackfillIntervalMinutes()).
	BackfillIntervalMinutes int `yaml:"backfill_interval_minutes,omitempty"`

	// GeoIPCityPath is the path to the city-level GeoIP .mmdb file.
	// Default: "data/geoip/dbip-city.mmdb" relative to the binary.
	// The directory is auto-created on download if it doesn't exist.
	GeoIPCityPath string `yaml:"geoip_city_path,omitempty"`

	// GeoIPASNPath is the path to the ASN-level GeoIP .mmdb file (IPinfo format).
	// Default: "data/geoip/ipinfo-asn.mmdb" relative to the binary.
	// The directory is auto-created on download if it doesn't exist.
	GeoIPASNPath string `yaml:"geoip_asn_path,omitempty"`

	// IPInfoToken is the IPinfo API token for downloading the free ASN .mmdb file.
	// Required for automatic ASN database updates.
	IPInfoToken string `yaml:"ipinfo_token,omitempty"`

	// Sources is the list of access log sources to ingest from.
	Sources []AccessLogSourceConfig `yaml:"sources,omitempty"`
}

// AccessLogSourceConfig defines a single access log source.
type AccessLogSourceConfig struct {
	// Type is the source type (e.g. "gcplogging").
	Type string `yaml:"type"`

	// Name is a unique identifier for this source.
	Name string `yaml:"name"`

	// Role determines how this source contributes to the pipeline.
	// "primary" — owns request counts, client IPs, user agents.
	// "enrichment" — adds fields only, no duplicate counting.
	Role string `yaml:"role"`

	// ProjectID is the GCP project ID (for gcplogging sources).
	ProjectID string `yaml:"project_id,omitempty"`

	// BigQueryTable is the Log Analytics linked dataset table.
	// Format: "project.dataset._AllLogs"
	BigQueryTable string `yaml:"bigquery_table,omitempty"`

	// BQFilter is a BigQuery SQL WHERE clause fragment for filtering.
	// Example: "resource.type = 'gce_instance' AND REGEXP_CONTAINS(...)"
	BQFilter string `yaml:"bq_filter,omitempty"`

	// FieldMapping maps log JSON keys to standard names.
	FieldMapping map[string]string `yaml:"field_mapping,omitempty"`

	// IntervalMinutes is the collection interval in minutes.
	// Default: 5.
	IntervalMinutes int `yaml:"interval_minutes,omitempty"`

	// CredentialsJSON is the raw JSON bytes of the service account.
	// Optional — falls back to Application Default Credentials (ADC) if empty.
	CredentialsJSON []byte `yaml:"credentials_json,omitempty"`
}

// AdminConfig holds admin web UI configuration.
type AdminConfig struct {
	// Addr is the listen address for the admin HTTP server.
	// Default: ":8000" (see Service.AdminAddr()).
	Addr string `yaml:"addr,omitempty"`

	// UI holds the frontend UI configuration (project name, sidebar nav, etc.).
	UI AdminUIConfig `yaml:"ui"`
}

// AdminUIConfig defines the user-customizable frontend layout.
type AdminUIConfig struct {
	// Name is the project name shown in the sidebar header.
	// Default: "Hotpot".
	Name string `yaml:"name,omitempty"`

	// Description is a short subtitle shown under the name in the sidebar.
	// Example: "Unified security data platform".
	Description string `yaml:"description,omitempty"`

	// Title is the browser page title.
	// Default: "{name} - {description}" if description is set, otherwise "{name}".
	Title string `yaml:"title,omitempty"`

	// Icon is a single character, emoji, or image path for the sidebar logo.
	// Supports: "H", an emoji, or a path to an image in admin/ui/public/
	// (e.g., "/logo.png" for admin/ui/public/logo.png).
	// Images in public/ are embedded in the production binary.
	// Default: "🍲".
	Icon string `yaml:"icon,omitempty"`

	// Color is the CSS color for the sidebar logo background.
	// Default: "" (uses the default gradient).
	Color string `yaml:"color,omitempty"`

	// Disable is a list of API path prefixes to exclude from the admin UI.
	// Matching routes are not served and do not appear in the sidebar nav.
	// Trailing "/" matches all routes under that prefix; exact paths match one route.
	// Example: ["/api/v1/bronze/gcp/dns/", "/api/v1/gold/lifecycle/software"]
	Disable []string `yaml:"disable,omitempty"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db,omitempty"`
}
