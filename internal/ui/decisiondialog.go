package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DecisionDialog represents the decision logging dialog
type DecisionDialog struct {
	categoryInput  textinput.Model
	decisionInput  textarea.Model
	rationaleInput textarea.Model
	focusIndex     int
	width          int
	height         int
	visible        bool
	errorMsg       string
}

// NewDecisionDialog creates a new DecisionDialog instance
func NewDecisionDialog() *DecisionDialog {
	// Create category input
	categoryInput := textinput.New()
	categoryInput.Placeholder = "architecture, testing, dependencies..."
	categoryInput.Focus()
	categoryInput.CharLimit = 50
	categoryInput.Width = 50

	// Create decision input (text area for longer content)
	decisionInput := textarea.New()
	decisionInput.Placeholder = "What decision was made?"
	decisionInput.CharLimit = 500
	decisionInput.SetWidth(50)
	decisionInput.SetHeight(3)
	decisionInput.ShowLineNumbers = false

	// Create rationale input (text area for longer content)
	rationaleInput := textarea.New()
	rationaleInput.Placeholder = "Why was this decision made?"
	rationaleInput.CharLimit = 1000
	rationaleInput.SetWidth(50)
	rationaleInput.SetHeight(4)
	rationaleInput.ShowLineNumbers = false

	return &DecisionDialog{
		categoryInput:  categoryInput,
		decisionInput:  decisionInput,
		rationaleInput: rationaleInput,
		focusIndex:     0,
		visible:        false,
	}
}

// SetSize sets the dialog dimensions
func (d *DecisionDialog) SetSize(width, height int) {
	d.width = width
	d.height = height

	// Adjust input widths based on dialog size
	inputWidth := 50
	if width > 0 && width < 70 {
		inputWidth = width - 20
		if inputWidth < 30 {
			inputWidth = 30
		}
	}

	d.categoryInput.Width = inputWidth
	d.decisionInput.SetWidth(inputWidth)
	d.rationaleInput.SetWidth(inputWidth)
}

// Show makes the dialog visible
func (d *DecisionDialog) Show() {
	d.visible = true
	d.focusIndex = 0
	d.errorMsg = ""

	// Clear inputs
	d.categoryInput.SetValue("")
	d.decisionInput.SetValue("")
	d.rationaleInput.SetValue("")

	// Focus first input
	d.updateFocus()
}

// Hide hides the dialog
func (d *DecisionDialog) Hide() {
	d.visible = false
	d.errorMsg = ""
}

// IsVisible returns whether the dialog is visible
func (d *DecisionDialog) IsVisible() bool {
	return d.visible
}

// GetValues returns the current dialog values
func (d *DecisionDialog) GetValues() (category, decision, rationale string) {
	category = strings.TrimSpace(d.categoryInput.Value())
	decision = strings.TrimSpace(d.decisionInput.Value())
	rationale = strings.TrimSpace(d.rationaleInput.Value())
	return category, decision, rationale
}

// Validate checks if the dialog values are valid and returns an error message if not
func (d *DecisionDialog) Validate() string {
	category := strings.TrimSpace(d.categoryInput.Value())
	decision := strings.TrimSpace(d.decisionInput.Value())

	if category == "" {
		return "Category cannot be empty"
	}

	if decision == "" {
		return "Decision cannot be empty"
	}

	return "" // Valid
}

// SetError sets an error message to display
func (d *DecisionDialog) SetError(msg string) {
	d.errorMsg = msg
}

// updateFocus updates which input has focus
func (d *DecisionDialog) updateFocus() {
	d.categoryInput.Blur()
	d.decisionInput.Blur()
	d.rationaleInput.Blur()

	switch d.focusIndex {
	case 0:
		d.categoryInput.Focus()
	case 1:
		d.decisionInput.Focus()
	case 2:
		d.rationaleInput.Focus()
	}
}

// Update handles key messages
func (d *DecisionDialog) Update(msg tea.Msg) (*DecisionDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "ctrl+n":
			// Move to next field
			d.focusIndex = (d.focusIndex + 1) % 3
			d.updateFocus()
			return d, nil

		case "shift+tab", "ctrl+p":
			// Move to previous field
			d.focusIndex--
			if d.focusIndex < 0 {
				d.focusIndex = 2
			}
			d.updateFocus()
			return d, nil

		case "esc":
			d.Hide()
			return d, nil

		case "ctrl+s":
			// Submit shortcut - let parent handle via IsVisible check
			return d, nil
		}
	}

	// Update focused input
	switch d.focusIndex {
	case 0:
		d.categoryInput, cmd = d.categoryInput.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		d.decisionInput, cmd = d.decisionInput.Update(msg)
		cmds = append(cmds, cmd)
	case 2:
		d.rationaleInput, cmd = d.rationaleInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return d, tea.Batch(cmds...)
}

// View renders the dialog
func (d *DecisionDialog) View() string {
	if !d.visible {
		return ""
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPurple).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(ColorText)

	activeLabelStyle := lipgloss.NewStyle().
		Foreground(ColorPurple).
		Bold(true)

	// Responsive dialog width
	dialogWidth := 65
	if d.width > 0 && d.width < dialogWidth+10 {
		dialogWidth = d.width - 10
		if dialogWidth < 45 {
			dialogWidth = 45
		}
	}

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPurple).
		Background(ColorSurface).
		Padding(2, 4).
		Width(dialogWidth)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("📝 Log Decision"))
	content.WriteString("\n\n")

	// Category input
	if d.focusIndex == 0 {
		content.WriteString(activeLabelStyle.Render("▶ Category:"))
	} else {
		content.WriteString(labelStyle.Render("  Category:"))
	}
	content.WriteString("\n")
	content.WriteString("  ")
	content.WriteString(d.categoryInput.View())
	content.WriteString("\n\n")

	// Decision input
	if d.focusIndex == 1 {
		content.WriteString(activeLabelStyle.Render("▶ Decision:"))
	} else {
		content.WriteString(labelStyle.Render("  Decision:"))
	}
	content.WriteString("\n")
	content.WriteString("  ")
	content.WriteString(d.decisionInput.View())
	content.WriteString("\n\n")

	// Rationale input
	if d.focusIndex == 2 {
		content.WriteString(activeLabelStyle.Render("▶ Rationale:"))
	} else {
		content.WriteString(labelStyle.Render("  Rationale:"))
	}
	content.WriteString(" ")
	content.WriteString(lipgloss.NewStyle().Foreground(ColorComment).Render("(optional)"))
	content.WriteString("\n")
	content.WriteString("  ")
	content.WriteString(d.rationaleInput.View())
	content.WriteString("\n\n")

	// Error message if any
	if d.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)
		content.WriteString(errorStyle.Render("⚠ " + d.errorMsg))
		content.WriteString("\n\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorComment).
		MarginTop(1)
	content.WriteString(helpStyle.Render("Tab next │ Shift+Tab prev │ Ctrl+S save │ Esc cancel"))

	// Wrap in dialog box
	dialog := dialogStyle.Render(content.String())

	// Center the dialog
	return lipgloss.Place(
		d.width,
		d.height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)
}
