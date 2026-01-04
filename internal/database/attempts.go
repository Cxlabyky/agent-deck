package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CreateAttempt creates a new AI attempt record.
func (db *DB) CreateAttempt(a *AIAttempt) error {
	if a.ID == "" {
		a.ID = generateID()
	}
	if a.ProjectID == "" {
		a.ProjectID = db.projectID
	}
	if a.Outcome == "" {
		a.Outcome = AttemptOutcomePending
	}
	a.CreatedAt = time.Now()

	_, err := db.conn.Exec(`
		INSERT INTO ai_attempts (id, project_id, session_id, problem, suggestion, outcome, failure_reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.ProjectID, a.SessionID, a.Problem, a.Suggestion, a.Outcome, a.FailureReason, a.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create attempt: %w", err)
	}
	return nil
}

// GetAttempt retrieves an attempt by ID.
func (db *DB) GetAttempt(id string) (*AIAttempt, error) {
	a := &AIAttempt{}
	var failureReason sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, project_id, session_id, problem, suggestion, outcome, failure_reason, created_at
		FROM ai_attempts WHERE id = ?
	`, id).Scan(&a.ID, &a.ProjectID, &a.SessionID, &a.Problem, &a.Suggestion, &a.Outcome, &failureReason, &a.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attempt: %w", err)
	}

	if failureReason.Valid {
		a.FailureReason = failureReason.String
	}
	return a, nil
}

// UpdateAttemptOutcome updates the outcome of an attempt.
func (db *DB) UpdateAttemptOutcome(id string, outcome AttemptOutcome, failureReason string) error {
	var reason interface{}
	if failureReason != "" {
		reason = failureReason
	}

	result, err := db.conn.Exec(`
		UPDATE ai_attempts SET outcome = ?, failure_reason = ? WHERE id = ?
	`, outcome, reason, id)

	if err != nil {
		return fmt.Errorf("failed to update attempt outcome: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("attempt not found: %s", id)
	}
	return nil
}

// MarkAttemptWorked marks an attempt as successful.
func (db *DB) MarkAttemptWorked(id string) error {
	return db.UpdateAttemptOutcome(id, AttemptOutcomeWorked, "")
}

// MarkAttemptFailed marks an attempt as failed with a reason.
func (db *DB) MarkAttemptFailed(id, reason string) error {
	return db.UpdateAttemptOutcome(id, AttemptOutcomeFailed, reason)
}

// MarkAttemptPartial marks an attempt as partially successful.
func (db *DB) MarkAttemptPartial(id, notes string) error {
	return db.UpdateAttemptOutcome(id, AttemptOutcomePartial, notes)
}

// ListAttempts returns attempts based on filter criteria.
func (db *DB) ListAttempts(filter AttemptFilter) ([]*AIAttempt, error) {
	query := `
		SELECT id, project_id, session_id, problem, suggestion, outcome, failure_reason, created_at
		FROM ai_attempts WHERE 1=1
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

	if filter.Outcome != "" {
		query += " AND outcome = ?"
		args = append(args, filter.Outcome)
	}

	if filter.Search != "" {
		query += " AND (problem LIKE ? OR suggestion LIKE ?)"
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
		return nil, fmt.Errorf("failed to list attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*AIAttempt
	for rows.Next() {
		a := &AIAttempt{}
		var failureReason sql.NullString

		if err := rows.Scan(&a.ID, &a.ProjectID, &a.SessionID, &a.Problem, &a.Suggestion, &a.Outcome, &failureReason, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if failureReason.Valid {
			a.FailureReason = failureReason.String
		}
		attempts = append(attempts, a)
	}
	return attempts, nil
}

// ListFailedAttempts returns all failed attempts for the current project.
func (db *DB) ListFailedAttempts() ([]*AIAttempt, error) {
	return db.ListAttempts(AttemptFilter{
		Outcome: AttemptOutcomeFailed,
	})
}

// FindSimilarFailedAttempts finds failed attempts with similar problem descriptions.
func (db *DB) FindSimilarFailedAttempts(problem string) ([]*AIAttempt, error) {
	// Simple keyword matching - extract significant words
	words := strings.Fields(strings.ToLower(problem))
	if len(words) == 0 {
		return nil, nil
	}

	// Build query with OR conditions for each word
	query := `
		SELECT id, project_id, session_id, problem, suggestion, outcome, failure_reason, created_at
		FROM ai_attempts
		WHERE project_id = ? AND outcome = 'failed' AND (
	`
	var args []interface{}
	args = append(args, db.projectID)

	var conditions []string
	for _, word := range words {
		if len(word) < 3 {
			continue // Skip short words
		}
		conditions = append(conditions, "LOWER(problem) LIKE ?")
		args = append(args, "%"+word+"%")
	}

	if len(conditions) == 0 {
		return nil, nil
	}

	query += strings.Join(conditions, " OR ") + ") ORDER BY created_at DESC LIMIT 10"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*AIAttempt
	for rows.Next() {
		a := &AIAttempt{}
		var failureReason sql.NullString

		if err := rows.Scan(&a.ID, &a.ProjectID, &a.SessionID, &a.Problem, &a.Suggestion, &a.Outcome, &failureReason, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if failureReason.Valid {
			a.FailureReason = failureReason.String
		}
		attempts = append(attempts, a)
	}
	return attempts, nil
}

// GetRecentAttempts returns the most recent attempts.
func (db *DB) GetRecentAttempts(limit int) ([]*AIAttempt, error) {
	return db.ListAttempts(AttemptFilter{
		Limit: limit,
	})
}

// GetAttemptStats returns statistics about attempts.
func (db *DB) GetAttemptStats() (map[AttemptOutcome]int, error) {
	rows, err := db.conn.Query(`
		SELECT outcome, COUNT(*) FROM ai_attempts
		WHERE project_id = ?
		GROUP BY outcome
	`, db.projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempt stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[AttemptOutcome]int)
	for rows.Next() {
		var outcome AttemptOutcome
		var count int
		if err := rows.Scan(&outcome, &count); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats[outcome] = count
	}
	return stats, nil
}

// DeleteAttempt deletes an attempt.
func (db *DB) DeleteAttempt(id string) error {
	result, err := db.conn.Exec("DELETE FROM ai_attempts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete attempt: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("attempt not found: %s", id)
	}
	return nil
}

// GetRecurringFailures finds suggestions that have failed multiple times.
func (db *DB) GetRecurringFailures(minFailures int) ([]struct {
	Suggestion   string
	FailureCount int
	LastFailure  time.Time
}, error) {
	rows, err := db.conn.Query(`
		SELECT suggestion, COUNT(*) as fail_count, MAX(created_at) as last_failure
		FROM ai_attempts
		WHERE project_id = ? AND outcome = 'failed'
		GROUP BY suggestion
		HAVING fail_count >= ?
		ORDER BY fail_count DESC
	`, db.projectID, minFailures)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring failures: %w", err)
	}
	defer rows.Close()

	var results []struct {
		Suggestion   string
		FailureCount int
		LastFailure  time.Time
	}

	for rows.Next() {
		var r struct {
			Suggestion   string
			FailureCount int
			LastFailure  time.Time
		}
		if err := rows.Scan(&r.Suggestion, &r.FailureCount, &r.LastFailure); err != nil {
			return nil, fmt.Errorf("failed to scan recurring failure: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}
