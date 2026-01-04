// Package database provides SQLite storage for Ledger's decision and attempt tracking.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the SQLite database connection with Ledger-specific operations.
type DB struct {
	conn      *sql.DB
	projectID string
	mu        sync.RWMutex
}

// Config holds database configuration options.
type Config struct {
	// ProjectPath is the full path to the project directory
	ProjectPath string
	// BaseDir overrides the default ~/.ledger location
	BaseDir string
}

// GenerateProjectSlug creates a filesystem-safe slug from a project path.
// Uses the last two path components (parent/name) for uniqueness.
func GenerateProjectSlug(projectPath string) string {
	// Clean and get absolute path
	cleanPath := filepath.Clean(projectPath)

	// Get parent and base for a more unique slug
	base := filepath.Base(cleanPath)
	parent := filepath.Base(filepath.Dir(cleanPath))

	// Combine for uniqueness (e.g., "repos/myproject" -> "repos-myproject")
	if parent != "" && parent != "." && parent != "/" {
		return sanitizeSlug(parent + "-" + base)
	}
	return sanitizeSlug(base)
}

// sanitizeSlug removes or replaces characters that are problematic for filesystems.
func sanitizeSlug(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
			result = append(result, c)
		case c >= 'A' && c <= 'Z':
			result = append(result, c+32) // lowercase
		case c >= '0' && c <= '9':
			result = append(result, c)
		case c == '-' || c == '_':
			result = append(result, c)
		case c == ' ' || c == '/':
			result = append(result, '-')
		}
	}
	return string(result)
}

// DefaultBasePath returns the default Ledger data directory.
func DefaultBasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".ledger"), nil
}

// New creates a new database connection for the given project.
func New(cfg Config) (*DB, error) {
	baseDir := cfg.BaseDir
	if baseDir == "" {
		var err error
		baseDir, err = DefaultBasePath()
		if err != nil {
			return nil, err
		}
	}

	// Generate project slug from path
	projectSlug := GenerateProjectSlug(cfg.ProjectPath)

	// Create project directory
	projectDir := filepath.Join(baseDir, projectSlug)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Open database
	dbPath := filepath.Join(projectDir, "ledger.db")
	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db := &DB{
		conn: conn,
	}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Ensure project exists and get its ID
	projectID, err := db.ensureProject(projectSlug, cfg.ProjectPath)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ensure project: %w", err)
	}
	db.projectID = projectID

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.conn.Close()
}

// ProjectID returns the current project's ID.
func (db *DB) ProjectID() string {
	return db.projectID
}

// Conn returns the underlying database connection for advanced queries.
// Use with caution - prefer the typed methods.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// ensureProject creates or retrieves the project, returning its ID.
func (db *DB) ensureProject(name, path string) (string, error) {
	// Try to get existing project
	var id string
	err := db.conn.QueryRow(
		"SELECT id FROM projects WHERE name = ?",
		name,
	).Scan(&id)

	if err == nil {
		// Update path if it changed
		_, _ = db.conn.Exec("UPDATE projects SET path = ? WHERE id = ?", path, id)
		return id, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	// Create new project
	project := &Project{
		Name: name,
		Path: path,
	}
	if err := db.CreateProject(project); err != nil {
		return "", err
	}

	return project.ID, nil
}

// Transaction executes a function within a database transaction.
func (db *DB) Transaction(fn func(tx *sql.Tx) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
