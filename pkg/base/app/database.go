package app

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// dbManager handles database connections with hot-reload support.
type dbManager struct {
	configService *config.Service
	gracePeriod   time.Duration
	onReconnect   func(oldDSN, newDSN string)

	entClient  *ent.Client
	currentDSN string
	mu         sync.RWMutex
}

// newDBManager creates a new database manager.
func newDBManager(configService *config.Service, gracePeriod time.Duration, onReconnect func(string, string)) *dbManager {
	if gracePeriod == 0 {
		gracePeriod = DefaultGracePeriod
	}
	return &dbManager{
		configService: configService,
		gracePeriod:   gracePeriod,
		onReconnect:   onReconnect,
	}
}

// connect establishes the initial database connection.
func (m *dbManager) connect() error {
	dsn := m.configService.DatabaseDSN()
	if dsn == "" {
		return fmt.Errorf("database DSN is empty")
	}

	// Open database connection
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Create Ent driver
	drv := entsql.OpenDB(dialect.Postgres, db)

	// Create Ent client with schema routing
	entClient := ent.NewClient(
		ent.Driver(drv),
		ent.AlternateSchema(ent.DefaultSchemaConfig()),
	)

	m.mu.Lock()
	m.entClient = entClient
	m.currentDSN = dsn
	m.mu.Unlock()

	return nil
}

// reconnectIfChanged checks if DSN changed and reconnects if needed.
// Called on config reload.
func (m *dbManager) reconnectIfChanged() {
	newDSN := m.configService.DatabaseDSN()

	m.mu.RLock()
	currentDSN := m.currentDSN
	m.mu.RUnlock()

	// No change, nothing to do
	if newDSN == currentDSN {
		return
	}

	log.Printf("Database config changed, reconnecting...")

	// Open new database connection
	db, err := sql.Open("pgx", newDSN)
	if err != nil {
		log.Printf("Failed to open new database connection: %v (keeping old connection)", err)
		return
	}

	// Create Ent driver and client
	drv := entsql.OpenDB(dialect.Postgres, db)
	newClient := ent.NewClient(
		ent.Driver(drv),
		ent.AlternateSchema(ent.DefaultSchemaConfig()),
	)

	// Swap connections
	m.mu.Lock()
	oldClient := m.entClient
	oldDSN := m.currentDSN
	m.entClient = newClient
	m.currentDSN = newDSN
	m.mu.Unlock()

	log.Printf("Database reconnected successfully")

	// Notify callback
	if m.onReconnect != nil {
		m.onReconnect(oldDSN, newDSN)
	}

	// Close old connection after grace period (in background)
	go m.closeAfterGracePeriod(oldClient)
}

// closeAfterGracePeriod waits then closes the old connection.
func (m *dbManager) closeAfterGracePeriod(oldClient *ent.Client) {
	if oldClient == nil {
		return
	}

	log.Printf("Waiting %v before closing old database connection...", m.gracePeriod)
	time.Sleep(m.gracePeriod)

	if err := oldClient.Close(); err != nil {
		log.Printf("Failed to close old database connection: %v", err)
		return
	}

	log.Printf("Old database connection closed")
}

// EntClient returns the current Ent client (thread-safe).
func (m *dbManager) EntClient() *ent.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.entClient
}

// close closes the current database connection.
func (m *dbManager) close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.entClient == nil {
		return nil
	}

	return m.entClient.Close()
}
