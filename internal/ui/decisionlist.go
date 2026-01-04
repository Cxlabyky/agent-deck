package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/asheshgoplani/agent-deck/internal/database"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ViewModeSessions  ViewMode = iota // Default: show sessions
	ViewModeDecisions                 // Show decisions list
)

// DecisionListPanel displays a list of decisions for a project
type DecisionListPanel struct {
	decisions    []*database.Decision
	cursor       int
	viewOffset   int
	width        int
	height       int
	projectPath  string
	lastRefresh  time.Time
}

// NewDecisionListPanel creates a new decision list panel
func NewDecisionListPanel() *DecisionListPanel {
	return &DecisionListPanel{
		decisions: []*database.Decision{},
	}
}

// SetSize sets the panel dimensions
func (p *DecisionListPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetDecisions updates the decisions list
func (p *DecisionListPanel) SetDecisions(decisions []*database.Decision) {
	p.decisions = decisions
	p.lastRefresh = time.Now()
	// Reset cursor if out of bounds
	if p.cursor >= len(decisions) {
		p.cursor = len(decisions) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// SetProjectPath sets the current project path
func (p *DecisionListPanel) SetProjectPath(path string) {
	p.projectPath = path
}

// GetProjectPath returns the current project path
func (p *DecisionListPanel) GetProjectPath() string {
	return p.projectPath
}

// Decisions returns the current decisions list
func (p *DecisionListPanel) Decisions() []*database.Decision {
	return p.decisions
}

// Cursor returns the current cursor position
func (p *DecisionListPanel) Cursor() int {
	return p.cursor
}

// Selected returns the currently selected decision
func (p *DecisionListPanel) Selected() *database.Decision {
	if p.cursor >= 0 && p.cursor < len(p.decisions) {
		return p.decisions[p.cursor]
	}
	return nil
}

// MoveUp moves the cursor up
func (p *DecisionListPanel) MoveUp() {
	if p.cursor > 0 {
		p.cursor--
		p.syncViewport()
	}
}

// MoveDown moves the cursor down
func (p *DecisionListPanel) MoveDown() {
	if p.cursor < len(p.decisions)-1 {
		p.cursor++
		p.syncViewport()
	}
}

// syncViewport ensures the cursor is visible
func (p *DecisionListPanel) syncViewport() {
	visibleLines := p.height - 2 // Account for header
	if visibleLines < 1 {
		visibleLines = 1
	}

	if p.cursor < p.viewOffset {
		p.viewOffset = p.cursor
	} else if p.cursor >= p.viewOffset+visibleLines {
		p.viewOffset = p.cursor - visibleLines + 1
	}
}

// Render renders the decision list
func (p *DecisionListPanel) Render(width, height int) string {
	p.width = width
	p.height = height

	if len(p.decisions) == 0 {
		return p.renderEmpty(width, height)
	}

	var b strings.Builder

	// Calculate visible items
	visibleLines := height
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Ensure viewOffset is valid
	if p.viewOffset < 0 {
		p.viewOffset = 0
	}
	maxOffset := len(p.decisions) - visibleLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if p.viewOffset > maxOffset {
		p.viewOffset = maxOffset
	}

	// Render visible decisions
	endIdx := p.viewOffset + visibleLines
	if endIdx > len(p.decisions) {
		endIdx = len(p.decisions)
	}

	linesRendered := 0
	for i := p.viewOffset; i < endIdx; i++ {
		decision := p.decisions[i]
		line := p.renderDecisionLine(decision, i == p.cursor, width)
		b.WriteString(line)
		b.WriteString("\n")
		linesRendered++
	}

	// Pad remaining lines
	for linesRendered < height {
		b.WriteString(strings.Repeat(" ", width))
		b.WriteString("\n")
		linesRendered++
	}

	result := b.String()
	// Remove trailing newline
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result
}

// renderEmpty renders the empty state
func (p *DecisionListPanel) renderEmpty(width, height int) string {
	var b strings.Builder

	// Center the message
	emptyLines := height / 2
	for i := 0; i < emptyLines-1; i++ {
		b.WriteString(strings.Repeat(" ", width))
		b.WriteString("\n")
	}

	// Empty state message
	msg := "No decisions yet"
	hint := "Press Ctrl+D to log a decision"

	msgStyle := lipgloss.NewStyle().
		Foreground(ColorComment).
		Width(width).
		Align(lipgloss.Center)

	hintStyle := lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Width(width).
		Align(lipgloss.Center)

	b.WriteString(msgStyle.Render(msg))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render(hint))
	b.WriteString("\n")

	// Fill remaining lines
	linesUsed := emptyLines + 1
	for linesUsed < height {
		b.WriteString(strings.Repeat(" ", width))
		b.WriteString("\n")
		linesUsed++
	}

	result := b.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result
}

// renderDecisionLine renders a single decision line
func (p *DecisionListPanel) renderDecisionLine(d *database.Decision, selected bool, width int) string {
	// Status indicator
	var statusIcon string
	var statusColor lipgloss.Color
	switch d.Status {
	case database.DecisionStatusActive:
		statusIcon = "●"
		statusColor = ColorGreen
	case database.DecisionStatusArchived:
		statusIcon = "○"
		statusColor = ColorComment
	case database.DecisionStatusOverridden:
		statusIcon = "✕"
		statusColor = ColorRed
	default:
		statusIcon = "●"
		statusColor = ColorGreen
	}

	// Category tag (compact)
	categoryWidth := 12
	category := d.Category
	if len(category) > categoryWidth-2 {
		category = category[:categoryWidth-3] + "…"
	}

	// Decision text (truncated)
	// Reserve space: 2 (padding) + 2 (status) + categoryWidth + 2 (spacing)
	reservedWidth := 4 + categoryWidth + 2
	decisionWidth := width - reservedWidth
	if decisionWidth < 10 {
		decisionWidth = 10
	}

	decisionText := d.Decision
	// Replace newlines with spaces for display
	decisionText = strings.ReplaceAll(decisionText, "\n", " ")
	if len(decisionText) > decisionWidth {
		decisionText = decisionText[:decisionWidth-1] + "…"
	}

	// Build the line
	var line strings.Builder

	if selected {
		// Selected style
		bgStyle := lipgloss.NewStyle().
			Background(ColorAccent).
			Foreground(ColorBg).
			Bold(true)

		statusStyle := lipgloss.NewStyle().
			Background(ColorAccent).
			Foreground(ColorBg)

		categoryStyle := lipgloss.NewStyle().
			Background(ColorAccent).
			Foreground(ColorBg).
			Width(categoryWidth)

		line.WriteString(bgStyle.Render("▶ "))
		line.WriteString(statusStyle.Render(statusIcon + " "))
		line.WriteString(categoryStyle.Render(category))
		line.WriteString(bgStyle.Render(" "))

		// Fill remaining width with selection color
		remainingWidth := width - 4 - categoryWidth - 1
		text := decisionText
		if len(text) < remainingWidth {
			text = text + strings.Repeat(" ", remainingWidth-len(text))
		}
		line.WriteString(bgStyle.Render(text))
	} else {
		// Normal style
		statusStyle := lipgloss.NewStyle().
			Foreground(statusColor)

		categoryStyle := lipgloss.NewStyle().
			Foreground(ColorPurple).
			Width(categoryWidth)

		decisionStyle := lipgloss.NewStyle().
			Foreground(ColorText)

		line.WriteString("  ")
		line.WriteString(statusStyle.Render(statusIcon + " "))
		line.WriteString(categoryStyle.Render(category))
		line.WriteString(" ")
		line.WriteString(decisionStyle.Render(decisionText))
	}

	return line.String()
}

// RenderDecisionPreview renders the preview for a selected decision
// sessionName is optional - pass empty string if no session is linked
func RenderDecisionPreview(d *database.Decision, width, height int, sessionName string) string {
	if d == nil {
		return renderNoDecisionSelected(width, height)
	}

	var b strings.Builder

	// Styles
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorCyan).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(ColorPurple).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(ColorText)

	dimStyle := lipgloss.NewStyle().
		Foreground(ColorComment)

	sessionStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Italic(true)

	// Status badge
	var statusBadge string
	switch d.Status {
	case database.DecisionStatusActive:
		statusBadge = lipgloss.NewStyle().
			Background(ColorGreen).
			Foreground(ColorBg).
			Padding(0, 1).
			Render("ACTIVE")
	case database.DecisionStatusArchived:
		statusBadge = lipgloss.NewStyle().
			Background(ColorYellow).
			Foreground(ColorBg).
			Padding(0, 1).
			Render("ARCHIVED")
	case database.DecisionStatusOverridden:
		statusBadge = lipgloss.NewStyle().
			Background(ColorRed).
			Foreground(ColorBg).
			Padding(0, 1).
			Render("OVERRIDDEN")
	}

	// Header
	b.WriteString(headerStyle.Render("📋 DECISION DETAILS"))
	b.WriteString("\n\n")

	// Status and Date
	b.WriteString(statusBadge)
	b.WriteString("  ")
	b.WriteString(dimStyle.Render(formatTime(d.CreatedAt)))
	b.WriteString("\n\n")

	// Category
	b.WriteString(labelStyle.Render("Category: "))
	b.WriteString(valueStyle.Render(d.Category))
	b.WriteString("\n\n")

	// Session link (if present)
	if sessionName != "" {
		b.WriteString(labelStyle.Render("Session: "))
		b.WriteString(sessionStyle.Render(sessionName))
		b.WriteString("\n\n")
	} else if d.SessionID != "" {
		// Show session ID if we couldn't resolve the name
		b.WriteString(labelStyle.Render("Session: "))
		b.WriteString(dimStyle.Render(d.SessionID[:8] + "..."))
		b.WriteString("\n\n")
	}

	// Decision
	b.WriteString(labelStyle.Render("Decision:"))
	b.WriteString("\n")
	// Word wrap the decision text
	wrapped := wrapText(d.Decision, width-4)
	for _, line := range strings.Split(wrapped, "\n") {
		b.WriteString("  ")
		b.WriteString(valueStyle.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Rationale (if present)
	if d.Rationale != "" {
		b.WriteString(labelStyle.Render("Rationale:"))
		b.WriteString("\n")
		wrapped := wrapText(d.Rationale, width-4)
		for _, line := range strings.Split(wrapped, "\n") {
			b.WriteString("  ")
			b.WriteString(valueStyle.Render(line))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Metadata
	b.WriteString(dimStyle.Render("─────────────────────────"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("ID: %s", d.ID)))
	b.WriteString("\n")

	return b.String()
}

// renderNoDecisionSelected renders the empty preview state
func renderNoDecisionSelected(width, height int) string {
	msg := "Select a decision to view details"

	style := lipgloss.NewStyle().
		Foreground(ColorComment).
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(msg)
}

// formatTime formats a time for display
func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	return t.Format("Jan 2, 2006")
}

// wrapText wraps text to fit within width
func wrapText(text string, width int) string {
	if width <= 0 {
		width = 40
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			}
		}
		result.WriteString(currentLine)
	}

	return result.String()
}
