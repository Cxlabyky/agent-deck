package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDecisionDialog(t *testing.T) {
	d := NewDecisionDialog()

	if d == nil {
		t.Fatal("NewDecisionDialog should return non-nil")
	}

	if d.IsVisible() {
		t.Error("dialog should not be visible initially")
	}
}

func TestDecisionDialog_ShowHide(t *testing.T) {
	d := NewDecisionDialog()

	d.Show()
	if !d.IsVisible() {
		t.Error("dialog should be visible after Show()")
	}

	d.Hide()
	if d.IsVisible() {
		t.Error("dialog should not be visible after Hide()")
	}
}

func TestDecisionDialog_SetSize(t *testing.T) {
	d := NewDecisionDialog()
	d.SetSize(100, 50)

	// SetSize should work without panic
	if d.width != 100 || d.height != 50 {
		t.Errorf("SetSize should set dimensions, got width=%d, height=%d", d.width, d.height)
	}
}

func TestDecisionDialog_GetValues(t *testing.T) {
	d := NewDecisionDialog()
	d.Show()

	// Set values manually through the inputs
	d.categoryInput.SetValue("architecture")
	d.decisionInput.SetValue("Use SQLite for storage")
	d.rationaleInput.SetValue("Simple, embedded, no server needed")

	category, decision, rationale := d.GetValues()

	if category != "architecture" {
		t.Errorf("expected category 'architecture', got %q", category)
	}
	if decision != "Use SQLite for storage" {
		t.Errorf("expected decision 'Use SQLite for storage', got %q", decision)
	}
	if rationale != "Simple, embedded, no server needed" {
		t.Errorf("expected rationale 'Simple, embedded, no server needed', got %q", rationale)
	}
}

func TestDecisionDialog_Validate(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		decision  string
		wantError bool
	}{
		{
			name:      "valid input",
			category:  "architecture",
			decision:  "Use SQLite",
			wantError: false,
		},
		{
			name:      "empty category",
			category:  "",
			decision:  "Use SQLite",
			wantError: true,
		},
		{
			name:      "empty decision",
			category:  "architecture",
			decision:  "",
			wantError: true,
		},
		{
			name:      "both empty",
			category:  "",
			decision:  "",
			wantError: true,
		},
		{
			name:      "whitespace only category",
			category:  "   ",
			decision:  "Use SQLite",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionDialog()
			d.Show()

			d.categoryInput.SetValue(tt.category)
			d.decisionInput.SetValue(tt.decision)

			err := d.Validate()
			hasError := err != ""

			if hasError != tt.wantError {
				t.Errorf("Validate() error = %q, wantError = %v", err, tt.wantError)
			}
		})
	}
}

func TestDecisionDialog_EscClosesDialog(t *testing.T) {
	d := NewDecisionDialog()
	d.Show()

	if !d.IsVisible() {
		t.Fatal("dialog should be visible after Show()")
	}

	// Simulate Esc key
	msg := tea.KeyMsg{Type: tea.KeyEscape}
	d.Update(msg)

	if d.IsVisible() {
		t.Error("dialog should be hidden after Esc key")
	}
}

func TestDecisionDialog_TabNavigatesFields(t *testing.T) {
	d := NewDecisionDialog()
	d.Show()

	// Initially should be on first field
	if d.focusIndex != 0 {
		t.Errorf("initial focus should be 0, got %d", d.focusIndex)
	}

	// Tab to next field
	msg := tea.KeyMsg{Type: tea.KeyTab}
	d.Update(msg)

	if d.focusIndex != 1 {
		t.Errorf("focus should be 1 after Tab, got %d", d.focusIndex)
	}

	// Tab again
	d.Update(msg)
	if d.focusIndex != 2 {
		t.Errorf("focus should be 2 after second Tab, got %d", d.focusIndex)
	}

	// Tab wraps around
	d.Update(msg)
	if d.focusIndex != 0 {
		t.Errorf("focus should wrap to 0 after third Tab, got %d", d.focusIndex)
	}
}

func TestDecisionDialog_ShiftTabNavigatesBackward(t *testing.T) {
	d := NewDecisionDialog()
	d.Show()

	// Shift+Tab from first field should go to last
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	d.Update(msg)

	if d.focusIndex != 2 {
		t.Errorf("focus should be 2 after Shift+Tab from 0, got %d", d.focusIndex)
	}
}

func TestDecisionDialog_View(t *testing.T) {
	d := NewDecisionDialog()
	d.SetSize(80, 40)

	// View should return empty when not visible
	if d.View() != "" {
		t.Error("View() should return empty string when dialog is not visible")
	}

	d.Show()
	view := d.View()

	if view == "" {
		t.Error("View() should return non-empty string when dialog is visible")
	}

	// Check for expected content
	expectedStrings := []string{
		"Log Decision",
		"Category",
		"Decision",
		"Rationale",
		"Ctrl+S save",
		"Esc cancel",
	}

	for _, expected := range expectedStrings {
		if !containsString(view, expected) {
			t.Errorf("View() should contain %q", expected)
		}
	}
}

func TestDecisionDialog_SetError(t *testing.T) {
	d := NewDecisionDialog()
	d.SetSize(80, 40)
	d.Show()

	d.SetError("Test error message")

	view := d.View()
	if !containsString(view, "Test error message") {
		t.Error("View() should contain error message when set")
	}
}

func TestDecisionDialog_ShowClearsInputs(t *testing.T) {
	d := NewDecisionDialog()
	d.Show()

	// Set some values
	d.categoryInput.SetValue("test")
	d.decisionInput.SetValue("test decision")
	d.rationaleInput.SetValue("test rationale")

	// Show again should clear
	d.Show()

	category, decision, rationale := d.GetValues()
	if category != "" || decision != "" || rationale != "" {
		t.Error("Show() should clear all input values")
	}
}

// containsString is a helper to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
