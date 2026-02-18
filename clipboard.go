package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
)

func (m *Model) CopyMarksToClipboard() string {
	text := m.ExportMarks()
	if text == "" {
		return "No marks to export"
	}
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Sprintf("Clipboard write failed: %v", err)
	}
	return fmt.Sprintf("Copied %d marks to clipboard", len(m.marks))
}

func (m *Model) PasteMarksToPane() string {
	text := m.ExportMarks()
	if text == "" {
		return "No marks to export"
	}
	if err := exec.Command("tmux", "set-buffer", text).Run(); err != nil {
		return fmt.Sprintf("Failed to set tmux buffer: %v", err)
	}
	out, err := exec.Command("tmux", "paste-buffer", "-t", m.tmuxPane).CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Sprintf("Failed to paste to pane: %s", errMsg)
	}
	return fmt.Sprintf("Pasted %d marks to left pane", len(m.marks))
}
