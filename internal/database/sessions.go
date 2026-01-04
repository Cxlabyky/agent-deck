package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateSession creates a new session.
func (db *DB) CreateSession(s *Session) error {
	if s.ID == "" {
		s.ID = generateID()
	}
	if s.ProjectID == "" {
		s.ProjectID = db.projectID
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	var parentID interface{}
	if s.ParentSessionID != "" {
		parentID = s.ParentSessionID
	}

	_, err := db.conn.Exec(`
		INSERT INTO sessions (id, project_id, name, parent_session_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, s.ID, s.ProjectID, s.Name, parentID, s.CreatedAt, s.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by ID.
func (db *DB) GetSession(id string) (*Session, error) {
	s := &Session{}
	var parentID sql.NullString
	err := db.conn.QueryRow(`
		SELECT id, project_id, name, parent_session_id, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id).Scan(&s.ID, &s.ProjectID, &s.Name, &parentID, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if parentID.Valid {
		s.ParentSessionID = parentID.String
	}
	return s, nil
}

// UpdateSession updates a session.
func (db *DB) UpdateSession(s *Session) error {
	s.UpdatedAt = time.Now()
	var parentID interface{}
	if s.ParentSessionID != "" {
		parentID = s.ParentSessionID
	}

	result, err := db.conn.Exec(`
		UPDATE sessions SET name = ?, parent_session_id = ?, updated_at = ?
		WHERE id = ?
	`, s.Name, parentID, s.UpdatedAt, s.ID)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("session not found: %s", s.ID)
	}
	return nil
}

// ListSessions returns sessions for the current project.
func (db *DB) ListSessions() ([]*Session, error) {
	return db.ListSessionsByProject(db.projectID)
}

// ListSessionsByProject returns sessions for a specific project.
func (db *DB) ListSessionsByProject(projectID string) ([]*Session, error) {
	rows, err := db.conn.Query(`
		SELECT id, project_id, name, parent_session_id, created_at, updated_at
		FROM sessions WHERE project_id = ? ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		var parentID sql.NullString
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &parentID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		if parentID.Valid {
			s.ParentSessionID = parentID.String
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// GetOrCreateSession gets an existing session by name or creates a new one.
func (db *DB) GetOrCreateSession(name string) (*Session, error) {
	// Try to find existing session
	var s Session
	var parentID sql.NullString
	err := db.conn.QueryRow(`
		SELECT id, project_id, name, parent_session_id, created_at, updated_at
		FROM sessions WHERE project_id = ? AND name = ?
	`, db.projectID, name).Scan(&s.ID, &s.ProjectID, &s.Name, &parentID, &s.CreatedAt, &s.UpdatedAt)

	if err == nil {
		if parentID.Valid {
			s.ParentSessionID = parentID.String
		}
		return &s, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Create new session
	newSession := &Session{
		ProjectID: db.projectID,
		Name:      name,
	}
	if err := db.CreateSession(newSession); err != nil {
		return nil, err
	}
	return newSession, nil
}

// DeleteSession deletes a session.
func (db *DB) DeleteSession(id string) error {
	result, err := db.conn.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("session not found: %s", id)
	}
	return nil
}

// ForkSession creates a copy of a session for exploration.
func (db *DB) ForkSession(parentID string, newName string) (*Session, error) {
	parent, err := db.GetSession(parentID)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, fmt.Errorf("parent session not found: %s", parentID)
	}

	forked := &Session{
		ProjectID:       parent.ProjectID,
		Name:            newName,
		ParentSessionID: parentID,
	}
	if err := db.CreateSession(forked); err != nil {
		return nil, err
	}
	return forked, nil
}
