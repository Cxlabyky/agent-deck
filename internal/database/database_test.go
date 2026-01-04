package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "ledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test project path
	projectPath := "/Users/test/projects/myproject"

	// Create database
	cfg := Config{
		ProjectPath: projectPath,
		BaseDir:     tmpDir,
	}
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Verify database file was created
	slug := GenerateProjectSlug(projectPath)
	dbPath := filepath.Join(tmpDir, slug, "ledger.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file not created at %s", dbPath)
	}

	// Verify project ID is set
	if db.ProjectID() == "" {
		t.Error("project ID should not be empty")
	}
}

func TestGenerateProjectSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/test/projects/myproject", "projects-myproject"},
		{"/home/user/code/api-server", "code-api-server"},
		{"/tmp/test", "tmp-test"},
		{"myproject", "myproject"},
		{"/Users/Test/Projects/MyProject", "projects-myproject"}, // Should lowercase
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateProjectSlug(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateProjectSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecisionCRUD(t *testing.T) {
	// Create temp database
	tmpDir, err := os.MkdirTemp("", "ledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Config{
		ProjectPath: "/test/project",
		BaseDir:     tmpDir,
	}
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create a decision
	decision := &Decision{
		Category:  "architecture",
		Decision:  "Use SQLite for local storage",
		Rationale: "Simple, embedded, no server required",
	}
	if err := db.CreateDecision(decision); err != nil {
		t.Fatalf("failed to create decision: %v", err)
	}

	// Verify ID was generated
	if decision.ID == "" {
		t.Error("decision ID should be generated")
	}

	// Get decision
	retrieved, err := db.GetDecision(decision.ID)
	if err != nil {
		t.Fatalf("failed to get decision: %v", err)
	}
	if retrieved == nil {
		t.Fatal("decision not found")
	}
	if retrieved.Decision != decision.Decision {
		t.Errorf("got decision %q, want %q", retrieved.Decision, decision.Decision)
	}
	if retrieved.Status != DecisionStatusActive {
		t.Errorf("got status %q, want %q", retrieved.Status, DecisionStatusActive)
	}

	// List decisions
	decisions, err := db.ListActiveDecisions()
	if err != nil {
		t.Fatalf("failed to list decisions: %v", err)
	}
	if len(decisions) != 1 {
		t.Errorf("got %d decisions, want 1", len(decisions))
	}
}

func TestAttemptCRUD(t *testing.T) {
	// Create temp database
	tmpDir, err := os.MkdirTemp("", "ledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Config{
		ProjectPath: "/test/project",
		BaseDir:     tmpDir,
	}
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create a session first
	sess := &Session{
		Name: "test-session",
	}
	if err := db.CreateSession(sess); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create an attempt
	attempt := &AIAttempt{
		SessionID:  sess.ID,
		Problem:    "Tests failing with timeout errors",
		Suggestion: "Increase test timeout to 30 seconds",
	}
	if err := db.CreateAttempt(attempt); err != nil {
		t.Fatalf("failed to create attempt: %v", err)
	}

	// Verify ID was generated and status is pending
	if attempt.ID == "" {
		t.Error("attempt ID should be generated")
	}
	if attempt.Outcome != AttemptOutcomePending {
		t.Errorf("got outcome %q, want %q", attempt.Outcome, AttemptOutcomePending)
	}

	// Mark as worked
	if err := db.MarkAttemptWorked(attempt.ID); err != nil {
		t.Fatalf("failed to mark attempt as worked: %v", err)
	}

	// Verify status changed
	updated, err := db.GetAttempt(attempt.ID)
	if err != nil {
		t.Fatalf("failed to get attempt: %v", err)
	}
	if updated.Outcome != AttemptOutcomeWorked {
		t.Errorf("got outcome %q, want %q", updated.Outcome, AttemptOutcomeWorked)
	}
}

func TestNoteCRUD(t *testing.T) {
	// Create temp database
	tmpDir, err := os.MkdirTemp("", "ledger-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := Config{
		ProjectPath: "/test/project",
		BaseDir:     tmpDir,
	}
	db, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Create a quick note
	note, err := db.QuickNote("Remember to update the README")
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	// Verify note was created
	if note.ID == "" {
		t.Error("note ID should be generated")
	}
	if note.Content != "Remember to update the README" {
		t.Errorf("got content %q, want %q", note.Content, "Remember to update the README")
	}

	// Search notes
	notes, err := db.SearchNotes("README")
	if err != nil {
		t.Fatalf("failed to search notes: %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("got %d notes, want 1", len(notes))
	}
}
