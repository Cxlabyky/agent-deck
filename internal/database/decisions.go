package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CreateDecision creates a new decision.
func (db *DB) CreateDecision(d *Decision) error {
	if d.ID == "" {
		d.ID = generateID()
	}
	if d.ProjectID == "" {
		d.ProjectID = db.projectID
	}
	if d.Status == "" {
		d.Status = DecisionStatusActive
	}
	d.CreatedAt = time.Now()

	var sessionID interface{}
	if d.SessionID != "" {
		sessionID = d.SessionID
	}

	_, err := db.conn.Exec(`
		INSERT INTO decisions (id, project_id, session_id, category, decision, rationale, alternatives_rejected, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.ProjectID, sessionID, d.Category, d.Decision, d.Rationale, d.AlternativesRejected, d.Status, d.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create decision: %w", err)
	}
	return nil
}

// GetDecision retrieves a decision by ID.
func (db *DB) GetDecision(id string) (*Decision, error) {
	d := &Decision{}
	var sessionID sql.NullString
	var category, rationale, alternatives sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, project_id, session_id, category, decision, rationale, alternatives_rejected, status, created_at
		FROM decisions WHERE id = ?
	`, id).Scan(&d.ID, &d.ProjectID, &sessionID, &category, &d.Decision, &rationale, &alternatives, &d.Status, &d.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get decision: %w", err)
	}

	if sessionID.Valid {
		d.SessionID = sessionID.String
	}
	if category.Valid {
		d.Category = category.String
	}
	if rationale.Valid {
		d.Rationale = rationale.String
	}
	if alternatives.Valid {
		d.AlternativesRejected = alternatives.String
	}

	return d, nil
}

// UpdateDecision updates a decision.
func (db *DB) UpdateDecision(d *Decision) error {
	var sessionID interface{}
	if d.SessionID != "" {
		sessionID = d.SessionID
	}

	result, err := db.conn.Exec(`
		UPDATE decisions SET category = ?, decision = ?, rationale = ?, alternatives_rejected = ?, status = ?
		WHERE id = ?
	`, d.Category, d.Decision, d.Rationale, d.AlternativesRejected, d.Status, d.ID)

	_ = sessionID // Not updating session_id

	if err != nil {
		return fmt.Errorf("failed to update decision: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("decision not found: %s", d.ID)
	}
	return nil
}

// ListDecisions returns decisions based on filter criteria.
func (db *DB) ListDecisions(filter DecisionFilter) ([]*Decision, error) {
	query := `
		SELECT id, project_id, session_id, category, decision, rationale, alternatives_rejected, status, created_at
		FROM decisions WHERE 1=1
	`
	var args []interface{}

	// Apply filters
	if filter.ProjectID != "" {
		query += " AND project_id = ?"
		args = append(args, filter.ProjectID)
	} else {
		query += " AND project_id = ?"
		args = append(args, db.projectID)
	}

	if filter.SessionID != "" {
		query += " AND session_id = ?"
		args = append(args, filter.SessionID)
	}

	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}

	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}

	if filter.Search != "" {
		query += " AND (decision LIKE ? OR rationale LIKE ?)"
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions: %w", err)
	}
	defer rows.Close()

	var decisions []*Decision
	for rows.Next() {
		d := &Decision{}
		var sessionID, category, rationale, alternatives sql.NullString

		if err := rows.Scan(&d.ID, &d.ProjectID, &sessionID, &category, &d.Decision, &rationale, &alternatives, &d.Status, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}

		if sessionID.Valid {
			d.SessionID = sessionID.String
		}
		if category.Valid {
			d.Category = category.String
		}
		if rationale.Valid {
			d.Rationale = rationale.String
		}
		if alternatives.Valid {
			d.AlternativesRejected = alternatives.String
		}

		decisions = append(decisions, d)
	}
	return decisions, nil
}

// ListActiveDecisions returns all active decisions for the current project.
func (db *DB) ListActiveDecisions() ([]*Decision, error) {
	return db.ListDecisions(DecisionFilter{
		Status: DecisionStatusActive,
	})
}

// FindRelevantDecisions finds decisions matching keywords in the query.
func (db *DB) FindRelevantDecisions(query string) ([]*Decision, error) {
	// Simple keyword matching - extract significant words
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil, nil
	}

	// Build query with OR conditions for each word
	sqlQuery := `
		SELECT id, project_id, session_id, category, decision, rationale, alternatives_rejected, status, created_at
		FROM decisions
		WHERE project_id = ? AND status = 'active' AND (
	`
	var args []interface{}
	args = append(args, db.projectID)

	var conditions []string
	for _, word := range words {
		if len(word) < 3 {
			continue // Skip short words
		}
		conditions = append(conditions, "LOWER(decision) LIKE ? OR LOWER(category) LIKE ?")
		searchTerm := "%" + word + "%"
		args = append(args, searchTerm, searchTerm)
	}

	if len(conditions) == 0 {
		return nil, nil
	}

	sqlQuery += strings.Join(conditions, " OR ") + ") ORDER BY created_at DESC LIMIT 10"

	rows, err := db.conn.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find relevant decisions: %w", err)
	}
	defer rows.Close()

	var decisions []*Decision
	for rows.Next() {
		d := &Decision{}
		var sessionID, category, rationale, alternatives sql.NullString

		if err := rows.Scan(&d.ID, &d.ProjectID, &sessionID, &category, &d.Decision, &rationale, &alternatives, &d.Status, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan decision: %w", err)
		}

		if sessionID.Valid {
			d.SessionID = sessionID.String
		}
		if category.Valid {
			d.Category = category.String
		}
		if rationale.Valid {
			d.Rationale = rationale.String
		}
		if alternatives.Valid {
			d.AlternativesRejected = alternatives.String
		}

		decisions = append(decisions, d)
	}
	return decisions, nil
}

// ArchiveDecision marks a decision as archived.
func (db *DB) ArchiveDecision(id string) error {
	result, err := db.conn.Exec(`
		UPDATE decisions SET status = ? WHERE id = ?
	`, DecisionStatusArchived, id)

	if err != nil {
		return fmt.Errorf("failed to archive decision: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("decision not found: %s", id)
	}
	return nil
}

// OverrideDecision marks a decision as overridden and creates an override record.
func (db *DB) OverrideDecision(decisionID, sessionID, rationale string) (*Override, error) {
	// Update decision status
	result, err := db.conn.Exec(`
		UPDATE decisions SET status = ? WHERE id = ?
	`, DecisionStatusOverridden, decisionID)

	if err != nil {
		return nil, fmt.Errorf("failed to override decision: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("decision not found: %s", decisionID)
	}

	// Create override record
	override := &Override{
		DecisionID: decisionID,
		SessionID:  sessionID,
		Rationale:  rationale,
	}
	if err := db.CreateOverride(override); err != nil {
		return nil, err
	}

	return override, nil
}

// DeleteDecision deletes a decision.
func (db *DB) DeleteDecision(id string) error {
	result, err := db.conn.Exec("DELETE FROM decisions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete decision: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("decision not found: %s", id)
	}
	return nil
}

// GetDecisionCategories returns all unique categories used in the project.
func (db *DB) GetDecisionCategories() ([]string, error) {
	rows, err := db.conn.Query(`
		SELECT DISTINCT category FROM decisions
		WHERE project_id = ? AND category IS NOT NULL AND category != ''
		ORDER BY category
	`, db.projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}
	return categories, nil
}
