package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/asheshgoplani/agent-deck/internal/database"
)

// Manager handles ledger database operations across multiple projects.
// It maintains a cache of open databases and provides thread-safe access.
type Manager struct {
	databases map[string]*database.DB
	mu        sync.RWMutex
	baseDir   string // Base directory for ledger data (~/.ledger)
}

// Global manager instance
var (
	globalManager *Manager
	managerOnce   sync.Once
)

// GetManager returns the global ledger manager, initializing it if needed.
func GetManager() *Manager {
	managerOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		globalManager = &Manager{
			databases: make(map[string]*database.DB),
			baseDir:   filepath.Join(homeDir, ".ledger"),
		}
	})
	return globalManager
}

// GetDB returns a database for the given project path.
// Creates and caches the database if not already open.
func (m *Manager) GetDB(projectPath string) (*database.DB, error) {
	m.mu.RLock()
	if db, ok := m.databases[projectPath]; ok {
		m.mu.RUnlock()
		return db, nil
	}
	m.mu.RUnlock()

	// Need to create database
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if db, ok := m.databases[projectPath]; ok {
		return db, nil
	}

	// Create new database for this project
	cfg := database.Config{
		ProjectPath: projectPath,
		BaseDir:     m.baseDir,
	}

	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger database for %s: %w", projectPath, err)
	}

	m.databases[projectPath] = db
	return db, nil
}

// CloseDB closes the database for a specific project.
func (m *Manager) CloseDB(projectPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if db, ok := m.databases[projectPath]; ok {
		delete(m.databases, projectPath)
		return db.Close()
	}
	return nil
}

// CloseAll closes all open databases.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for path, db := range m.databases {
		if err := db.Close(); err != nil {
			lastErr = err
		}
		delete(m.databases, path)
	}
	return lastErr
}

// GetBaseDir returns the base directory for ledger data.
func (m *Manager) GetBaseDir() string {
	return m.baseDir
}

// IsInitialized checks if a project has a ledger database.
func (m *Manager) IsInitialized(projectPath string) bool {
	slug := database.GenerateProjectSlug(projectPath)
	dbPath := filepath.Join(m.baseDir, slug, "ledger.db")
	_, err := os.Stat(dbPath)
	return err == nil
}
