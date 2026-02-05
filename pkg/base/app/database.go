package app

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
)

// dbManager handles database connections with hot-reload support.
type dbManager struct {
	configService *config.Service
	gracePeriod   time.Duration
	onReconnect   func(oldDSN, newDSN string)

	db         *gorm.DB
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	m.mu.Lock()
	m.db = db
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

	// Create new connection
	newDB, err := gorm.Open(postgres.Open(newDSN), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to reconnect to database: %v (keeping old connection)", err)
		return
	}

	// Swap connections
	m.mu.Lock()
	oldDB := m.db
	oldDSN := m.currentDSN
	m.db = newDB
	m.currentDSN = newDSN
	m.mu.Unlock()

	log.Printf("Database reconnected successfully")

	// Notify callback
	if m.onReconnect != nil {
		m.onReconnect(oldDSN, newDSN)
	}

	// Close old connection after grace period (in background)
	go m.closeAfterGracePeriod(oldDB)
}

// closeAfterGracePeriod waits then closes the old connection.
func (m *dbManager) closeAfterGracePeriod(oldDB *gorm.DB) {
	if oldDB == nil {
		return
	}

	log.Printf("Waiting %v before closing old database connection...", m.gracePeriod)
	time.Sleep(m.gracePeriod)

	sqlDB, err := oldDB.DB()
	if err != nil {
		log.Printf("Failed to get underlying DB for closure: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close old database connection: %v", err)
		return
	}

	log.Printf("Old database connection closed")
}

// DB returns the current database connection (thread-safe).
func (m *dbManager) DB() *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// close closes the current database connection.
func (m *dbManager) close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return nil
	}

	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("get underlying DB: %w", err)
	}

	return sqlDB.Close()
}
