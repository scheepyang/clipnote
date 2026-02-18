package main

import (
	"fmt"
	"strings"
)

const noteExportPrefix = "[Q] "

// Mark represents a user's annotation on a specific line
type Mark struct {
	Line int
	Text string // original line text (truncated to 60 chars)
	Note string // user's annotation
}

func (m *Model) ToggleMark(line int) {
	if m.HasMark(line) {
		m.RemoveMark(line)
		return
	}
	text := truncate(m.lines[line], 60)
	m.marks = append(m.marks, Mark{Line: line, Text: text})
}

func (m *Model) AddMarkWithNote(line int, note string) {
	// update note if mark already exists
	for i, mk := range m.marks {
		if mk.Line == line {
			m.marks[i].Note = note
			return
		}
	}
	text := truncate(m.lines[line], 60)
	m.marks = append(m.marks, Mark{Line: line, Text: text, Note: note})
}

func (m *Model) RemoveMark(line int) {
	for i, mk := range m.marks {
		if mk.Line == line {
			m.marks = append(m.marks[:i], m.marks[i+1:]...)
			return
		}
	}
}

func (m *Model) HasMark(line int) bool {
	for _, mk := range m.marks {
		if mk.Line == line {
			return true
		}
	}
	return false
}

func (m *Model) ExportMarks() string {
	if len(m.marks) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, mk := range m.marks {
		sb.WriteString(mk.Text + "\n")
		if mk.Note != "" {
			sb.WriteString(fmt.Sprintf("> %s%s\n", noteExportPrefix, mk.Note))
		}
	}
	return strings.TrimSpace(sb.String())
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
