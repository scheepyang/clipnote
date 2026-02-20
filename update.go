package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.ready = true
		}
		return m, nil

	case IPCMsg:
		resp := m.handleIPC(msg.Request)
		msg.ReplyCh <- resp
		return m, nil

	case CaptureAppendMsg:
		return m.handleCaptureAppend(string(msg)), nil

	case tea.KeyMsg:
		if m.overlayType != overlayNone {
			m.overlayType = overlayNone
			return m, nil
		}

		if m.captureConfirm {
			return m.handleCaptureConfirm(msg)
		}

		if m.captureInput {
			return m.handleCaptureInput(msg)
		}

		if m.inputMode {
			return m.handleInputMode(msg)
		}

		return m.handleBrowseMode(msg)
	}

	return m, nil
}

// handleCaptureAppend appends captured content to existing lines
func (m Model) handleCaptureAppend(content string) Model {
	m.captureCount++
	if len(m.lines) > 0 {
		separator := fmt.Sprintf("─── Capture #%d ───", m.captureCount)
		m.lines = append(m.lines, separator)
	}
	newLines := strings.Split(content, "\n")
	jumpTo := len(m.lines) // start position of newly captured content
	m.lines = append(m.lines, newLines...)
	m.cursorLine = jumpTo
	m.syncViewport()
	m.statusMsg = "Captured " + itoa(len(newLines)) + " lines (total " + itoa(len(m.lines)) + ")"
	return m
}

func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.SubmitNote):
		note := m.noteInput.Value()
		if note != "" {
			m.AddMarkWithNote(m.cursorLine, note)
		}
		m.noteInput.Reset()
		m.noteInput.Blur()
		m.inputMode = false
		m.statusMsg = ""
		return m, nil

	case key.Matches(msg, keys.Escape):
		m.noteInput.Reset()
		m.noteInput.Blur()
		m.inputMode = false
		m.statusMsg = ""
		return m, nil

	default:
		var cmd tea.Cmd
		m.noteInput, cmd = m.noteInput.Update(msg)
		return m, cmd
	}
}

// handleCaptureInput handles R key line count input mode
func (m Model) handleCaptureInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.captureInput = false
		m.captureInputBuf = ""
		m.statusMsg = ""
		return m, nil

	case tea.KeyEnter:
		m.captureInput = false
		input := strings.TrimSpace(m.captureInputBuf)
		m.captureInputBuf = ""

		if input == "" {
			// empty = full scrollback, requires confirmation
			m.captureConfirm = true
			m.statusMsg = "Scrollback may contain lots of content. Confirm capture? (y/n)"
			return m, nil
		}

		n, err := strconv.Atoi(input)
		if err != nil || n <= 0 {
			m.statusMsg = "Please enter a positive integer"
			return m, nil
		}
		return m, captureRange(m.tmuxPane, n)

	case tea.KeyBackspace:
		if len(m.captureInputBuf) > 0 {
			m.captureInputBuf = m.captureInputBuf[:len(m.captureInputBuf)-1]
		}
		return m, nil

	default:
		// accept digits only
		for _, r := range msg.String() {
			if unicode.IsDigit(r) {
				m.captureInputBuf += string(r)
			}
		}
		return m, nil
	}
}

// handleCaptureConfirm handles full scrollback capture confirmation
func (m Model) handleCaptureConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.captureConfirm = false
	switch msg.String() {
	case "y", "Y":
		m.statusMsg = "Capturing full scrollback..."
		return m, captureAll(m.tmuxPane)
	default:
		m.statusMsg = "Cancelled"
		return m, nil
	}
}

func (m Model) handleBrowseMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.ClearAll):
		m.lines = []string{}
		m.marks = []Mark{}
		m.captureCount = 0
		m.cursorLine = 0
		m.statusMsg = "Cleared all content and marks"

	case key.Matches(msg, keys.Capture):
		return m, captureVisible(m.tmuxPane)

	case key.Matches(msg, keys.CaptureRange):
		m.captureInput = true
		m.captureInputBuf = ""
		m.statusMsg = "Enter line count (empty = full scrollback)"
		return m, nil

	case key.Matches(msg, keys.Down):
		if m.cursorLine < len(m.lines)-1 {
			m.cursorLine++
			m.syncViewport()
		}
		m.statusMsg = ""

	case key.Matches(msg, keys.Up):
		if m.cursorLine > 0 {
			m.cursorLine--
			m.syncViewport()
		}
		m.statusMsg = ""

	case key.Matches(msg, keys.Top):
		m.cursorLine = 0
		m.scrollOffset = 0
		m.statusMsg = ""

	case key.Matches(msg, keys.Bottom):
		m.cursorLine = len(m.lines) - 1
		m.syncViewport()
		m.statusMsg = ""

	case key.Matches(msg, keys.Mark):
		if len(m.lines) == 0 {
			break
		}
		m.ToggleMark(m.cursorLine)
		if m.HasMark(m.cursorLine) {
			m.statusMsg = "Marked L" + itoa(m.cursorLine+1)
		} else {
			m.statusMsg = "Unmarked L" + itoa(m.cursorLine+1)
		}

	case key.Matches(msg, keys.Comment):
		if len(m.lines) == 0 {
			break
		}
		if !m.HasMark(m.cursorLine) {
			m.ToggleMark(m.cursorLine)
		}
		m.inputMode = true
		m.noteInput.Focus()
		return m, m.noteInput.Focus()

	case key.Matches(msg, keys.Submit):
		m.statusMsg = m.CopyMarksToClipboard()

	case key.Matches(msg, keys.PasteToPane):
		m.statusMsg = m.PasteMarksToPane()

	case key.Matches(msg, keys.ViewNote):
		for _, mk := range m.marks {
			if mk.Line == m.cursorLine && mk.Note != "" {
				m.overlayType = overlayNote
				return m, nil
			}
		}

	case key.Matches(msg, keys.Help):
		m.overlayType = overlayHelp

	case key.Matches(msg, keys.ShrinkLeft):
		if m.splitRatio > 30 {
			m.splitRatio -= 5
		}

	case key.Matches(msg, keys.ExpandLeft):
		if m.splitRatio < 90 {
			m.splitRatio += 5
		}
	}

	return m, nil
}

// syncViewport ensures the cursor stays within the visible range
func (m *Model) syncViewport() {
	contentHeight := m.height - 3
	if contentHeight <= 0 {
		return
	}

	// cursor above visible area
	if m.cursorLine < m.scrollOffset {
		m.scrollOffset = m.cursorLine
	}

	// cursor below visible area
	if m.cursorLine >= m.scrollOffset+contentHeight {
		m.scrollOffset = m.cursorLine - contentHeight + 1
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
