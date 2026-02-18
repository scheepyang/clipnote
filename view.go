package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62"))

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	markSymbol = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Render("●")

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	leftWidth := m.width*m.splitRatio/100 - 2
	rightWidth := m.width - leftWidth - 4
	statusHeight := 1
	if m.inputMode {
		statusHeight = noteInputHeight + 2
	} else if m.captureInput {
		statusHeight = 2
	}
	contentHeight := m.height - statusHeight - 2

	// left: content panel
	leftBorderStyle := borderStyle.
		BorderLeft(false).
		BorderRight(false)
	leftContent := m.renderContent(leftWidth, contentHeight)
	leftPanel := leftBorderStyle.
		Width(leftWidth).
		Height(contentHeight).
		Render(leftContent)

	// right: marks panel
	rightContent := m.renderMarks(rightWidth, contentHeight)
	rightPanel := borderStyle.
		Width(rightWidth).
		Height(contentHeight).
		Render(rightContent)

	// join left and right panels
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// bottom status bar
	var statusBar string
	if m.inputMode {
		hint := statusStyle.Render(fmt.Sprintf("  Note L%d  (Enter newline | Ctrl+S submit | Esc cancel)", m.cursorLine+1))
		statusBar = hint + "\n" + m.noteInput.View()
	} else if m.captureInput {
		input := m.captureInputBuf
		if input == "" {
			input = "(empty = full scrollback)"
		}
		statusBar = statusStyle.Render(fmt.Sprintf("  Lines: %s  (Enter confirm | Esc cancel)", input))
	} else if m.statusMsg != "" {
		statusBar = statusStyle.Render(fmt.Sprintf("  %s", m.statusMsg))
	} else {
		statusBar = m.renderStatusBar()
	}

	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)
}

func (m Model) renderContent(width, height int) string {
	if len(m.lines) == 0 {
		return helpStyle.Render("Press r to capture left pane content")
	}

	// ensure scrollOffset keeps cursor within visible range
	start := m.scrollOffset
	if start > m.cursorLine {
		start = m.cursorLine
	}
	if m.cursorLine >= start+height {
		start = m.cursorLine - height + 1
	}
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > len(m.lines) {
		end = len(m.lines)
	}

	var lines []string
	for i := start; i < end; i++ {
		lineNum := fmt.Sprintf("%4d", i+1)
		mark := "  "
		if m.HasMark(i) {
			mark = markSymbol + " "
		}

		lineText := truncateLine(m.lines[i], width-8)

		if i == m.cursorLine {
			line := cursorStyle.Render(fmt.Sprintf("▶%s %s%s", lineNum, mark, lineText))
			lines = append(lines, line)
		} else {
			line := fmt.Sprintf(" %s %s%s", lineNum, mark, lineText)
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderMarks(width, height int) string {
	title := titleStyle.Render(fmt.Sprintf("Marks (%d)", len(m.marks)))
	lines := []string{title, ""}

	for _, mk := range m.marks {
		entry := fmt.Sprintf("L%d", mk.Line+1)
		if mk.Note != "" {
			entry += " " + truncateLine(mk.Note, width-8)
		} else {
			entry += " " + truncateLine(mk.Text, width-8)
		}
		lines = append(lines, entry)
	}

	if len(m.marks) == 0 {
		lines = append(lines, helpStyle.Render("Press m to mark a line"))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderStatusBar() string {
	left := statusStyle.Render("  [r]capture [m]mark [c]note [S]export [P]paste [?]help [q]quit")

	right := statusStyle.Render(fmt.Sprintf("L%d/%d  Marks: %d  ", m.cursorLine+1, len(m.lines), len(m.marks)))

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderHelp() string {
	help := `
  ai-review keybindings

  r           Capture visible area (append)
  R           Custom range capture (append)
  Ctrl+r      Clear all content and marks
  j / ↓       Move down
  k / ↑       Move up
  g           Jump to top
  G           Jump to bottom
  m           Toggle mark on current line
  c           Mark current line + open note input
  Enter       Newline (input mode)
  Ctrl+S      Submit note (input mode)
  Esc         Cancel input (input mode)
  S           Export all marks to clipboard
  P           Paste marks to left pane
  [ / ]       Shrink/expand content panel
  q           Quit
  ?           Toggle this help

  Press any key to close...
`
	return helpStyle.Render(help)
}

func truncateLine(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	// truncate, reserve 1 cell for "…"
	w := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth-1 {
			return s[:i] + "…"
		}
		w += rw
	}
	return s
}
