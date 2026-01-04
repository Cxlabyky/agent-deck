package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateOverride creates a new override record.
func (db *DB) CreateOverride(o *Override) error {
	if o.ID == "" {
		o.ID = generateID()
	}
	o.CreatedAt = time.Now()

	_, err := db.conn.Exec(`
		INSERT INTO overrides (id, decision_id, session_id, rationale, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, o.ID, o.DecisionID, o.SessionID, o.Rationale, o.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create override: %w", err)
	}
	return nil
}

// GetOverride retrieves an override by ID.
func (db *DB) GetOverride(id string) (*Override, error) {
	o := &Override{}
	err := db.conn.QueryRow(`
		SELECT id, decision_id, session_id, rationale, created_at
		FROM overrides WHERE id = ?
	`, id).Scan(&o.ID, &o.DecisionID, &o.SessionID, &o.Rationale, &o.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get override: %w", err)
	}
	return o, nil
}

// ListOverridesForDecision returns all overrides for a specific decision.
func (db *DB) ListOverridesForDecision(decisionID string) ([]*Override, error) {
	rows, err := db.conn.Query(`
		SELECT id, decision_id, session_id, rationale, created_at
		FROM overrides WHERE decision_id = ? ORDER BY created_at DESC
	`, decisionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list overrides: %w", err)
	}
	defer rows.Close()

	var overrides []*Override
	for rows.Next() {
		o := &Override{}
		if err := rows.Scan(&o.ID, &o.DecisionID, &o.SessionID, &o.Rationale, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}
		overrides = append(overrides, o)
	}
	return overrides, nil
}

// CountOverridesForDecision returns the number of times a decision has been overridden.
func (db *DB) CountOverridesForDecision(decisionID string) (int, error) {
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM overrides WHERE decision_id = ?
	`, decisionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count overrides: %w", err)
	}
	return count, nil
}

// GetOverridePatterns finds decisions that have been overridden multiple times.
// Returns decisions with override count >= minOverrides.
func (db *DB) GetOverridePatterns(minOverrides int) ([]struct {
	Decision      *Decision
	OverrideCount int
}, error) {
	rows, err := db.conn.Query(`
		SELECT d.id, d.project_id, d.session_id, d.category, d.decision, d.rationale,
		       d.alternatives_rejected, d.status, d.created_at, COUNT(o.id) as override_count
		FROM decisions d
		JOIN overrides o ON d.id = o.decision_id
		WHERE d.project_id = ?
		GROUP BY d.id
		HAVING override_count >= ?
		ORDER BY override_count DESC
	`, db.projectID, minOverrides)
	if err != nil {
		return nil, fmt.Errorf("failed to get override patterns: %w", err)
	}
	defer rows.Close()

	var results []struct {
		Decision      *Decision
		OverrideCount int
	}

	for rows.Next() {
		d := &Decision{}
		var sessionID, category, rationale, alternatives sql.NullString
		var count int

		if err := rows.Scan(&d.ID, &d.ProjectID, &sessionID, &category, &d.Decision, &rationale,
			&alternatives, &d.Status, &d.CreatedAt, &count); err != nil {
			return nil, fmt.Errorf("failed to scan pattern: %w", err)
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

		results = append(results, struct {
			Decision      *Decision
			OverrideCount int
		}{Decision: d, OverrideCount: count})
	}
	return results, nil
}

// FindTemporaryPatterns finds overrides with "temporary" in the rationale.
func (db *DB) FindTemporaryPatterns() ([]*Override, error) {
	rows, err := db.conn.Query(`
		SELECT o.id, o.decision_id, o.session_id, o.rationale, o.created_at
		FROM overrides o
		JOIN decisions d ON o.decision_id = d.id
		WHERE d.project_id = ? AND (
			LOWER(o.rationale) LIKE '%temporary%' OR
			LOWER(o.rationale) LIKE '%temp%' OR
			LOWER(o.rationale) LIKE '%quick fix%' OR
			LOWER(o.rationale) LIKE '%for now%' OR
			LOWER(o.rationale) LIKE '%hack%'
		)
		ORDER BY o.created_at DESC
	`, db.projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find temporary patterns: %w", err)
	}
	defer rows.Close()

	var overrides []*Override
	for rows.Next() {
		o := &Override{}
		if err := rows.Scan(&o.ID, &o.DecisionID, &o.SessionID, &o.Rationale, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan override: %w", err)
		}
		overrides = append(overrides, o)
	}
	return overrides, nil
}
