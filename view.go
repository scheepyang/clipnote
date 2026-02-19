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

	noteMarkSymbol = lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
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
		hint := statusStyle.Render(fmt.Sprintf("  Note L%d  (Ctrl+S submit | Esc cancel)", m.cursorLine+1))
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

	result := lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)

	if m.overlayType != overlayNone {
		content := m.overlayContent()
		result = renderOverlay(result, content, m.width, m.height)
	}
	return result
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
		if mk := m.GetMark(i); mk != nil {
			if mk.Note != "" {
				mark = noteMarkSymbol + " "
			} else {
				mark = markSymbol + " "
			}
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
	leftText := "  ? help | q quit | m mark | S export"
	right := statusStyle.Render(fmt.Sprintf("L%d/%d  Marks: %d  ", m.cursorLine+1, len(m.lines), len(m.marks)))

	rightW := lipgloss.Width(right)
	maxLeft := m.width - rightW
	if maxLeft < 0 {
		maxLeft = 0
	}
	left := statusStyle.Render(truncateLine(leftText, maxLeft))

	gap := m.width - lipgloss.Width(left) - rightW
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) overlayContent() string {
	switch m.overlayType {
	case overlayHelp:
		return `ai-review shortcuts

r         capture visible area
R         custom range capture
Ctrl+r    clear all content
j/k ↑/↓   move cursor
g / G     top / bottom
m         toggle mark
c         mark + note
v         view note
S         export to clipboard
P         paste to left pane
[ / ]     resize panels
q         quit
?         this help

press any key to close...`

	case overlayNote:
		for _, mk := range m.marks {
			if mk.Line == m.cursorLine && mk.Note != "" {
				return fmt.Sprintf("Note on L%d\n\n%s\n\npress any key to close...", mk.Line+1, mk.Note)
			}
		}
	}
	return ""
}

func renderOverlay(base, content string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	for len(baseLines) < height {
		baseLines = append(baseLines, strings.Repeat(" ", width))
	}

	contentLines := strings.Split(content, "\n")

	// calculate overlay dimensions (including 1-cell border padding)
	maxContentW := 0
	for _, cl := range contentLines {
		w := runewidth.StringWidth(cl)
		if w > maxContentW {
			maxContentW = w
		}
	}
	boxInnerW := maxContentW + 2 // 1-cell padding on each side
	if boxInnerW > width-4 {
		boxInnerW = width - 4
	}
	boxH := len(contentLines) + 2 // 1-cell padding on top and bottom
	if boxH > height-2 {
		boxH = height - 2
	}

	overlay := lipgloss.NewStyle().
		Width(boxInnerW).
		Height(boxH).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Render(content)

	overlayLines := strings.Split(overlay, "\n")
	overlayW := 0
	for _, ol := range overlayLines {
		w := runewidth.StringWidth(ol)
		if w > overlayW {
			overlayW = w
		}
	}

	// center the overlay
	startRow := (height - len(overlayLines)) / 2
	startCol := (width - overlayW) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	for i, ol := range overlayLines {
		row := startRow + i
		if row >= len(baseLines) {
			break
		}
		baseLine := baseLines[row]
		baseRunes := []rune(baseLine)

		// pad overlay line to uniform width to prevent base text bleeding through
		olW := runewidth.StringWidth(ol)
		if olW < overlayW {
			ol += strings.Repeat(" ", overlayW-olW)
		}

		// splice overlay line into base at startCol
		prefix := padToWidth(baseRunes, startCol)
		suffix := sliceFromWidth(baseRunes, startCol+overlayW)
		baseLines[row] = prefix + ol + suffix
	}

	return strings.Join(baseLines, "\n")
}

// padToWidth returns the first targetW display-width of runes, space-padded if shorter
func padToWidth(runes []rune, targetW int) string {
	w := 0
	for i, r := range runes {
		rw := runewidth.RuneWidth(r)
		if w+rw > targetW {
			return string(runes[:i]) + strings.Repeat(" ", targetW-w)
		}
		w += rw
	}
	return string(runes) + strings.Repeat(" ", targetW-w)
}

// sliceFromWidth returns the substring starting at display-width offset startW
func sliceFromWidth(runes []rune, startW int) string {
	w := 0
	for i, r := range runes {
		if w >= startW {
			return string(runes[i:])
		}
		w += runewidth.RuneWidth(r)
	}
	return ""
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
