package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateNote creates a new note.
func (db *DB) CreateNote(n *Note) error {
	if n.ID == "" {
		n.ID = generateID()
	}
	if n.ProjectID == "" {
		n.ProjectID = db.projectID
	}
	n.CreatedAt = time.Now()

	var sessionID interface{}
	if n.SessionID != "" {
		sessionID = n.SessionID
	}

	_, err := db.conn.Exec(`
		INSERT INTO notes (id, project_id, session_id, content, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, n.ID, n.ProjectID, sessionID, n.Content, n.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// GetNote retrieves a note by ID.
func (db *DB) GetNote(id string) (*Note, error) {
	n := &Note{}
	var sessionID sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, project_id, session_id, content, created_at
		FROM notes WHERE id = ?
	`, id).Scan(&n.ID, &n.ProjectID, &sessionID, &n.Content, &n.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if sessionID.Valid {
		n.SessionID = sessionID.String
	}
	return n, nil
}

// UpdateNote updates a note's content.
func (db *DB) UpdateNote(n *Note) error {
	result, err := db.conn.Exec(`
		UPDATE notes SET content = ? WHERE id = ?
	`, n.Content, n.ID)

	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("note not found: %s", n.ID)
	}
	return nil
}

// ListNotes returns all notes for the current project.
func (db *DB) ListNotes() ([]*Note, error) {
	return db.ListNotesByProject(db.projectID)
}

// ListNotesByProject returns all notes for a specific project.
func (db *DB) ListNotesByProject(projectID string) ([]*Note, error) {
	rows, err := db.conn.Query(`
		SELECT id, project_id, session_id, content, created_at
		FROM notes WHERE project_id = ? ORDER BY created_at DESC
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		var sessionID sql.NullString

		if err := rows.Scan(&n.ID, &n.ProjectID, &sessionID, &n.Content, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if sessionID.Valid {
			n.SessionID = sessionID.String
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// ListNotesBySession returns notes for a specific session.
func (db *DB) ListNotesBySession(sessionID string) ([]*Note, error) {
	rows, err := db.conn.Query(`
		SELECT id, project_id, session_id, content, created_at
		FROM notes WHERE session_id = ? ORDER BY created_at DESC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		var sessID sql.NullString

		if err := rows.Scan(&n.ID, &n.ProjectID, &sessID, &n.Content, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if sessID.Valid {
			n.SessionID = sessID.String
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// GetRecentNotes returns the most recent notes.
func (db *DB) GetRecentNotes(limit int) ([]*Note, error) {
	rows, err := db.conn.Query(`
		SELECT id, project_id, session_id, content, created_at
		FROM notes WHERE project_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, db.projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		var sessionID sql.NullString

		if err := rows.Scan(&n.ID, &n.ProjectID, &sessionID, &n.Content, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if sessionID.Valid {
			n.SessionID = sessionID.String
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// SearchNotes searches notes by content.
func (db *DB) SearchNotes(query string) ([]*Note, error) {
	searchTerm := "%" + query + "%"
	rows, err := db.conn.Query(`
		SELECT id, project_id, session_id, content, created_at
		FROM notes WHERE project_id = ? AND content LIKE ?
		ORDER BY created_at DESC
	`, db.projectID, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		n := &Note{}
		var sessionID sql.NullString

		if err := rows.Scan(&n.ID, &n.ProjectID, &sessionID, &n.Content, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if sessionID.Valid {
			n.SessionID = sessionID.String
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// DeleteNote deletes a note.
func (db *DB) DeleteNote(id string) error {
	result, err := db.conn.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("note not found: %s", id)
	}
	return nil
}

// QuickNote creates a note with just content (uses current project).
func (db *DB) QuickNote(content string) (*Note, error) {
	note := &Note{
		Content: content,
	}
	if err := db.CreateNote(note); err != nil {
		return nil, err
	}
	return note, nil
}
