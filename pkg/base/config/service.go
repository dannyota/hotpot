package config

import (
	"context"
	"fmt"
	"log"
	"sync"
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
// Defaults to "localhost:7233" if not configured.
func (s *Service) TemporalHostPort() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil || s.config.Temporal.HostPort == "" {
		return "localhost:7233"
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
