package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"hotpot/pkg/base/config"
	"hotpot/pkg/storage/ent"
)

// App provides a unified interface for config and database with hot-reload.
// It manages config.Service and ent.Client lifecycle, including automatic
// reconnection when database configuration changes.
type App struct {
	configService *config.Service
	dbManager     *dbManager

	// Context management
	ctx    context.Context
	cancel context.CancelFunc

	// Runner management
	wg     sync.WaitGroup
	errMu  sync.Mutex
	errors []error
}

// New creates a new App.
// If ConfigSource is not provided in options, it auto-detects from env vars.
func New(opts Options) (*App, error) {
	// Auto-detect config source if not provided
	source := opts.ConfigSource
	if source == nil {
		source = detectConfigSource()
	}

	// Create config service
	configService := config.NewService(config.ServiceOptions{
		Source:      source,
		EnableWatch: true,
	})

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Create app
	app := &App{
		configService: configService,
		dbManager:     newDBManager(configService, opts.GracePeriod, opts.OnDBReconnect),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Setup signal handler
	go app.handleSignals()

	return app, nil
}

// handleSignals listens for SIGINT/SIGTERM and cancels the context.
func (a *App) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		a.cancel()
	case <-a.ctx.Done():
		// Context was cancelled elsewhere
	}

	signal.Stop(sigCh)
}

// Start initializes the app: starts config watching and connects to database.
func (a *App) Start(ctx context.Context) error {
	// Start config service first
	if err := a.configService.Start(ctx); err != nil {
		return fmt.Errorf("start config service: %w", err)
	}

	// Connect to database using loaded config
	if err := a.dbManager.connect(); err != nil {
		a.configService.Stop() // Cleanup on failure
		return fmt.Errorf("connect to database: %w", err)
	}

	// Register for config reload notifications
	a.configService.OnReload(func(cfg *config.Config) {
		a.dbManager.reconnectIfChanged()
	})

	return nil
}

// Stop gracefully shuts down the app.
func (a *App) Stop() error {
	// Cancel context to signal all runners to stop
	a.cancel()

	// Stop config watching
	if err := a.configService.Stop(); err != nil {
		return fmt.Errorf("stop config service: %w", err)
	}

	// Close database connection
	if err := a.dbManager.close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}

// Context returns the app's context that will be cancelled on shutdown.
func (a *App) Context() context.Context {
	return a.ctx
}

// ConfigService returns the underlying config service.
func (a *App) ConfigService() *config.Service {
	return a.configService
}

// EntClient returns the current Ent client.
// The client may change after a config reload.
func (a *App) EntClient() *ent.Client {
	return a.dbManager.EntClient()
}

// Config returns a copy of current configuration.
func (a *App) Config() config.Config {
	return a.configService.Config()
}

// RunFunc is the signature for service runner functions.
// The context is cancelled when the app receives SIGINT/SIGTERM.
type RunFunc func(ctx context.Context, configService *config.Service, entClient *ent.Client) error

// Run starts a service runner in a goroutine (non-blocking).
// Use Wait() to block until all runners complete.
// Multiple runners can be started concurrently.
// The runner receives a context that is cancelled on SIGINT/SIGTERM.
func (a *App) Run(runner RunFunc) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := runner(a.ctx, a.configService, a.dbManager.EntClient()); err != nil {
			a.errMu.Lock()
			a.errors = append(a.errors, err)
			a.errMu.Unlock()
		}
	}()
}

// Wait blocks until all runners complete and returns the first error if any.
func (a *App) Wait() error {
	a.wg.Wait()

	a.errMu.Lock()
	defer a.errMu.Unlock()

	if len(a.errors) > 0 {
		return a.errors[0]
	}
	return nil
}
