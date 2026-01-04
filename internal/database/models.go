package database

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// generateID creates a new random ID.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Project represents a Ledger project.
type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Session represents a working session within a project.
type Session struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	Name            string    `json:"name"`
	ParentSessionID string    `json:"parent_session_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// DecisionStatus represents the status of a decision.
type DecisionStatus string

const (
	DecisionStatusActive     DecisionStatus = "active"
	DecisionStatusOverridden DecisionStatus = "overridden"
	DecisionStatusArchived   DecisionStatus = "archived"
)

// Decision represents a logged decision.
type Decision struct {
	ID                   string         `json:"id"`
	ProjectID            string         `json:"project_id"`
	SessionID            string         `json:"session_id,omitempty"`
	Category             string         `json:"category"`
	Decision             string         `json:"decision"`
	Rationale            string         `json:"rationale"`
	AlternativesRejected string         `json:"alternatives_rejected,omitempty"` // JSON array stored as string
	Status               DecisionStatus `json:"status"`
	CreatedAt            time.Time      `json:"created_at"`
}

// Override represents a decision override with rationale.
type Override struct {
	ID         string    `json:"id"`
	DecisionID string    `json:"decision_id"`
	SessionID  string    `json:"session_id"`
	Rationale  string    `json:"rationale"`
	CreatedAt  time.Time `json:"created_at"`
}

// AttemptOutcome represents the outcome of an AI attempt.
type AttemptOutcome string

const (
	AttemptOutcomePending AttemptOutcome = "pending"
	AttemptOutcomeWorked  AttemptOutcome = "worked"
	AttemptOutcomeFailed  AttemptOutcome = "failed"
	AttemptOutcomePartial AttemptOutcome = "partial"
)

// AIAttempt represents a tracked AI suggestion and its outcome.
type AIAttempt struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	SessionID     string         `json:"session_id"`
	Problem       string         `json:"problem"`
	Suggestion    string         `json:"suggestion"`
	Outcome       AttemptOutcome `json:"outcome"`
	FailureReason string         `json:"failure_reason,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// Note represents a quick note.
type Note struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	SessionID string    `json:"session_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// DecisionFilter holds filter options for querying decisions.
type DecisionFilter struct {
	ProjectID string
	SessionID string
	Category  string
	Status    DecisionStatus
	Search    string // Search in decision text
	Limit     int
	Offset    int
}

// AttemptFilter holds filter options for querying AI attempts.
type AttemptFilter struct {
	ProjectID string
	SessionID string
	Outcome   AttemptOutcome
	Search    string // Search in problem/suggestion
	Limit     int
	Offset    int
}
