package config

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// Service manages configuration lifecycle with hot reload support.
type Service struct {
	source      ConfigSource
	enableWatch bool

	config *Config
	mu     sync.RWMutex

	onReload  []func(*Config)
	stopWatch func()
}

// ServiceOptions configures the Service.
type ServiceOptions struct {
	// Source is the config backend (Vault or File).
	Source ConfigSource

	// EnableWatch enables config hot-reload.
	EnableWatch bool
}

// NewService creates a new config service.
func NewService(opts ServiceOptions) *Service {
	return &Service{
		source:      opts.Source,
		enableWatch: opts.EnableWatch,
	}
}

// Start loads initial config and starts watching for changes.
func (s *Service) Start(ctx context.Context) error {
	// Load initial config
	if err := s.reload(ctx); err != nil {
		return fmt.Errorf("load initial config: %w", err)
	}

	log.Printf("Config loaded from %s source", s.source.Type())

	// Start watching if enabled
	if s.enableWatch {
		stop, err := s.source.Watch(ctx, func() {
			if err := s.reload(ctx); err != nil {
				log.Printf("Config reload failed: %v", err)
				return
			}
			log.Printf("Config reloaded from %s source", s.source.Type())
		})
		if err != nil {
			return fmt.Errorf("start config watch: %w", err)
		}
		s.stopWatch = stop
	}

	return nil
}

// Stop stops watching and releases resources.
func (s *Service) Stop() error {
	if s.stopWatch != nil {
		s.stopWatch()
		s.stopWatch = nil
	}
	return nil
}

// OnReload registers a callback invoked when config changes.
// Callback receives the new config after successful reload.
func (s *Service) OnReload(fn func(*Config)) {
	s.onReload = append(s.onReload, fn)
}

// reload reloads config from source.
func (s *Service) reload(ctx context.Context) error {
	newConfig, err := s.source.Load(ctx)
	if err != nil {
		return err
	}

	// Validate required fields
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	s.mu.Lock()
	s.config = newConfig
	s.mu.Unlock()

	// Notify listeners (outside lock to prevent deadlock)
	for _, fn := range s.onReload {
		fn(newConfig)
	}

	return nil
}

// Config returns a copy of current config (thread-safe).
func (s *Service) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return Config{}
	}
	return *s.config
}

// LogLevel returns the configured slog.Level.
// Defaults to slog.LevelInfo if not configured or invalid.
func (s *Service) LogLevel() slog.Level {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.LogLevel == "" {
		return slog.LevelInfo
	}
	switch strings.ToLower(s.config.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GCPRateLimitPerMinute returns the max API requests per minute for GCP.
// Defaults to 600 if not configured.
func (s *Service) GCPRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.GCP.RateLimitPerMinute <= 0 {
		return 600
	}
	return s.config.GCP.RateLimitPerMinute
}

// GCPCredentialsJSON returns credentials JSON for GCP client options.
// Returns nil if not configured (caller should fall back to ADC).
func (s *Service) GCPCredentialsJSON() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || len(s.config.GCP.CredentialsJSON) == 0 {
		return nil
	}
	// Return a copy to prevent mutation
	result := make([]byte, len(s.config.GCP.CredentialsJSON))
	copy(result, s.config.GCP.CredentialsJSON)
	return result
}

// GCPEnabled returns true if GCP ingestion is enabled in config.
func (s *Service) GCPEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.GCP.Enabled
}

// EnabledProviders returns the list of provider names that are enabled in config.
func (s *Service) EnabledProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return nil
	}
	var providers []string
	if s.config.GCP.Enabled {
		providers = append(providers, "gcp")
	}
	if s.config.S1.Enabled {
		providers = append(providers, "s1")
	}
	if s.config.DO.Enabled {
		providers = append(providers, "do")
	}
	if s.config.AWS.Enabled {
		providers = append(providers, "aws")
	}
	if s.config.GreenNode.Enabled {
		providers = append(providers, "greennode")
	}
	if s.config.Vault.Enabled {
		providers = append(providers, "vault")
	}
	if s.config.Jenkins.Enabled {
		providers = append(providers, "jenkins")
	}
	if s.config.MEEC.Enabled {
		providers = append(providers, "meec")
	}
	if s.config.Reference.Enabled {
		providers = append(providers, "reference")
	}
	return providers
}

// DatabaseDSN returns the database connection string.
func (s *Service) DatabaseDSN() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	db := s.config.Database
	sslmode := db.SSLMode
	if sslmode == "" {
		sslmode = "require"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.DBName, sslmode)
}

// TemporalHostPort returns the Temporal server address.
// Panics if config is not loaded (should never happen after Start()).
func (s *Service) TemporalHostPort() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		panic("config not loaded")
	}
	return s.config.Temporal.HostPort
}

// TemporalNamespace returns the Temporal namespace.
// Defaults to "default" if not configured.
func (s *Service) TemporalNamespace() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Temporal.Namespace == "" {
		return "default"
	}
	return s.config.Temporal.Namespace
}

// S1BaseURL returns the SentinelOne management console base URL.
func (s *Service) S1BaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.S1.BaseURL
}

// S1APIToken returns the SentinelOne API token.
func (s *Service) S1APIToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.S1.APIToken
}

// S1RateLimitPerMinute returns the max API requests per minute for SentinelOne.
// Defaults to 180 if not configured (S1 has undocumented nginx rate limits).
func (s *Service) S1RateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.S1.RateLimitPerMinute <= 0 {
		return 180
	}
	return s.config.S1.RateLimitPerMinute
}

// S1BatchSize returns the batch size for SentinelOne API pagination.
// Defaults to 1000 if not configured. Capped at 1000 (API max for most endpoints).
func (s *Service) S1BatchSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.S1.BatchSize <= 0 {
		return 1000
	}
	if s.config.S1.BatchSize > 1000 {
		return 1000
	}
	return s.config.S1.BatchSize
}

// S1Enabled returns true if SentinelOne ingestion is enabled in config.
func (s *Service) S1Enabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.S1.Enabled
}

// DOAPIToken returns the DigitalOcean API token.
func (s *Service) DOAPIToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.DO.APIToken
}

// DORateLimitPerMinute returns the max API requests per minute for DigitalOcean.
// Defaults to 300 if not configured.
func (s *Service) DORateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.DO.RateLimitPerMinute <= 0 {
		return 300
	}
	return s.config.DO.RateLimitPerMinute
}

// DOEnabled returns true if DigitalOcean ingestion is enabled in config.
func (s *Service) DOEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.DO.Enabled
}

// AWSEnabled returns true if AWS ingestion is enabled in config.
func (s *Service) AWSEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.AWS.Enabled
}

// AWSAccessKeyID returns the AWS access key ID.
// Returns empty string if not configured (caller should fall back to default credential chain).
func (s *Service) AWSAccessKeyID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.AWS.AccessKeyID
}

// AWSSecretAccessKey returns the AWS secret access key.
// Returns empty string if not configured (caller should fall back to default credential chain).
func (s *Service) AWSSecretAccessKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.AWS.SecretAccessKey
}

// AWSRegions returns the optional region filter for AWS.
// Returns nil if not configured (caller should discover all enabled regions).
func (s *Service) AWSRegions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || len(s.config.AWS.Regions) == 0 {
		return nil
	}
	result := make([]string, len(s.config.AWS.Regions))
	copy(result, s.config.AWS.Regions)
	return result
}

// AWSRateLimitPerMinute returns the max API requests per minute for AWS.
// Defaults to 600 if not configured.
func (s *Service) AWSRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.AWS.RateLimitPerMinute <= 0 {
		return 600
	}
	return s.config.AWS.RateLimitPerMinute
}

// GreenNodeEnabled returns true if GreenNode ingestion is enabled in config.
func (s *Service) GreenNodeEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.GreenNode.Enabled
}

// GreenNodeRegions returns the configured GreenNode regions (e.g., ["hcm-3", "han-1"]).
func (s *Service) GreenNodeRegions() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || len(s.config.GreenNode.Regions) == 0 {
		return nil
	}
	result := make([]string, len(s.config.GreenNode.Regions))
	copy(result, s.config.GreenNode.Regions)
	return result
}

// GreenNodeClientID returns the GreenNode service account client ID.
func (s *Service) GreenNodeClientID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.ClientID
}

// GreenNodeClientSecret returns the GreenNode service account client secret.
func (s *Service) GreenNodeClientSecret() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.ClientSecret
}

// GreenNodeProjectID returns the GreenNode project ID.
func (s *Service) GreenNodeProjectID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.ProjectID
}

// GreenNodeRootEmail returns the GreenNode IAM root account email.
func (s *Service) GreenNodeRootEmail() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.RootEmail
}

// GreenNodeUsername returns the GreenNode IAM username.
func (s *Service) GreenNodeUsername() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.Username
}

// GreenNodePassword returns the GreenNode IAM password.
func (s *Service) GreenNodePassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.Password
}

// GreenNodeTOTPSecret returns the GreenNode TOTP secret for 2FA.
func (s *Service) GreenNodeTOTPSecret() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.GreenNode.TOTPSecret
}

// GreenNodeRateLimitPerMinute returns the max API requests per minute for GreenNode.
// Defaults to 300 if not configured.
func (s *Service) GreenNodeRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.GreenNode.RateLimitPerMinute <= 0 {
		return 300
	}
	return s.config.GreenNode.RateLimitPerMinute
}

// VaultEnabled returns true if Vault ingestion is enabled in config.
func (s *Service) VaultEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.Vault.Enabled
}

// VaultRateLimitPerMinute returns the max API requests per minute for Vault.
// Defaults to 60 if not configured.
func (s *Service) VaultRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Vault.RateLimitPerMinute <= 0 {
		return 60
	}
	return s.config.Vault.RateLimitPerMinute
}

// VaultInstances returns a copy of configured Vault instances.
func (s *Service) VaultInstances() []VaultInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || len(s.config.Vault.Instances) == 0 {
		return nil
	}
	result := make([]VaultInstance, len(s.config.Vault.Instances))
	copy(result, s.config.Vault.Instances)
	return result
}

// VaultInstance looks up a Vault instance by name.
// Returns nil if not found.
func (s *Service) VaultInstance(name string) *VaultInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return nil
	}
	for _, inst := range s.config.Vault.Instances {
		if inst.Name == name {
			v := inst
			return &v
		}
	}
	return nil
}

// JenkinsEnabled returns true if Jenkins ingestion is enabled in config.
func (s *Service) JenkinsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.Jenkins.Enabled
}

// JenkinsBaseURL returns the Jenkins server base URL.
func (s *Service) JenkinsBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.Jenkins.BaseURL
}

// JenkinsUsername returns the Jenkins username.
func (s *Service) JenkinsUsername() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.Jenkins.Username
}

// JenkinsAPIToken returns the Jenkins API token.
func (s *Service) JenkinsAPIToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.Jenkins.APIToken
}

// JenkinsVerifySSL returns whether to verify SSL certificates for Jenkins.
// Defaults to true if not configured.
func (s *Service) JenkinsVerifySSL() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Jenkins.VerifySSL == nil {
		return true
	}
	return *s.config.Jenkins.VerifySSL
}

// JenkinsTimeout returns the HTTP request timeout in seconds for Jenkins.
// Defaults to 30 if not configured.
func (s *Service) JenkinsTimeout() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Jenkins.Timeout <= 0 {
		return 30
	}
	return s.config.Jenkins.Timeout
}

// JenkinsSince returns the since date for filtering Jenkins jobs.
// Returns zero time if not configured or invalid.
func (s *Service) JenkinsSince() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Jenkins.Since == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s.config.Jenkins.Since)
	if err != nil {
		return time.Time{}
	}
	return t
}

// JenkinsMaxBuildsPerJob returns the max builds to pull per job per run.
// Defaults to 1000 if not configured.
func (s *Service) JenkinsMaxBuildsPerJob() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Jenkins.MaxBuildsPerJob <= 0 {
		return 1000
	}
	return s.config.Jenkins.MaxBuildsPerJob
}

// JenkinsExcludeRepos returns repo URL patterns to exclude.
func (s *Service) JenkinsExcludeRepos() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || len(s.config.Jenkins.ExcludeRepos) == 0 {
		return nil
	}
	result := make([]string, len(s.config.Jenkins.ExcludeRepos))
	copy(result, s.config.Jenkins.ExcludeRepos)
	return result
}

// JenkinsRateLimitPerMinute returns the max API requests per minute for Jenkins.
// Defaults to 120 if not configured.
func (s *Service) JenkinsRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Jenkins.RateLimitPerMinute <= 0 {
		return 120
	}
	return s.config.Jenkins.RateLimitPerMinute
}

// MEECEnabled returns true if MEEC ingestion is enabled in config.
func (s *Service) MEECEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.MEEC.Enabled
}

// MEECBaseURL returns the MEEC server base URL.
func (s *Service) MEECBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.MEEC.BaseURL
}

// MEECUsername returns the MEEC login username.
func (s *Service) MEECUsername() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.MEEC.Username
}

// MEECPassword returns the MEEC login password.
func (s *Service) MEECPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.MEEC.Password
}

// MEECAuthType returns the MEEC authentication type.
// Defaults to "local_authentication" if not configured.
func (s *Service) MEECAuthType() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.MEEC.AuthType == "" {
		return "local_authentication"
	}
	return s.config.MEEC.AuthType
}

// MEECTOTPSecret returns the MEEC TOTP secret for 2FA.
func (s *Service) MEECTOTPSecret() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return ""
	}
	return s.config.MEEC.TOTPSecret
}

// MEECAPIVersion returns the MEEC API version.
// Defaults to "1.4" if not configured.
func (s *Service) MEECAPIVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.MEEC.APIVersion == "" {
		return "1.4"
	}
	return s.config.MEEC.APIVersion
}

// MEECVerifySSL returns whether to verify SSL certificates for MEEC.
// Defaults to true if not configured.
func (s *Service) MEECVerifySSL() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.MEEC.VerifySSL == nil {
		return true
	}
	return *s.config.MEEC.VerifySSL
}

// MEECRateLimitPerMinute returns the max API requests per minute for MEEC.
// Defaults to 120 if not configured.
func (s *Service) MEECRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.MEEC.RateLimitPerMinute <= 0 {
		return 120
	}
	return s.config.MEEC.RateLimitPerMinute
}

// ReferenceEnabled returns true if reference data ingestion is enabled in config.
func (s *Service) ReferenceEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.Reference.Enabled
}

// ReferenceRateLimitPerMinute returns the max HTTP requests per minute for reference data.
// Defaults to 30 if not configured (gentle on public servers).
func (s *Service) ReferenceRateLimitPerMinute() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Reference.RateLimitPerMinute <= 0 {
		return 30
	}
	return s.config.Reference.RateLimitPerMinute
}

// RedisConfig returns the Redis configuration.
// Returns nil if not configured.
func (s *Service) RedisConfig() *RedisConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Redis.Address == "" {
		return nil
	}
	cfg := s.config.Redis
	return &cfg
}
