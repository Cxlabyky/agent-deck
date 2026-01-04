package ui

import (
	"testing"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/database"
)

func TestNewDecisionListPanel(t *testing.T) {
	p := NewDecisionListPanel()

	if p == nil {
		t.Fatal("NewDecisionListPanel should return non-nil")
	}

	if len(p.Decisions()) != 0 {
		t.Error("new panel should have no decisions")
	}

	if p.Cursor() != 0 {
		t.Error("initial cursor should be 0")
	}
}

func TestDecisionListPanel_SetDecisions(t *testing.T) {
	p := NewDecisionListPanel()

	decisions := []*database.Decision{
		{ID: "1", Category: "arch", Decision: "Use SQLite", Status: database.DecisionStatusActive, CreatedAt: time.Now()},
		{ID: "2", Category: "test", Decision: "Use pytest", Status: database.DecisionStatusActive, CreatedAt: time.Now()},
	}

	p.SetDecisions(decisions)

	if len(p.Decisions()) != 2 {
		t.Errorf("expected 2 decisions, got %d", len(p.Decisions()))
	}
}

func TestDecisionListPanel_Navigation(t *testing.T) {
	p := NewDecisionListPanel()

	decisions := []*database.Decision{
		{ID: "1", Category: "arch", Decision: "Decision 1", Status: database.DecisionStatusActive},
		{ID: "2", Category: "test", Decision: "Decision 2", Status: database.DecisionStatusActive},
		{ID: "3", Category: "deps", Decision: "Decision 3", Status: database.DecisionStatusActive},
	}

	p.SetDecisions(decisions)

	// Initial cursor should be 0
	if p.Cursor() != 0 {
		t.Errorf("initial cursor should be 0, got %d", p.Cursor())
	}

	// Move down
	p.MoveDown()
	if p.Cursor() != 1 {
		t.Errorf("cursor should be 1 after MoveDown, got %d", p.Cursor())
	}

	// Move down again
	p.MoveDown()
	if p.Cursor() != 2 {
		t.Errorf("cursor should be 2 after second MoveDown, got %d", p.Cursor())
	}

	// Move down at end should not change cursor
	p.MoveDown()
	if p.Cursor() != 2 {
		t.Errorf("cursor should still be 2 at end, got %d", p.Cursor())
	}

	// Move up
	p.MoveUp()
	if p.Cursor() != 1 {
		t.Errorf("cursor should be 1 after MoveUp, got %d", p.Cursor())
	}

	// Move up to beginning
	p.MoveUp()
	if p.Cursor() != 0 {
		t.Errorf("cursor should be 0, got %d", p.Cursor())
	}

	// Move up at beginning should not change cursor
	p.MoveUp()
	if p.Cursor() != 0 {
		t.Errorf("cursor should still be 0 at beginning, got %d", p.Cursor())
	}
}

func TestDecisionListPanel_Selected(t *testing.T) {
	p := NewDecisionListPanel()

	// No decisions - should return nil
	if p.Selected() != nil {
		t.Error("Selected() should return nil when no decisions")
	}

	decisions := []*database.Decision{
		{ID: "1", Category: "arch", Decision: "Decision 1", Status: database.DecisionStatusActive},
		{ID: "2", Category: "test", Decision: "Decision 2", Status: database.DecisionStatusActive},
	}

	p.SetDecisions(decisions)

	// First should be selected by default
	selected := p.Selected()
	if selected == nil {
		t.Fatal("Selected() should return non-nil")
	}
	if selected.ID != "1" {
		t.Errorf("expected ID '1', got %q", selected.ID)
	}

	// Move to second
	p.MoveDown()
	selected = p.Selected()
	if selected.ID != "2" {
		t.Errorf("expected ID '2', got %q", selected.ID)
	}
}

func TestDecisionListPanel_SetSize(t *testing.T) {
	p := NewDecisionListPanel()
	p.SetSize(100, 50)

	if p.width != 100 || p.height != 50 {
		t.Errorf("expected width=100, height=50, got width=%d, height=%d", p.width, p.height)
	}
}

func TestDecisionListPanel_SetProjectPath(t *testing.T) {
	p := NewDecisionListPanel()
	p.SetProjectPath("/test/project")

	if p.GetProjectPath() != "/test/project" {
		t.Errorf("expected '/test/project', got %q", p.GetProjectPath())
	}
}

func TestDecisionListPanel_RenderEmpty(t *testing.T) {
	p := NewDecisionListPanel()
	p.SetSize(60, 20)

	result := p.Render(60, 20)

	if result == "" {
		t.Error("Render should return non-empty string even with no decisions")
	}

	// Should contain empty state message
	if !containsString(result, "No decisions yet") {
		t.Error("empty state should contain 'No decisions yet'")
	}
}

func TestDecisionListPanel_Render(t *testing.T) {
	p := NewDecisionListPanel()
	p.SetSize(60, 20)

	decisions := []*database.Decision{
		{ID: "1", Category: "arch", Decision: "Use SQLite for storage", Status: database.DecisionStatusActive, CreatedAt: time.Now()},
		{ID: "2", Category: "testing", Decision: "Use pytest for tests", Status: database.DecisionStatusActive, CreatedAt: time.Now()},
	}
	p.SetDecisions(decisions)

	result := p.Render(60, 20)

	if result == "" {
		t.Error("Render should return non-empty string")
	}

	// Should contain decision content (category is truncated at 10 chars)
	if !containsString(result, "arch") {
		t.Error("render should contain category 'arch'")
	}
}

func TestDecisionListPanel_CursorBoundsOnSetDecisions(t *testing.T) {
	p := NewDecisionListPanel()

	// Set some decisions and move cursor
	decisions := []*database.Decision{
		{ID: "1", Category: "arch", Decision: "Decision 1", Status: database.DecisionStatusActive},
		{ID: "2", Category: "test", Decision: "Decision 2", Status: database.DecisionStatusActive},
		{ID: "3", Category: "deps", Decision: "Decision 3", Status: database.DecisionStatusActive},
	}
	p.SetDecisions(decisions)
	p.MoveDown()
	p.MoveDown() // cursor at 2

	// Set fewer decisions - cursor should be clamped
	fewerDecisions := []*database.Decision{
		{ID: "1", Category: "arch", Decision: "Decision 1", Status: database.DecisionStatusActive},
	}
	p.SetDecisions(fewerDecisions)

	if p.Cursor() != 0 {
		t.Errorf("cursor should be clamped to 0, got %d", p.Cursor())
	}
}

func TestRenderDecisionPreview(t *testing.T) {
	// Test nil decision
	result := RenderDecisionPreview(nil, 60, 40, "")
	if !containsString(result, "Select a decision") {
		t.Error("nil decision should show 'Select a decision' message")
	}

	// Test with decision
	decision := &database.Decision{
		ID:        "test-123",
		Category:  "architecture",
		Decision:  "Use SQLite for local storage",
		Rationale: "Simple, embedded, no server required",
		Status:    database.DecisionStatusActive,
		CreatedAt: time.Now(),
	}

	result = RenderDecisionPreview(decision, 60, 40, "")

	expectedStrings := []string{
		"DECISION DETAILS",
		"ACTIVE",
		"architecture",
		"SQLite",
	}

	for _, expected := range expectedStrings {
		if !containsString(result, expected) {
			t.Errorf("preview should contain %q", expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input    time.Time
		contains string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "minutes ago"},
		{now.Add(-2 * time.Hour), "hours ago"},
		{now.Add(-24 * time.Hour), "yesterday"},
		{now.Add(-3 * 24 * time.Hour), "days ago"},
		{now.Add(-30 * 24 * time.Hour), "2"}, // Should contain month/day
	}

	for _, tt := range tests {
		result := formatTime(tt.input)
		if !containsString(result, tt.contains) {
			t.Errorf("formatTime(%v) = %q, should contain %q", tt.input, result, tt.contains)
		}
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected int // expected number of lines
	}{
		{"short", 40, 1},
		{"This is a longer text that should wrap to multiple lines when width is small", 20, 4},
		{"word1 word2\nword3 word4", 40, 2}, // Preserves newlines
	}

	for _, tt := range tests {
		result := wrapText(tt.input, tt.width)
		lines := 1
		for _, c := range result {
			if c == '\n' {
				lines++
			}
		}
		if lines < tt.expected {
			t.Errorf("wrapText(%q, %d) produced %d lines, expected at least %d", tt.input, tt.width, lines, tt.expected)
		}
	}
}

func TestViewMode(t *testing.T) {
	// Test constants
	if ViewModeSessions != 0 {
		t.Error("ViewModeSessions should be 0 (default)")
	}

	if ViewModeDecisions != 1 {
		t.Error("ViewModeDecisions should be 1")
	}
}
