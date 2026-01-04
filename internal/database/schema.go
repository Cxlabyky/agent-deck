package database

// Schema version for migrations
const schemaVersion = 1

// initSchema creates all tables if they don't exist.
func (db *DB) initSchema() error {
	// Create schema version table
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Check current version
	var currentVersion int
	err = db.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Apply migrations
	if currentVersion < 1 {
		if err := db.migrateV1(); err != nil {
			return err
		}
	}

	return nil
}

// migrateV1 creates the initial schema.
func (db *DB) migrateV1() error {
	schema := `
	-- Projects
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Sessions (within projects)
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		name TEXT,
		parent_session_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_session_id) REFERENCES sessions(id) ON DELETE SET NULL
	);

	-- Decisions
	CREATE TABLE IF NOT EXISTS decisions (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		session_id TEXT,
		category TEXT,
		decision TEXT NOT NULL,
		rationale TEXT,
		alternatives_rejected TEXT,
		status TEXT DEFAULT 'active' CHECK(status IN ('active', 'overridden', 'archived')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
	);

	-- Overrides
	CREATE TABLE IF NOT EXISTS overrides (
		id TEXT PRIMARY KEY,
		decision_id TEXT NOT NULL,
		session_id TEXT NOT NULL,
		rationale TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (decision_id) REFERENCES decisions(id) ON DELETE CASCADE,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	-- AI Attempts
	CREATE TABLE IF NOT EXISTS ai_attempts (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		session_id TEXT NOT NULL,
		problem TEXT NOT NULL,
		suggestion TEXT NOT NULL,
		outcome TEXT DEFAULT 'pending' CHECK(outcome IN ('pending', 'worked', 'failed', 'partial')),
		failure_reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	-- Notes
	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		session_id TEXT,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
	);

	-- Indexes for common queries
	CREATE INDEX IF NOT EXISTS idx_decisions_project ON decisions(project_id);
	CREATE INDEX IF NOT EXISTS idx_decisions_status ON decisions(status);
	CREATE INDEX IF NOT EXISTS idx_decisions_category ON decisions(category);
	CREATE INDEX IF NOT EXISTS idx_ai_attempts_project ON ai_attempts(project_id);
	CREATE INDEX IF NOT EXISTS idx_ai_attempts_session ON ai_attempts(session_id);
	CREATE INDEX IF NOT EXISTS idx_ai_attempts_outcome ON ai_attempts(outcome);
	CREATE INDEX IF NOT EXISTS idx_overrides_decision ON overrides(decision_id);
	CREATE INDEX IF NOT EXISTS idx_notes_project ON notes(project_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project_id);

	-- Record migration
	INSERT INTO schema_version (version) VALUES (1);
	`

	_, err := db.conn.Exec(schema)
	return err
}
